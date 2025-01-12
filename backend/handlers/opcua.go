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
	//通过ID(实例ID)获取实例的配置信息
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
	config, ok := configMap.(map[string]any) // 类型断言为 map[string]any
	if !ok {
		fmt.Println("Config is not a map[string]any or does not exist")
		return
	}
	//host := "localhost"
	//progID := "Matrikon.OPC.Simulation.1"
	endpoint, ok := config["host"].(string)
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
	certFile, ok := config["certFile"].(string)
	if !ok {
		fmt.Println("certFile is not a string or does not exist")
	}
	keyFile, ok := config["keyFile"].(string)
	if !ok {
		fmt.Println("keyFile is not a string or does not exist")
	}
	interval, ok := config["interval"].(time.Duration)
	if !ok {
		fmt.Println("interval is not a string or does not exist")
	}
	fmt.Printf("endpoint: %+v, policy: %+v, mode: %+v, certFile: %+v, keyFile: %+v\n", endpoint, policy, mode, certFile, keyFile)
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
	// 通过设备ID获取设备点表信息
	// 用于存储所有订阅数据的OPC标签
	opctags := make([]string, 0)
	// 用于存储每个标签的父设备
	opcBind := make(map[string]string, 0)
	opcParent := make(map[string]string, 0)
	for devkey := range devMap {
		// 从设备点表中获取配置信息
		tags, err2 := cfgdb.Hash().Items(devkey)
		if err2 != nil {
			fmt.Printf("Err: %v\n", err2)
			continue
		}
		if len(tags) != 0 {
			// 遍历设备点表获取数据
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

	// 创建 WaitGroup 用于同步
	wg := &sync.WaitGroup{}

	// 启动子线程
	wg.Add(1)
	go func() {
		defer wg.Done()
		startCallbackSub(ctx, m, interval, 1, wg, rtdb, opctags...)
	}()
	for {
		select {
		case <-stopChan: // 如果收到停止信号，退出循环
			fmt.Printf("Worker %v stopped\n", id)
			return
		}
	}
}

func startCallbackSub(ctx context.Context, m *monitor.NodeMonitor, interval, lag time.Duration, wg *sync.WaitGroup, rtdb *redka.DB, nodes ...string) {
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

	defer cleanup(ctx, sub, wg)

	<-ctx.Done()
}

func cleanup(ctx context.Context, sub *monitor.Subscription, wg *sync.WaitGroup) {
	log.Printf("stats: sub=%d delivered=%d dropped=%d", sub.SubscriptionID(), sub.Delivered(), sub.Dropped())
	sub.Unsubscribe(ctx)
	wg.Done()
}
