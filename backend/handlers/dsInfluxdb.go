package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/influxdata/influxdb-client-go/v2"
	"github.com/nalgeon/redka"
)

// influxdbWriteData 函数：周期性地读取 redka 数据并写入 InfluxDB
func dsInfluxdb(id string, stopChan chan struct{}, cfgdb *redka.DB, rtdb *redka.DB) {
	// 通过 ID(实例ID) 获取实例的配置信息
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

	// 获取 InfluxDB 连接配置
	host, ok := config["host"].(string)
	if !ok {
		fmt.Println("host is not a string or does not exist")
		return
	}
	token, ok := config["token"].(string)
	if !ok {
		fmt.Println("token is not a string or does not exist")
		return
	}
	org, ok := config["org"].(string)
	if !ok {
		fmt.Println("org is not a string or does not exist")
		return
	}
	bucket, ok := config["bucket"].(string)
	if !ok {
		fmt.Println("bucket is not a string or does not exist")
		return
	}
	cycle, ok := config["cycle"].(float64)
	if !ok {
		fmt.Println("cycle is not a number or does not exist")
		cycle = 5 // 默认5秒
	}

	deviceListany, ok := config["deviceList"].([]any)
	if !ok {
		fmt.Println("deviceList is not a []string or does not exist")
	}

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

	// 通过 ID(实例ID) 获取当前函数可读写的设备配置信息和设备点表信息
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

	// 创建 InfluxDB 客户端
	client := influxdb2.NewClient(host, token)
	defer client.Close()

	// 创建写入器
	writeAPI := client.WriteAPIBlocking(org, bucket)

	// 创建队列
	queue := NewDataQueue()

	// 生产者 goroutine - 从 redka 读取数据
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
						log.Printf("Error reading from database: %v", erra)
						continue
					}
					if len(values) == 0 {
						log.Printf("%+v no value", devkey)
						continue
					} else {
						InnerMap := make(map[string][]any)
						for key, value := range values {
							var newValue []any
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

	// 消费者 goroutine - 写入 InfluxDB
	go func() {
		for {
			select {
			case <-stopChan:
				fmt.Println("消费者收到停止信号，退出")
				return
			default:
				for queue.Len() > 0 {
					if val, ok := queue.Dequeue(); ok {
						var datasmap map[string]map[string][]any
						errc := json.Unmarshal([]byte(val), &datasmap)
						if errc != nil {
							fmt.Println("解析失败:", errc)
							continue
						}

						for devkey, deviceData := range datasmap {
							for measurement, values := range deviceData {
								v := values[1] // 值
								tsFloat := values[2].(float64)
								tsInt := int64(tsFloat)
								if tsInt <= 1e12 { // 毫秒级时间戳
									tsInt = tsInt * 1000
								}

								// 构建 Line Protocol
								point := influxdb2.NewPoint(
									measurement,
									map[string]string{"dev_id": devkey},
									map[string]interface{}{"value": v},
									time.Unix(0, tsInt),
								)

								// 写入 InfluxDB
								err := writeAPI.WritePoint(context.Background(), point)
								if err != nil {
									log.Printf("写入 InfluxDB 失败: %v", err)
								}
							}
						}
					} else {
						log.Println("队列数据取出失败")
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
		case <-stopChan:
			fmt.Printf("子线程 InfluxDB 实例 %s 收到停止信号，退出\n", id)
			return
		}
	}
}
