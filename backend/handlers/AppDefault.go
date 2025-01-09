package handlers

var (
	app_default = map[string]map[string]any{
		"modbus": {"appcode": "modbus",
			"appType":   "modbus",
			"instid":    "",
			"instName":  "modbus app",
			"autoStart": false,
			"config":    "{}",
		},
		"simulator": {
			"appcode":   "simulator",
			"instid":    "",
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
				"bool1":    []any{"bool1", "布尔量1", "bool"},
				"analog1":  []any{"analog1", "模拟量1", "float"},
				"digital1": []any{"digital1", "数字量1", "int"},
				"string1":  []any{"string1", "字符量1", "string"},
			},
		},
		"simulator": {
			"devId":  "DEV_7JF3ZMbgvQfvAYpo",
			"instid": "simulator@463tOZn138pdXqyz",
			"tagsMap": map[string]any{
				"bool1":    []any{"bool1", "布尔量1", "bool"},
				"analog1":  []any{"analog1", "模拟量1", "float"},
				"digital1": []any{"digital1", "数字量1", "int"},
				"string1":  []any{"string1", "字符量1", "string"},
			},
		},
	}
)
