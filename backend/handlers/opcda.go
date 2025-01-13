package handlers

import (
	"encoding/json"
	//"encoding/json"
	"fmt"
	"github.com/huskar-t/opcda"
	"github.com/huskar-t/opcda/com"
	"github.com/nalgeon/redka"
	"log"
	"time"
)

// OpcDARead函数：去设备点表中获取配置信息，然后连接OPC Server订阅数据
func OpcDARead(id string, stopChan chan struct{}, cfgdb *redka.DB, rtdb *redka.DB) {
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
	host, ok := config["host"].(string)
	if !ok {
		fmt.Println("host is not a string or does not exist")
	}
	progID, ok := config["progID"].(string)
	if !ok {
		fmt.Println("progID is not a string or does not exist")
	}
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
	if len(opctags) == 0 {
		fmt.Printf("instid %v no tag\n", id)
		return
	}
	//从OPCDA Server读取数据处理逻辑
	com.Initialize()
	defer com.Uninitialize()
	server, err := opcda.Connect(progID, host)
	if err != nil {
		log.Printf("connect to opc server failed: %s\n", err)
		return
	}
	defer server.Disconnect()
	// 使用当前时间的纳秒级时间戳作为种子
	groups := server.GetOPCGroups()
	group, err := groups.Add("group1")
	if err != nil {
		log.Printf("add group failed: %s\n", err)
	}
	items := group.OPCItems()
	itemList, errs, err := items.AddItems(opctags)
	if err != nil {
		log.Printf("add items failed: %s\n", err)
	}
	for i, err := range errs {
		if err != nil {
			log.Printf("add item %s failed: %s\n", opctags[i], err)
		}
	}
	// Wait for the OPC server to be ready
	time.Sleep(time.Second * 2)
	ch := make(chan *opcda.DataChangeCallBackData, 1000)
	go func() {
		for {
			select {

			case data := <-ch:
				//fmt.Printf("data change received, transaction id: %d, group handle: %d, masterQuality: %d, masterError: %v\n", data.TransID, data.GroupHandle, data.MasterQuality, data.MasterErr)
				datasmap := make(map[string]map[string]any)
				for i := 0; i < len(data.ItemClientHandles); i++ {
					opcitem := ""
					for _, item := range itemList {
						if item.GetClientHandle() == data.ItemClientHandles[i] {
							opcitem = trimInvisible(item.GetItemID())
						}
					}
					// 将 data.Values[i] 转换为字符串
					unixTime := data.TimeStamps[i].Unix()
					timestampstr := data.TimeStamps[i].Format("2006-01-02 15:04:05")
					//valueStr := fmt.Sprintf("%v", data.Values[i])
					//quality := uint8(data.Qualities[i])
					//fmt.Printf("data : %s %s %d %s %d\n", opcitem, timestampstr, quality, valueStr, unixTime)
					//	将数据增加到设备数据集合中
					valueMap := []any{timestampstr, data.Values[i], unixTime}
					valueMapJson, _ := json.Marshal(valueMap)
					devkey := opcParent[opcitem]
					if datasmap[devkey] == nil {
						datasmap[devkey] = make(map[string]any)
					}
					tagkey := opcBind[opcitem]
					datasmap[devkey][tagkey] = valueMapJson
					//fmt.Println(tagParent[tag])
				}
				//	统一将数据写入到redka数据库
				for devkey := range datasmap {
					_, errz := rtdb.Hash().SetMany(devkey, datasmap[devkey])
					if errz != nil {
						fmt.Printf("Err: %v\n", errz)
						continue
					}
				}
			}
		}
	}()
	err = group.RegisterDataChange(ch)
	if err != nil {
		log.Printf("register data change failed: %s\n", err)
	}
	log.Println("Registered data change in OPCDA")
	select {
	case <-stopChan:
		group.Release()
		fmt.Printf("子线程OPCDA实例 %s 收到停止信号，退出\n", id)
		err := server.Disconnect()
		if err != nil {
			return
		} // 断开连接
		return
	}
}
