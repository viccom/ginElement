package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/nalgeon/redka"
	_ "github.com/taosdata/driver-go/v3/taosWS"
)

// tdengineWriteData 函数：周期性地读取redka数据并写入TDengine
func dsTDengine(id string, stopChan chan struct{}, cfgdb *redka.DB, rtdb *redka.DB) {
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

	// 获取TDengine连接配置
	host, ok := config["host"].(string)
	if !ok {
		fmt.Println("host is not a string or does not exist")
		return
	}
	port, ok := config["port"].(float64)
	if !ok {
		fmt.Println("port is not a number or does not exist")
		return
	}
	username, ok := config["username"].(string)
	if !ok {
		fmt.Println("username is not a string or does not exist")
		return
	}
	password, ok := config["password"].(string)
	if !ok {
		fmt.Println("password is not a string or does not exist")
		return
	}
	database, ok := config["database"].(string)
	if !ok {
		fmt.Println("database is not a string or does not exist")
		return
	}
	cycle, ok := config["cycle"].(float64)
	if !ok {
		fmt.Println("cycle is not a number or does not exist")
		cycle = 5 // 默认5秒
	}

	//taosTbType, ok := config["tbType"].(string)
	//if !ok {
	//	fmt.Println("tbType is not a string or does not exist")
	//}

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

	// 构建TDengine连接DSN
	taosDSN := fmt.Sprintf("%s:%s@ws(%s:%d)/%s", username, password, host, int(port), database)

	// 创建队列
	queue := NewDataQueue()

	// 生产者goroutine - 从redka读取数据
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

	// 消费者goroutine - 写入TDengine
	go func() {
		var db *sql.DB
		var err error

		// 连接TDengine
		connectTDengine := func() bool {
			if db != nil {
				db.Close()
			}

			db, err = sql.Open("taosWs", taosDSN)
			if err != nil {
				log.Printf("Failed to connect to TDengine: %v", err)
				return false
			}

			// 测试连接
			err = db.Ping()
			if err != nil {
				log.Printf("Failed to ping TDengine: %v", err)
				return false
			}
			log.Printf("Connected to TDengine successfully")
			// create database
			_, err := db.Exec("CREATE DATABASE IF NOT EXISTS " + database)
			if err != nil {
				log.Printf("Failed to create database %v, ErrMessage: %v", database, err.Error())
			}
			for _, v := range deviceList {
				values, _ := cfgdb.Hash().Items(v)
				newtag := make(map[string][]any)
				for key, value := range values {
					var newValue []any
					erra := json.Unmarshal([]byte(value.String()), &newValue)
					if erra != nil {
						fmt.Println("Error unmarshalling JSON:", erra)
						break
					}
					newtag[key] = newValue
				}
				sqlstr := CreateTableSQL(newtag)
				// create table
				_, err = db.Exec(sqlstr)
				if err != nil {
					log.Println("Failed to create stable ErrMessage: " + err.Error())
					return false
				}
			}

			return true
		}

		// 初始连接
		if !connectTDengine() {
			log.Println("Initial connection to TDengine failed, will retry...")
		}

		defer func() {
			if db != nil {
				db.Close()
			}
		}()

		for {
			select {
			case <-stopChan:
				fmt.Println("消费者收到停止信号，退出")
				return
			default:
				// 如果没有连接，尝试重连
				if db == nil {
					if !connectTDengine() {
						time.Sleep(5 * time.Second)
						continue
					}
				}

				// 处理队列中的数据
				for queue.Len() > 0 {
					if val, ok := queue.Dequeue(); ok {
						var datasmap map[string]map[string][]any
						errc := json.Unmarshal([]byte(val), &datasmap)
						if errc != nil {
							fmt.Println("解析失败:", errc)
							continue
						}

						// 遍历设备数据
						for devkey, deviceData := range datasmap {
							println(devkey, deviceData)
							err := WriteTable(db, deviceData)
							if err != nil {
								return
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
			fmt.Printf("子线程tdengine实例 %s 收到停止信号，退出\n", id)
			return
		}
	}
}

// 构建创建超级表的 SQL 语句
func CreateSuperTableSQL(tableName, devID string, fields map[string][]any) string {
	// 定义字段映射关系（JSON 数据类型 -> TDengine 数据类型）
	typeMapping := map[string]string{
		"float":  "float",
		"bool":   "bool",
		"int":    "int",
		"string": "varchar(64)",
	}
	// 构建字段部分
	var fieldParts []string
	fieldParts = append(fieldParts, "ts timestamp") // 固定字段
	for fieldName, fieldInfo := range fields {
		//info := fieldInfo.([]any)
		dataType := fieldInfo[2].(string) // 获取数据类型
		tdengineType := typeMapping[dataType]
		fieldParts = append(fieldParts, fmt.Sprintf("%s %s", fieldName, tdengineType))
	}
	// 构建 TAGS 部分
	tagsPart := fmt.Sprintf("dev_id varchar(64)")
	// 拼接完整的 SQL 语句
	sqlexc := fmt.Sprintf(
		"CREATE STABLE IF NOT EXISTS %s(\n    %s\n) TAGS (\n    %s\n);",
		tableName,
		strings.Join(fieldParts, ",\n    "),
		tagsPart,
	)
	return sqlexc
}

// 构建创建普通表的 SQL 语句
func CreateTableSQL(fields map[string][]any) string {
	// 定义字段映射关系（JSON 数据类型 -> TDengine 数据类型）
	typeMapping := map[string]string{
		"float":  "float",
		"bool":   "bool",
		"int":    "int",
		"string": "varchar(64)",
	}
	// 构建 SQL 语句
	var sqlParts []string
	for tableName, fieldInfo := range fields {
		//info := fieldInfo.([]any)
		dataType := fieldInfo[2].(string) // 获取数据类型
		tdengineType := typeMapping[dataType]
		sqlexc := fmt.Sprintf(
			"CREATE TABLE IF NOT EXISTS %s(\n    ts timestamp,\n    v %s\n);",
			tableName,
			tdengineType,
		)
		sqlParts = append(sqlParts, sqlexc)
	}
	// 拼接所有 SQL 语句
	return strings.Join(sqlParts, "\n")
}

// 写入数据到 TDengine 普通表
func WriteTable(db *sql.DB, data map[string][]any) error {
	var sqlBuilder strings.Builder
	// 遍历 JSON 数据
	for tableName, values := range data {
		// 解析值
		tsStr := values[0].(string) // 时间戳字符串
		v := values[1]              // 值
		//tsInt := values[2].(float64) // 时间戳（毫秒）
		// 将时间戳字符串转换为 time.Time
		ts, err := time.Parse("2006-01-02 15:04:05", tsStr)
		if err != nil {
			return fmt.Errorf("解析时间戳失败: %v", err)
		}
		// 构建 SQL 插入语句
		sqlexc := fmt.Sprintf("INSERT INTO %s (ts, v) VALUES ('%s', %v)", tableName, ts.Format(time.RFC3339), v)
		sqlBuilder.WriteString(sqlexc + "\n")
	}
	// 执行批量插入
	batchSQL := sqlBuilder.String()
	fmt.Printf("批量插入 SQL 语句:\n%s\n", batchSQL)
	_, err := db.Exec(batchSQL)
	if err != nil {
		return fmt.Errorf("批量插入失败: %v", err)
	}

	return nil
}
