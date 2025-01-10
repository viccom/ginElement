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
		"simulator": {
			"appCode":   "simulator",
			"appType":   "toSouth",
			"instId":    "",
			"instName":  "simulator app",
			"autoStart": false,
			"config":    "{}",
		},
	}

	tags_default = map[string]map[string]any{
		"modbus": {
			"devId":  "DEV_7JF3ZMbgvQfvAYpo",
			"instid": "modbus@463tOZn138pdXqyz",
			"tagsMap": map[string]any{
				"bool1":    []any{"bool1", "布尔量1", 1, "01", 0, "bool"},
				"analog1":  []any{"analog1", "模拟量1", 1, "03", 0, "float32"},
				"digital1": []any{"digital1", "数字量1", 1, "03", 0, "int16"},
				"digital2": []any{"digital2", "数字量2", 1, "04", 0, "int32"},
			},
		},
		"opcda": {
			"devId":  "DEV_7JF3ZMbgvQfvAYpo",
			"instid": "opcda@g53tOZn138pdXnup",
			"tagsMap": map[string]any{
				"bool1":    []any{"bool1", "布尔量1", "bool"},
				"analog1":  []any{"analog1", "模拟量1", "float"},
				"digital1": []any{"digital1", "数字量1", "int"},
				"string1":  []any{"string1", "字符量1", "string"},
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
