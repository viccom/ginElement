package handlers

import (
	"encoding/json"
	"fmt"
	"github.com/nalgeon/redka"
	_ "github.com/ncruces/go-sqlite3/driver"
	_ "github.com/ncruces/go-sqlite3/embed"
	"github.com/simonvetter/modbus"
	"log"
	"strconv"
	"time"
)

// ModbusRead 函数：周期性地从 10 个随机数中找到最大值并打印
func ModbusRead(id string, stopChan chan struct{}, cfgdb *redka.DB, rtdb *redka.DB) {
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
	channel, ok := config["channel"].(string)
	if !ok {
		fmt.Println("channel is not a string or does not exist")
	}
	host, ok := config["host"].(string)
	if !ok {
		fmt.Println("host is not a string or does not exist")
	}
	port, ok := config["port"].(int16)
	if !ok {
		fmt.Println("port is not a string or does not exist")
	}
	slaveId, ok := config["slaveId"].(string)
	if !ok {
		fmt.Println("slaveId is not a string or does not exist")
	}
	protocol, ok := config["protocol"].(string)
	if !ok {
		fmt.Println("protocol is not a string or does not exist")
	}
	fmt.Printf("%+v, %+v, %+v, %+v, %+v\n", channel, host, port, slaveId, protocol)

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
	mbtags := make([][]string, 0)
	// 用于存储每个标签的父设备
	mbParent := make(map[string]string, 0)
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
				mbtags = append(mbtags, newValue)
				mbParent[tagkey] = devkey
			}
		}
	}
	var client *modbus.ModbusClient
	var mbConnected = false
	client, err = modbus.NewClient(&modbus.ClientConfiguration{
		URL:     protocol + "://" + host + ":" + strconv.Itoa(int(port)),
		Speed:   9600,
		Timeout: 1 * time.Second,
	})
	if err != nil {
		defer client.Close()
	}

	mbfcode := map[string]modbus.RegType{
		"03": modbus.INPUT_REGISTER,
		"04": modbus.HOLDING_REGISTER,
	}
	for {
		select {
		case <-stopChan: // 如果收到停止信号，退出循环
			fmt.Printf("Worker %v stopped\n", id)
			return
		default:
			if !mbConnected {
				for {
					err = client.Open()
					if err != nil {
						log.Printf("Failed to connect to Modbus server at %s:%d. Retrying...\n", host, port)
						time.Sleep(1)
						continue
					}
					mbConnected = true
					break
				}
			}
			if len(mbtags) == 0 {
				return
			}
			// 遍历切片中的每个 map
			now := time.Now()
			formattedDate := now.Format("2006-01-02 15:04:05")
			unixMilliTimestamp := now.UnixMilli()
			datasmap := make(map[string]map[string]any)
			for _, m := range mbtags {
				//fmt.Printf("index: %d, deviceUnitId: %d \n", i, uint8(m["deviceUnitId"].(int)))
				// Switch to unit ID
				deviceUnitid, _ := strconv.Atoi(m[2])
				registerAddress, _ := strconv.Atoi(m[4])
				err = client.SetUnitId(uint8(deviceUnitid))
				var value any
				var errmb error
				if m[5] == "int16" {
					value, errmb = client.ReadRegister(uint16(registerAddress), mbfcode[m[3]])
					if errmb != nil {
						log.Printf("Error reading Modbus data: %v\n", errmb)
						continue
					}
				}
				if m[5] == "float32" {
					value, errmb = client.ReadFloat32(uint16(registerAddress), mbfcode[m[3]])
					if errmb != nil {
						log.Printf("Error reading Modbus data: %v\n", errmb)
						continue
					}
				}
				tagid := m[0]
				valueMap := []any{formattedDate, value, unixMilliTimestamp}
				valueMapJson, _ := json.Marshal(valueMap)
				devkey := mbParent[tagid]
				if datasmap[devkey] == nil {
					datasmap[devkey] = make(map[string]any)
				}
				datasmap[devkey][tagid] = valueMapJson
			}

			//	统一将数据写入到redka数据库
			for devkey := range datasmap {
				_, errz := rtdb.Hash().SetMany(devkey, datasmap[devkey])
				if errz != nil {
					fmt.Printf("Err: %v\n", errz)
					continue
				}
			}
			time.Sleep(1 * time.Second)
		}
	}
}
