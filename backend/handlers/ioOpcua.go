package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gopcua/opcua"
	"github.com/gopcua/opcua/monitor"
	"github.com/gopcua/opcua/ua"
	"github.com/nalgeon/redka"
	"log"
	"sync"
	"time"
)

// OpcUARead 函数：去设备点表中获取配置信息，然后连接OPC Server订阅数据
func OpcUARead(id string, stopChan chan struct{}, cfgdb *redka.DB, rtdb *redka.DB) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("OpcUARead 发生 panic: %v\n", r)
		}
	}()

	// 通过ID(实例ID)获取实例的配置信息
	appconfig, err := cfgdb.Hash().Get(InstListKey, id)
	if err != nil {
		log.Printf("数据库中没有实例ID: %v\n", id)
		return
	}
	configstr := appconfig.String()
	var newConfig AppConfig
	err = json.Unmarshal([]byte(configstr), &newConfig)
	configMap := newConfig.Config
	log.Printf("实例配置: %+v\n", configMap)

	// 提取外层的 "Config"
	config, ok := configMap.(map[string]any)
	if !ok {
		log.Println("配置不是 map[string]any 或不存在")
		return
	}

	endpoint, ok := config["endpoint"].(string)
	if !ok {
		log.Println("endpoint 不是字符串或不存在")
	}
	policy, ok := config["policy"].(string)
	if !ok {
		log.Println("policy 不是字符串或不存在")
	}
	mode, ok := config["mode"].(string)
	if !ok {
		log.Println("mode 不是字符串或不存在")
	}
	certFile, ok := config["cert"].(string)
	if !ok {
		log.Println("certFile 不是字符串或不存在")
	}
	keyFile, ok := config["key"].(string)
	if !ok {
		log.Println("keyFile 不是字符串或不存在")
	}

	log.Printf("endpoint: %+v, policy: %+v, mode: %+v, certFile: %+v, keyFile: %+v\n", endpoint, policy, mode, certFile, keyFile)
	// 通过ID(实例ID)获取当前函数可读写的设备配置信息和设备点表信息
	devValues, err1 := cfgdb.Hash().Items(DevAtInstKey)
	if err1 != nil {
		log.Printf("获取设备配置信息失败: %v\n", err1)
		return
	}
	if len(devValues) == 0 {
		log.Printf("数据库中没有设备\n")
		return
	}
	//通过设备ID获取设备点信息
	devMap := make(map[string]DevConfig)
	for key, value := range devValues {
		var newValue DevConfig
		erra := json.Unmarshal([]byte(value.String()), &newValue)
		if erra != nil {
			log.Println("解析 JSON 失败:", erra)
			return
		}
		log.Printf("键: %s, Queryid: %s, InstID: %s\n", key, id, newValue.InstID)
		if id == newValue.InstID {
			devMap[key] = newValue
		}
	}
	if len(devMap) == 0 {
		log.Printf("实例ID %v 没有匹配的设备\n", id)
		return
	}

	// 通过已经获取设备点表信息生成设备采集点表
	opctags := make([]string, 0)
	opcBind := make(map[string]string, 0)
	opcParent := make(map[string]string, 0)
	for devkey := range devMap {
		tags, err2 := cfgdb.Hash().Items(devkey)
		if err2 != nil {
			log.Printf("获取设备点表失败: %v\n", err2)
			continue
		}
		if len(tags) != 0 {
			for tagkey, tagvalue := range tags {
				var newValue []string
				erra := json.Unmarshal([]byte(tagvalue.String()), &newValue)
				if erra != nil {
					log.Println("解析 JSON 失败:", erra)
					return
				}
				opcitem := newValue[0]
				opctags = append(opctags, opcitem)
				opcParent[opcitem] = devkey
				opcBind[opcitem] = tagkey
			}
		}
	}
	if len(opctags) == 0 {
		log.Printf("实例ID %v 没有标签\n", id)
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var c *opcua.Client
	var m *monitor.NodeMonitor
	var wg sync.WaitGroup

	// 连接 OPC UA Server
	connect := func() error {
		endpoints, err := opcua.GetEndpoints(ctx, endpoint)
		if err != nil {
			return fmt.Errorf("获取端点失败: %v", err)
		}
		ep, err := opcua.SelectEndpoint(endpoints, policy, ua.MessageSecurityModeFromString(mode))
		if err != nil {
			return fmt.Errorf("选择端点失败: %v", err)
		}
		log.Print("*", ep.SecurityPolicyURI, ep.SecurityMode)

		opts := []opcua.Option{
			opcua.SecurityPolicy(policy),
			opcua.SecurityModeString(mode),
			opcua.CertificateFile(certFile),
			opcua.PrivateKeyFile(keyFile),
			opcua.AuthAnonymous(),
			opcua.SecurityFromEndpoint(ep, ua.UserTokenTypeAnonymous),
		}

		c, _ = opcua.NewClient(ep.EndpointURL, opts...)
		if err := c.Connect(ctx); err != nil {
			return fmt.Errorf("连接失败: %v", err)
		}

		m, err = monitor.NewNodeMonitor(c)
		if err != nil {
			return fmt.Errorf("创建监控器失败: %v", err)
		}

		m.SetErrorHandler(func(_ *opcua.Client, sub *monitor.Subscription, err error) {
			log.Printf("错误: sub=%d err=%s", sub.SubscriptionID(), err.Error())
		})

		return nil
	}

	// 重连机制
	reconnect := func() {
		for {
			select {
			case <-stopChan:
				log.Printf("收到停止信号，退出重连循环\n")
				return
			default:
				log.Printf("尝试连接 OPC UA Server\n")
				if err := connect(); err == nil {
					log.Println("连接成功")
					return
				}
				log.Printf("连接失败，等待 %v 后重试\n", reconnectDelay)
				time.Sleep(reconnectDelay)
			}
		}
	}

	// 初始连接
	reconnect()
	defer func() {
		if c != nil {
			c.Close(ctx)
		}
	}()

	// 创建队列
	queue := NewDataQueue()

	// 启动子线程
	wg.Add(1)
	go startCallbackSub(ctx, m, 1, 0, &wg, queue, opctags...)

	// 监听停止信号
	for {
		select {
		case <-stopChan:
			log.Printf("子线程 OPCUA 实例 %s 收到停止信号，退出\n", id)
			cancel()  // 取消上下文，确保 startCallbackSub 退出
			wg.Wait() // 等待子线程退出
			return
		default:
			// 检查连接状态
			if c == nil || c.State() != opcua.Connected {
				log.Println("检测到连接断开，尝试重新连接")
				reconnect()
				continue
			}

			// 处理数据
			datasmap := make(map[string]map[string]any)
			for queue.Len() > 0 {
				log.Println("队列长度:", queue.Len())
				if val, ok := queue.Dequeue(); ok {
					//log.Println("消费数据:", val)
					var data []any
					err := json.Unmarshal([]byte(val), &data)
					if err != nil {
						//log.Println("解析失败:", err)
						return
					}
					opcitem := data[0].(string)
					devkey := opcParent[opcitem]
					if datasmap[devkey] == nil {
						datasmap[devkey] = make(map[string]any)
					}
					tagkey := opcBind[opcitem]
					valueMap := []any{data[1], data[2], data[3]}
					valueMapJson, _ := json.Marshal(valueMap)
					datasmap[devkey][tagkey] = valueMapJson
				}
			}
			// 统一将数据写入到redka数据库
			for devkey := range datasmap {
				_, errz := rtdb.Hash().SetMany(devkey, datasmap[devkey])
				if errz != nil {
					log.Printf("写入数据库失败: %v\n", errz)
					continue
				}
			}
			time.Sleep(1 * time.Second)
		}
	}
}

