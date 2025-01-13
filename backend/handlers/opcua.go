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

// OpcUARead函数：去设备点表中获取配置信息，然后连接OPC Server订阅数据
func OpcUARead(id string, stopChan chan struct{}, cfgdb *redka.DB, rtdb *redka.DB) {
	// 通过ID(实例ID)获取实例的配置信息
	appconfig, err := cfgdb.Hash().Get(InstListKey, id)
	if err != nil {
		fmt.Printf("database no instid\n")
		return
	}
	configstr := appconfig.String()
	var newConfig AppConfig
	err = json.Unmarshal([]byte(configstr), &newConfig)
	configMap := newConfig.Config
	fmt.Printf("myConfig %+v\n", configMap)

	// 提取外层的 "Config"
	config, ok := configMap.(map[string]any)
	if !ok {
		fmt.Println("Config is not a map[string]any or does not exist")
		return
	}

	endpoint, ok := config["endpoint"].(string)
	if !ok {
		fmt.Println("host is not a string or does not exist")
	}
	policy, ok := config["policy"].(string)
	if !ok {
		fmt.Println("policy is not a string or does not exist")
	}
	mode, ok := config["mode"].(string)
	if !ok {
		fmt.Println("mode is not a string or does not exist")
	}
	certFile, ok := config["cert"].(string)
	if !ok {
		fmt.Println("certFile is not a string or does not exist")
	}
	keyFile, ok := config["key"].(string)
	if !ok {
		fmt.Println("keyFile is not a string or does not exist")
	}

	fmt.Printf("endpoint: %+v, policy: %+v, mode: %+v, certFile: %+v, keyFile: %+v\n", endpoint, policy, mode, certFile, keyFile)
	// 通过ID(实例ID)获取当前函数可读写的设备配置信息和设备点表信息
	devValues, err1 := cfgdb.Hash().Items(DevAtInstKey)
	if err1 != nil {
		fmt.Printf("Err: %v\n", err1)
		return
	}
	if len(devValues) == 0 {
		fmt.Printf("database no any device\n")
		return
	}
	//通过设备ID获取设备点信息
	devMap := make(map[string]DevConfig)
	for key, value := range devValues {
		var newValue DevConfig
		erra := json.Unmarshal([]byte(value.String()), &newValue)
		if erra != nil {
			fmt.Println("Error unmarshalling JSON:", erra)
			return
		}
		fmt.Printf("键: %s, Queryid: %s, InstID: %s\n", key, id, newValue.InstID)
		if id == newValue.InstID {
			devMap[key] = newValue
		}
	}
	if len(devMap) == 0 {
		fmt.Printf("instid %v no match device\n", id)
		return
	}

	// 通过已经获取设备点表信息生成设备采集点表
	opctags := make([]string, 0)
	opcBind := make(map[string]string, 0)
	opcParent := make(map[string]string, 0)
	for devkey := range devMap {
		tags, err2 := cfgdb.Hash().Items(devkey)
		if err2 != nil {
			fmt.Printf("Err: %v\n", err2)
			continue
		}
		if len(tags) != 0 {
			for tagkey, tagvalue := range tags {
				var newValue []string
				erra := json.Unmarshal([]byte(tagvalue.String()), &newValue)
				if erra != nil {
					fmt.Println("Error unmarshalling JSON:", erra)
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
		fmt.Printf("instid %v no tag\n", id)
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	endpoints, err := opcua.GetEndpoints(ctx, endpoint)
	if err != nil {
		log.Fatal(err)
	}
	ep, err := opcua.SelectEndpoint(endpoints, policy, ua.MessageSecurityModeFromString(mode))
	if err != nil {
		log.Fatal(err)
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

	c, err := opcua.NewClient(ep.EndpointURL, opts...)
	if err != nil {
		log.Fatal(err)
	}
	if err := c.Connect(ctx); err != nil {
		log.Fatal(err)
	}
	defer c.Close(ctx)

	m, err := monitor.NewNodeMonitor(c)
	if err != nil {
		log.Fatal(err)
	}

	m.SetErrorHandler(func(_ *opcua.Client, sub *monitor.Subscription, err error) {
		log.Printf("error: sub=%d err=%s", sub.SubscriptionID(), err.Error())
	})

	// 创建 WaitGroup 用于同步
	wg := &sync.WaitGroup{}
	// 启动子线程
	// start callback-based subscription
	wg.Add(1)
	go startCallbackSub(ctx, m, 1, 0, wg, opctags...)

	// 监听停止信号
	select {
	case <-stopChan:
		fmt.Printf("子线程OPCUA实例 %s 收到停止信号，退出\n", id)
		cancel()  // 取消上下文，确保 startCallbackSub 退出
		wg.Wait() // 等待子线程退出
		return
	}
}

func startCallbackSub(ctx context.Context, m *monitor.NodeMonitor, interval, lag time.Duration, wg *sync.WaitGroup, nodes ...string) {
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
				log.Printf("[callback] sub=%d ts=%s node=%s value=%v", s.SubscriptionID(), msg.ServerTimestamp.UTC().Format(time.RFC3339), msg.NodeID, msg.Value.Value())

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
	log.Printf("stats: sub=%d delivered=%d dropped=%d", sub.SubscriptionID(), sub.Delivered(), sub.Dropped())
	sub.Unsubscribe(ctx)
}
