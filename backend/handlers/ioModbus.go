package handlers

import (
	"encoding/json"
	"fmt"
	"github.com/nalgeon/redka"
	"github.com/simonvetter/modbus"
	"log"
	_ "modernc.org/sqlite"
	"strconv"
	"time"
)

// ModbusRead 函数：周期性地读取 Modbus 设备数据
func ModbusRead(id string, stopChan chan struct{}, cfgdb *redka.DB, rtdb *redka.DB) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("ModbusRead 发生 panic: %v\n", r)
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

	channel, ok := config["channel"].(string)
	if !ok {
		log.Println("channel 不是字符串或不存在")
	}
	host, ok := config["host"].(string)
	if !ok {
		log.Println("host 不是字符串或不存在")
	}
	port, ok := config["port"].(float64)
	if !ok {
		log.Println("port 不是整数或不存在")
	}
	slaveId, ok := config["slaveId"].(float64)
	if !ok {
		log.Println("slaveId 不是整数或不存在")
	}
	protocol, ok := config["protocol"].(string)
	if !ok {
		log.Println("protocol 不是字符串或不存在")
	}
	log.Printf("%+v, %+v, %+v, %+v, %+v\n", channel, host, port, slaveId, protocol)

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

	// 通过设备ID获取设备点表信息
	mbtags := make([][]string, 0)
	mbParent := make(map[string]string)
	for devkey := range devMap {
		tags, err2 := cfgdb.Hash().Items(devkey)
		if err2 != nil {
			log.Printf("获取设备点表失败: %v\n", err2)
			continue
		}
		if len(tags) != 0 {
			for tagkey, tagvalue := range tags {
				var newValue []any
				erra := json.Unmarshal([]byte(tagvalue.String()), &newValue)
				if erra != nil {
					log.Println("解析 JSON 失败:", erra)
					return
				}
				var strValues []string
				for _, v := range newValue {
					strValues = append(strValues, fmt.Sprintf("%v", v))
				}
				mbtags = append(mbtags, strValues)
				mbParent[tagkey] = devkey
			}
		}
	}
	if len(mbtags) == 0 {
		log.Printf("实例ID %v 没有标签\n", id)
		return
	}

	var client *modbus.ModbusClient
	var mbConnected = false

	// 连接 Modbus 服务器
	connect := func() error {
		client, err = modbus.NewClient(&modbus.ClientConfiguration{
			URL:     protocol + "://" + host + ":" + strconv.Itoa(int(port)),
			Speed:   9600,
			Timeout: 1 * time.Second,
		})
		if err != nil {
			return fmt.Errorf("创建 Modbus 客户端失败: %v", err)
		}

		err = client.Open()
		if err != nil {
			return fmt.Errorf("连接 Modbus 服务器失败: %v", err)
		}

		mbConnected = true
		log.Println("成功连接到 Modbus 服务器")
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
				log.Printf("尝试连接 Modbus 服务器\n")
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
		if client != nil {
			client.Close()
		}
	}()

	mbfcode := map[string]modbus.RegType{
		"03": modbus.INPUT_REGISTER,
		"04": modbus.HOLDING_REGISTER,
	}

	// 监听停止信号
	for {
		select {
		case <-stopChan:
			log.Printf("子线程 Modbus 实例 %s 收到停止信号，退出\n", id)
			return
		default:
			// 检查连接状态
			if !mbConnected {
				log.Println("检测到连接断开，尝试重新连接")
				reconnect()
				continue
			}

			// 读取数据
			loc, _ := time.LoadLocation("Local")
			now := time.Now().In(loc)
			formattedDate := now.Format("2006-01-02 15:04:05")
			unixMilliTimestamp := now.UnixMilli()
			datasmap := make(map[string]map[string]any)

			for _, m := range mbtags {
				deviceUnitid, _ := strconv.Atoi(m[2])
				registerAddress, _ := strconv.Atoi(m[4])
				fccode := m[3]
				dataType := m[5]

				err = client.SetUnitId(uint8(deviceUnitid))
				if err != nil {
					log.Printf("设置 Unit ID 失败: %v\n", err)
					mbConnected = false
					break
				}

				var value any
				var errmb error
				switch {
				case dataType == "int16":
					value, errmb = client.ReadRegister(uint16(registerAddress), mbfcode[fccode])
				case dataType == "float32":
					value, errmb = client.ReadFloat32(uint16(registerAddress), mbfcode[fccode])
				case dataType == "bool" && fccode == "01":
					value, errmb = client.ReadCoil(uint16(registerAddress))
				case dataType == "bool" && fccode == "02":
					value, errmb = client.ReadDiscreteInput(uint16(registerAddress))
				default:
					log.Printf("不支持的数据类型: %s\n", dataType)
					continue
				}

				if errmb != nil {
					log.Printf("读取 Modbus 数据失败: %v\n", errmb)
					mbConnected = false
					break
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

			// 统一将数据写入到 redka 数据库
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