func startCallbackSub(ctx context.Context, m *monitor.NodeMonitor, interval, lag time.Duration, wg *sync.WaitGroup, queue *DataQueue, nodes ...string) {
	defer wg.Done() // 确保在函数退出时调用 Done()
	sub, err := m.Subscribe(
		ctx,
		&opcua.SubscriptionParameters{
			Interval: interval,
		},

		func(s *monitor.Subscription, msg *monitor.DataChangeMessage) {
			if msg.Error != nil {
				log.Printf("[callback] sub=%d error=%s", s.SubscriptionID(), msg.Error)
			} else {
				valueMap := []any{msg.NodeID, msg.ServerTimestamp.Local().Format("2006-01-02 15:04:05"), msg.Value.Value(), msg.ServerTimestamp.Unix()}
				valueMapJson, _ := json.Marshal(valueMap)
				queue.Enqueue(string(valueMapJson))
			}
			time.Sleep(lag)
		},
		nodes...)

	if err != nil {
		log.Fatal(err)
	}

	defer cleanup(ctx, sub)

	<-ctx.Done()
}

func cleanup(ctx context.Context, sub *monitor.Subscription) {
	log.Printf("统计: sub=%d 已传递=%d 已丢弃=%d", sub.SubscriptionID(), sub.Delivered(), sub.Dropped())
	sub.Unsubscribe(ctx)
}
