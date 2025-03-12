package handlers

import (
	"encoding/json"
	"fmt"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/nalgeon/redka"
	"log"
	"time"
)

// mqttPubData 函数：周期性地读取modbus设备数据
func mqttPubData(id string, stopChan chan struct{}, cfgdb *redka.DB, rtdb *redka.DB) {
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
	config, ok := configMap.(map[string]any) // 类型断言为 map[string]any
	if !ok {
		fmt.Println("configMap is not a map[string]any or does not exist")
		return
	}
	broker, ok := config["broker"].(string)
	if !ok {
		fmt.Println("broker is not a string or does not exist")
	}
	port, ok := config["port"].(float64)
	if !ok {
		fmt.Println("port is not a int or does not exist")
	}
	username, ok := config["username"].(string)
	if !ok {
		fmt.Println("username is not a int or does not exist")
	}
	password, ok := config["password"].(string)
	if !ok {
		fmt.Println("password is not a string or does not exist")
	}
	cycle, ok := config["cycle"].(float64)
	if !ok {
		fmt.Println("cycle is not a string or does not exist")
	}
	deviceListany, ok := config["deviceList"].([]any)
	if !ok {
		fmt.Println("deviceList is not a []string or does not exist")
	}
	fmt.Printf("%+v\n", deviceListany)
	var deviceList []string
	if len(deviceListany) != 0 {
		for _, item := range deviceListany {
			device, ok := item.(string)
			if !ok {
				fmt.Println("deviceList contains non-string values")
				return
			}
			deviceList = append(deviceList, device)
		}
	}
	//fmt.Printf("%+v, %+v, %+v, %+v, %+v, %+v\n", broker, port, username, password, cycle, deviceList)

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
		if len(deviceList) == 0 {
			devMap[key] = newValue
		} else {
			if ContainsString(deviceList, key) {
				devMap[key] = newValue
			}
		}
	}
	if len(devMap) == 0 {
		fmt.Printf("no match device in %+v\n", deviceList)
		return
	}

	var f mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
		fmt.Printf("TOPIC: %s\n", msg.Topic())
		fmt.Printf("MSG: %s\n", msg.Payload())
	}
	opts := mqtt.NewClientOptions()
	opts.AddBroker(fmt.Sprintf("tcp://%s:%d", broker, int(port)))
	opts.SetClientID(id)
	opts.SetUsername(username)
	opts.SetPassword(password)
	opts.SetDefaultPublishHandler(f)
	mqClient := mqtt.NewClient(opts)
	// 通过设备ID获取设备点表信息

	// 创建队列
	queue := NewDataQueue()
	// 生产者goroutine
	go func() {
		for {
			select {
			case <-stopChan:
				fmt.Println("生产者收到停止信号，退出")
				return
			default:
				if queue.Len() > 1000 {
					log.Printf("队列长度超过1000，等待消费")
					time.Sleep(1 * time.Second)
					continue
				}
				OutterMap := make(map[string]map[string][]any)
				for devkey := range devMap {
					values, erra := rtdb.Hash().Items(devkey)
					if erra != nil {
						log.Printf("Error reading from database: ", erra)
						continue
					}
					if len(values) == 0 {
						log.Printf("%+v no value", devkey)
						continue
					} else {
						InnerMap := make(map[string][]any)
						for key, value := range values {
							var newValue []any
							//fmt.Printf("newValue: %v\n", newValue)
							errb := json.Unmarshal([]byte(value.String()), &newValue)
							if errb != nil {
								fmt.Println("Error unmarshalling JSON:", errb)
								return
							}
							InnerMap[key] = newValue
						}
						OutterMap[devkey] = InnerMap
					}

				}
				OutterMapstr, _ := json.Marshal(OutterMap)
				queue.Enqueue(string(OutterMapstr))
				time.Sleep(time.Duration(cycle) * time.Second)
			}
		}
	}()

	// 消费者goroutine
	go func() {
		for {
			select {
			case <-stopChan:
				fmt.Println("消费者收到停止信号，退出")
				return
			default:
				// 检查MQTT连接状态，如果未连接则尝试连接
				if !mqClient.IsConnected() {
					for {
						token := mqClient.Connect()
						if token.Wait() && token.Error() == nil {
							break
						}
						log.Printf("Failed to connect to MQTT broker. Retrying... Error: %v\n", token.Error())
						time.Sleep(5)
						if _, ok := <-stopChan; ok {
							mqClient.Disconnect(250)
							return
						}
					}
				}
				var datasmap map[string]map[string]any
				for queue.Len() > 0 {
					if val, ok := queue.Dequeue(); ok {
						errc := json.Unmarshal([]byte(val), &datasmap)
						if errc != nil {
							fmt.Println("解析失败:", errc)
							continue
						}
						for devkey := range datasmap {
							pubDatastr, _ := json.Marshal(datasmap[devkey])
							token := mqClient.Publish(devkey+"/datas", 0, false, pubDatastr)
							// 发布数据到MQTT
							if token.Wait() && token.Error() != nil {
								log.Printf("Error publishing to MQTT: %v\n", token.Error())
							} else {
								fmt.Printf("Published data to MQTT: %s\n", datasmap[devkey])
							}
						}

					}
				}
				if queue.Len() == 0 {
					time.Sleep(time.Duration(cycle) * time.Second)
				}
			}
		}
	}()

	// 当前线程处理退出信号
	for {
		select {
		case <-stopChan: // 如果收到停止信号，退出循环
			fmt.Printf("子线程mqttPub实例 %s 收到停止信号，退出\n", id)
			return
		}
	}
}
