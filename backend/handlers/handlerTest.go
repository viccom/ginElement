package handlers

import (
	"encoding/json"
	"fmt"
	"github.com/nalgeon/redka"
	"math/rand"
	"time"
)

// 判断字符串是否在数组中
func contains(arr []string, target string) bool {
	// 将数组转换为 map
	set := make(map[string]struct{})
	for _, item := range arr {
		set[item] = struct{}{}
	}

	// 判断目标字符串是否在 map 中
	_, exists := set[target]
	return exists
}

// 打印当前时间
func printTime() {
	// 定义时区
	var (
		beijingLocation    = time.FixedZone("CST", 8*60*60)  // 北京时间 (UTC+8)
		washingtonLocation = time.FixedZone("EDT", -4*60*60) // 华盛顿时间 (UTC-4)
		moscowLocation     = time.FixedZone("MSK", 3*60*60)  // 莫斯科时间 (UTC+3)
	)
	// 获取当前 UTC 时间
	now := time.Now().UTC()

	// 转换为不同时区的时间
	beijingTime := now.In(beijingLocation)
	washingtonTime := now.In(washingtonLocation)
	moscowTime := now.In(moscowLocation)

	// 打印时间
	fmt.Printf("北京时间: %s\n", beijingTime.Format("2006-01-02 15:04:05"))
	fmt.Printf("华盛顿时间: %s\n", washingtonTime.Format("2006-01-02 15:04:05"))
	fmt.Printf("莫斯科时间: %s\n", moscowTime.Format("2006-01-02 15:04:05"))
	fmt.Println("----------------------------------------")
}

// periodicPrint函数，周期打印时间
func PeriodicPrint(id string, stopChan chan struct{}) {
	for {
		select {
		case <-stopChan: // 如果收到停止信号，退出循环
			fmt.Printf("Worker %v stopped\n", id)
			return
		default:
			printTime()
			// 等待 1 秒
			time.Sleep(5 * time.Second)
		}
	}
}

// findmax函数：周期性地从 10 个随机数中找到最大值并打印
func findmax(id string, stopChan chan struct{}) {
	// 使用当前时间的纳秒级时间戳作为种子
	source := rand.NewSource(time.Now().UnixNano())
	r := rand.New(source)
	for {
		select {
		case <-stopChan: // 如果收到停止信号，退出循环
			fmt.Printf("Worker %v stopped\n", id)
			return
		default:
			// 生成 10 个随机数
			nums := make([]int, 10)
			for i := 0; i < 10; i++ {
				nums[i] = r.Intn(100)
			}
			// 找到最大值
			m := nums[0]
			for _, num := range nums {
				if num > m {
					m = num
				}
			}
			// 打印结果
			fmt.Printf("Worker %v: Max number is %d\n", id, m)
			// 等待 1 秒
			time.Sleep(1 * time.Second)
		}
	}
}

// Simulator函数：去设备点表中获取配置信息，然后模拟数据
func Simulator(id string, stopChan chan struct{}, cfgdb *redka.DB, rtdb *redka.DB) {
	// 使用当前时间的纳秒级时间戳作为种子
	source := rand.NewSource(time.Now().UnixNano())
	r := rand.New(source)
	// 通过ID(实例ID)获取当前函数可读写的设备配置信息和设备点表信息
	values, err1 := cfgdb.Hash().Items(DevAtInstKey)
	if err1 != nil {
		fmt.Printf("Err: %v\n", err1)
		return
	}
	if len(values) == 0 {
		fmt.Printf("database no any device\n")
		return
	}
	OutterMap := make(map[string]DevConfig)
	for key, value := range values {
		var newValue DevConfig
		erra := json.Unmarshal([]byte(value.String()), &newValue)
		if erra != nil {
			fmt.Println("Error unmarshalling JSON:", erra)
			return
		}
		//fmt.Printf("键: %s, Queryid: %s, InstID: %s\n", key, id, newValue.InstID)
		if id == newValue.InstID {
			OutterMap[key] = newValue
		}
	}
	if len(OutterMap) == 0 {
		fmt.Printf("instid %v no match device\n", id)
		return
	}
	for {
		select {
		case <-stopChan: // 如果收到停止信号，退出循环
			now := time.Now()
			formattedDate := now.Format("2006-01-02 15:04:05")
			fmt.Printf("%v 子线程Simulator实例 %v stopped\n", formattedDate, id)
			return
		default:
			for devkey := range OutterMap {
				// 从设备点表中获取配置信息
				tags, err2 := cfgdb.Hash().Items(devkey)
				if err2 != nil {
					fmt.Printf("Err: %v\n", err2)
					continue
				}
				if len(tags) != 0 {
					datasmap := make(map[string]any)
					loc, _ := time.LoadLocation("Local")
					// 获取当前时间（基于本地时区）
					now := time.Now().In(loc)
					formattedDate := now.Format("2006-01-02 15:04:05")
					unixMilliTimestamp := now.UnixMilli()
					// 遍历设备点表获取数据
					for tagkey, tagvalue := range tags {
						var newValue []string
						erra := json.Unmarshal([]byte(tagvalue.String()), &newValue)
						if erra != nil {
							fmt.Println("Error unmarshalling JSON:", erra)
							return
						}
						// 模拟数据
						var value interface{}
						if newValue[2] == "int" {
							value = r.Intn(100)
						}
						if newValue[2] == "float" {
							value = r.Float32() * 100
						}
						if newValue[2] == "bool" {
							value = pickRandomElement(boolArr)
						}
						if newValue[2] == "string" {
							value = pickRandomElement(stringArr)
						}

						valueMap := []any{formattedDate, value, unixMilliTimestamp}
						valueMapJson, _ := json.Marshal(valueMap)
						datasmap[tagkey] = valueMapJson

					}
					//fmt.Printf("设备： %s, 标签: %s, 数值三元组: %v\n", devkey, tagkey, valueMap)
					//	统一将数据写入到redka数据库
					_, err := rtdb.Hash().SetMany(devkey, datasmap)
					if err != nil {
						fmt.Printf("Err: %v\n", err)
						return
					}

				}

			}
			// 等待 1 秒
			time.Sleep(1 * time.Second)
		}
	}
}
