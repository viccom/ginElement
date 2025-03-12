package handlers

var (
	app_default = map[string]map[string]any{
		"modbus": {
			"appCode":   "modbus",
			"appType":   "toSouth",
			"instId":    "",
			"instName":  "modbus app",
			"autoStart": false,
			"config": map[string]any{
				"channel":  "tcp",
				"host":     "127.0.0.1",
				"port":     502,
				"slaveId":  1,
				"protocol": "rtuovertcp",
			},
		},
		"opcda": {
			"appCode":   "opcda",
			"appType":   "toSouth",
			"instId":    "",
			"instName":  "opcda app",
			"autoStart": false,
			"config": map[string]any{
				"host":   "localhost",
				"progID": "Matrikon.OPC.Simulation.1",
			},
		},
		"opcua": {
			"appCode":   "opcua",
			"appType":   "toSouth",
			"instId":    "",
			"instName":  "opcua app",
			"autoStart": false,
			"config": map[string]any{
				"endpoint": "opc.tcp://localhost:49320",
				"policy":   "Security policy: None, Basic128Rsa15, Basic256, Basic256Sha256. Default: auto",
				"mode":     "Sign and Encrypt, Sign, None. Default: auto",
				"cert":     "certificate file path",
				"key":      "private key file path",
				"interval": 1,
			},
		},
		"simulator": {
			"appCode":   "simulator",
			"appType":   "toSouth",
			"instId":    "",
			"instName":  "simulator app",
			"autoStart": false,
			"config":    "{}",
		},
		"mqttpub": {
			"appCode":   "mqttpub",
			"appType":   "toNorth",
			"instId":    "",
			"instName":  "mqttpub app",
			"autoStart": false,
			"config": map[string]any{
				"broker":   "mqbroker.metme.top",
				"port":     1883,
				"username": "username",
				"password": "password",
				"cycle":    5,
				"deviceList": []string{
					"DEV_7JF3ZMbgvQfvAYpo",
					"DEV_657ZMbgvQ4368Ypo",
				},
			},
		},
		"dsTdengine": {
			"appCode":   "dsTdengine",
			"appType":   "toNorth",
			"instId":    "",
			"instName":  "dsTdengine app",
			"autoStart": false,
			"config": map[string]any{
				"host":     "localhost",
				"port":     1883,
				"username": "username",
				"password": "password",
				"database": "db01",
				"tbType":   "table",
				"cycle":    5,
				"deviceList": []string{
					"DEV_7JF3ZMbgvQfvAYpo",
					"DEV_657ZMbgvQ4368Ypo",
				},
			},
		},
	}

	tags_default = map[string]map[string]any{
		"modbus": {
			"devId":  "DEV_7JF3ZMbgvQfvAYpo",
			"instid": "modbus@463tOZn138pdXqyz",
			"tagsMap": map[string]any{
				"bool1":    []any{"bool1", "布尔量1", 1, "01", 0, "bool"},
				"analog1":  []any{"analog1", "模拟量1", 1, "03", 0, "float32"},
				"digital1": []any{"digital1", "数字量1", 1, "03", 2, "int16"},
				"digital2": []any{"digital2", "数字量2", 1, "04", 4, "int32"},
			},
		},
		"opcda": {
			"devId":  "DEV_7JF3ZMbgvQfvAYpo",
			"instid": "opcda@g53tOZn138pdXnup",
			"tagsMap": map[string]any{
				"tag1": []any{"bool1", "tag1", "布尔量1", "bool"},
				"tag2": []any{"analog1", "tag2", "模拟量1", "float"},
				"tag3": []any{"digital1", "tag3", "数字量1", "int"},
				"tag4": []any{"string1", "tag4", "字符量1", "string"},
			},
		},
		"opcua": {
			"devId":  "DEV_7JF3ZMbgvQfvAYpo",
			"instid": "opcua@g53tOZn138pdXnup",
			"tagsMap": map[string]any{
				"tag1": []any{"ns=2;s=数据类型示例.8 位设备.B 寄存器.Boolean1", "tag1", "布尔量1", "bool"},
				"tag2": []any{"ns=2;s=模拟器示例.函数.Sine1", "tag2", "模拟量1", "float"},
				"tag3": []any{"ns=2;s=模拟器示例.函数.Ramp1", "tag3", "数字量1", "int"},
				"tag4": []any{"ns=2;s=数据类型示例.8 位设备.S 寄存器.String1", "tag4", "字符量1", "string"},
			},
		},
		"simulator": {
			"devId":  "DEV_7JF3ZMbgvQfvAYpo",
			"instid": "simulator@888tOZn138pdXqyz",
			"tagsMap": map[string]any{
				"bool1":    []any{"bool1", "布尔量1", "bool"},
				"analog1":  []any{"analog1", "模拟量1", "float"},
				"digital1": []any{"digital1", "数字量1", "int"},
				"string1":  []any{"string1", "字符量1", "string"},
			},
		},
	}
)
