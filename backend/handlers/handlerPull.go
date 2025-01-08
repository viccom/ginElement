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

var (
	// stopChan   = make(chan struct{})
	// data_queue = make(chan map[string]any, 100)
	db *redka.DB
)

type Config struct {
	Modbus struct {
		DeviceID string
		Channel  string
		Protocol string
		Host     string
		Port     int16
		Tagsfile string
	}
	Mqtt struct {
		Broker   string
		Port     int
		ID       string
		Username string
		Password string
		Topic    string
	}
	Reconnect struct {
		Delay time.Duration
	}
}

var config Config

type PUBData struct {
	DeviceUnitID    uint8
	TagID           string
	RegisterAddress uint16
	DataType        string
	Value           any
}

// handlermobus 函数：周期性地从 10 个随机数中找到最大值并打印
func handlermobus(id string, stopChan chan struct{}, cfgdb *redka.DB, rtdb *redka.DB) {
	_ = rtdb.Key()
	for {
		select {
		case <-stopChan: // 如果收到停止信号，退出循环
			fmt.Printf("Worker %v stopped\n", id)
			return
		default:
			// 打印结果
			fmt.Printf("Worker %v\n", id)
			// 等待 1 秒
			time.Sleep(1 * time.Second)
		}
	}
}

func readModbusData(config Config) {
	var client *modbus.ModbusClient
	var err error
	var mbConnected = false
	client, err = modbus.NewClient(&modbus.ClientConfiguration{
		URL:     config.Modbus.Protocol + "://" + config.Modbus.Host + ":" + strconv.Itoa(int(config.Modbus.Port)),
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
	arr := []map[string]string{
		{"device_unitid": "1", "tag_id": "i_Temperature", "fc": "04", "register_address": "0", "dataType": "int16"},
		{"device_unitid": "1", "tag_id": "i_humidity", "fc": "04", "register_address": "1", "dataType": "int16"},
		{"device_unitid": "1", "tag_id": "f_Temperature", "fc": "04", "register_address": "2", "dataType": "float32"},
		{"device_unitid": "1", "tag_id": "f_humidity", "fc": "04", "register_address": "4", "dataType": "float32"},
	}
	deviceID := config.Modbus.DeviceID
	configKey := deviceID + "@config"
	jsonCfg, _ := json.Marshal(config.Modbus)
	_, err = db.Hash().Set(configKey, deviceID, jsonCfg)
	if err != nil {
		log.Printf("Error writing device %v config to database: %v\n", deviceID, err)
	} else {
		fmt.Printf("Write device  %v config to database: %v\n", deviceID, config.Modbus)
	}
	tagsKey := deviceID + "@tags"
	jsonTags, _ := json.Marshal(arr)
	_, _ = db.Set().Add("DeviceList", deviceID)
	_, err = db.Hash().Set(tagsKey, deviceID, jsonTags)
	if err != nil {
		log.Printf("Error writing device %v tags to database: %v\n", deviceID, err)
	} else {
		fmt.Printf("Write device  %v tags to database: %v\n", deviceID, arr)
	}
	for {
		select {
		default:
			if !mbConnected {
				for {
					err = client.Open()
					if err != nil {
						log.Printf("Failed to connect to Modbus TCP server at %s:%d. Retrying...\n", config.Modbus.Host, config.Modbus.Port)
						time.Sleep(config.Reconnect.Delay)
						continue
					}
					mbConnected = true
					break
				}
			}

			// 遍历切片中的每个 map
			var mbdataarr []PUBData
			mbhashtb := make(map[string]any)
			for _, m := range arr {
				//fmt.Printf("index: %d, deviceUnitId: %d \n", i, uint8(m["deviceUnitId"].(int)))
				// Switch to unit ID
				deviceUnitid, _ := strconv.Atoi(m["device_unitid"])
				registerAddress, _ := strconv.Atoi(m["register_address"])
				err = client.SetUnitId(uint8(deviceUnitid))
				var modbusData PUBData
				if m["dataType"] == "int16" {
					var intValue uint16
					intValue, erra1 := client.ReadRegister(uint16(registerAddress), mbfcode[m["fc"]])
					if erra1 != nil {
						log.Printf("Error reading Modbus data: %v\n", erra1)
						continue
					}
					modbusData = PUBData{
						DeviceUnitID:    uint8(deviceUnitid),
						TagID:           m["tag_id"],
						RegisterAddress: uint16(registerAddress),
						DataType:        m["dataType"],
						Value:           intValue,
					}
				}
				if m["dataType"] == "float32" {
					var floatValue float32
					floatValue, erra2 := client.ReadFloat32(uint16(registerAddress), mbfcode[m["fc"]])
					if erra2 != nil {
						log.Printf("Error reading Modbus data: %v\n", erra2)
						continue
					}
					modbusData = PUBData{
						DeviceUnitID:    uint8(deviceUnitid),
						TagID:           m["tag_id"],
						RegisterAddress: uint16(registerAddress),
						DataType:        m["dataType"],
						Value:           floatValue,
					}
				}
				jsonmb, _ := json.Marshal(modbusData)
				mbhashtb[m["tag_id"]] = string(jsonmb)
				mbdataarr = append(mbdataarr, modbusData)

			}
			//将Modbus数据写入redka内存数据库的哈希表deviceID中

			_, err = db.Hash().SetMany(deviceID, mbhashtb)
			if err != nil {
				log.Printf("Error writing to database: %v\n", err)
				continue
			} else {
				fmt.Printf("Write data to database: %v\n", mbhashtb)
			}
			//for _, v := range mbdataarr {
			//	_, err = db.Hash().Set(config.Modbus.DeviceID, v.TagID, v.Value)
			//}
			jsonData, err := json.Marshal(mbdataarr)
			if err != nil {
				fmt.Printf("Error marshaling Modbus data to JSON: %v\n", err)
				continue
			}
			//dataQueue.Enqueue(string(jsonData))
			fmt.Printf("Read data from Modbus: %s\n", jsonData)
			time.Sleep(1 * time.Second)
		}
	}
}
