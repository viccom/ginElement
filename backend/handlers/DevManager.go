package handlers

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/nalgeon/redka"
	"net/http"
)

// 定义 DevConfig 结构体
type DevConfig struct {
	DevID   string `json:"devId"`
	DevName string `json:"devName"`
	DevDesc string `json:"devDesc"`
	InstID  string `json:"instId"`
	Config  any    `json:"config"`
}

// 定义 DevOpt 结构体
type DevOpt struct {
	DevList []string `json:"devList"`
}

// 定义 DevTags 结构体
type DevTags struct {
	DevID   string           `json:"devId"`
	InstID  string           `json:"instid"`
	TagsMap map[string][]any `json:"tagsMap"`
}

// @Summary 获取设备配置信息
// @Description 这是一个获取设备配置信息的接口
// @Tags DEV Manager
// @Accept json
// @Produce json
// @Param instid body InstInfo true "new InstInfo"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Router /api/v1/listDevices [post]
func ListDevices(c *gin.Context, cfgdb *redka.DB) {
	var instinfo InstInfo
	if err := c.ShouldBindJSON(&instinfo); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	//
	//fmt.Println("body:", instinfo.InstId)
	instid := instinfo.InstId
	values, err3 := cfgdb.Hash().Items(DevAtInstKey)
	if err3 != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Get DevList Fail",
			"details": fmt.Sprintf("err: '%v' ", err3),
		})
		return
	}
	if len(values) == 0 {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "no data",
		})
		return
	}

	workersLock.Lock()
	defer workersLock.Unlock()

	// 获取所有子线程的 ID
	ids := make([]string, 0, len(Workers))
	for id := range Workers {
		ids = append(ids, id)
	}
	type NewConfig struct {
		DevConfig      // 嵌入 AppConfig 结构体
		IsRunning bool `json:"isRunning"` // 新增字段
	}
	OutterMap := make(map[string]NewConfig)
	var isrun bool
	for key, value := range values {
		var newValue DevConfig
		erra := json.Unmarshal([]byte(value.String()), &newValue)
		if erra != nil {
			fmt.Println("Error unmarshalling JSON:", erra)
			return
		}
		if ContainsString(ids, newValue.InstID) {
			isrun = true
		} else {
			isrun = false
		}
		//fmt.Printf("键: %+v, 值: %+v, %v\n", ids, newValue.InstID, isrun)

		newData := NewConfig{
			DevConfig: newValue,
			IsRunning: isrun, // 设置新增的 Status 字段
		}
		if instid == "" {
			OutterMap[key] = newData
		} else if instid == newValue.InstID {
			OutterMap[key] = newData
		}
	}
	if len(OutterMap) == 0 {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "no data, instid is not exist",
		})
		return
	}
	// 返回数据库cfgdb中设备配置信息 列表
	c.JSON(http.StatusOK, gin.H{
		"message": "get DevList OK",
		"data":    OutterMap,
	})
}

// @Summary 创建设备配置信息
// @Description 这是一个创建设备配置信息的接口
// @Tags DEV Manager
// @Accept json
// @Produce json
// @Param devConfig body DevConfig true "new DevConfig"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Router /api/v1/newDev [post]
func NewDev(c *gin.Context, cfgdb *redka.DB) {
	var devConfig DevConfig
	if err := c.ShouldBindJSON(&devConfig); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	instId := devConfig.InstID
	devName := devConfig.DevName
	if instId == "" || devName == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "New Dev Creat Fail",
			"details": "instid or devname is not allowed to be empty",
		})
		return
	}
	// 生成一个新的16位 UUID
	uuidstr := "DEV_" + Gen16ID()
	devConfig.DevID = uuidstr
	jsonstr, _ := json.Marshal(devConfig)
	// 打印 anyConfig
	//fmt.Printf("anyConfig: %+v\n", jsonstr)
	_, err := cfgdb.Hash().Set(DevAtInstKey, uuidstr, jsonstr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "New Dev Creat Fail",
			"details": fmt.Sprintf("err: '%v' ", err),
		})
		return
	}
	// 返回数据库cfgdb中App配置信息 列表
	c.JSON(http.StatusOK, gin.H{
		"message":   "New Dev Creat OK",
		"devConfig": devConfig,
	})
}

// @Summary 删除设备配置信息
// @Description 这是一个删除设备配置信息的接口
// @Tags DEV Manager
// @Accept json
// @Produce json
// @Param devlist body DevOpt true "del DevList"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Router /api/v1/delDev [post]
func DelDev(c *gin.Context, cfgdb *redka.DB) {
	var devOpt DevOpt
	if err := c.ShouldBindJSON(&devOpt); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	devlist := devOpt.DevList
	delResult := make(map[string]string)
	delError := make([]string, 0)
	for _, devid := range devlist {
		_, err := cfgdb.Hash().Delete(DevAtInstKey, devid)
		if err != nil {
			delResult[devid] = err.Error()
			delError = append(delError, devid)
		} else {
			delResult[devid] = "Delete OK"
		}

	}

	if len(delError) > 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Del Dev Fail",
			"details": delResult,
		})
		return
	} else {
		// 返回数据库cfgdb中App配置信息 列表
		c.JSON(http.StatusOK, gin.H{
			"message": "Del Dev OK",
			"devlist": devOpt,
		})
	}

}

// @Summary 向设备增加点表信息
// @Description 这是一个向设备增加点表信息的接口
// @Tags DEV Manager
// @Accept json
// @Produce json
// @Param devTags body DevTags true "new DevTags"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Router /api/v1/newDevtags [post]
func NewDevTags(c *gin.Context, cfgdb *redka.DB) {
	var devTags DevTags
	if err := c.ShouldBindJSON(&devTags); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	//for key, value := range devTags.TagsMap {
	//	fmt.Printf("键: %s, 值: %s\n", key, value)
	//}
	// 转换后的 map
	tagsMap := make(map[string]any)

	// 遍历原始 map
	for key, values := range devTags.TagsMap {
		jsonData, err := json.Marshal(values)
		if err != nil {
			fmt.Println("Error marshalling to JSON:", err)
			continue
		}
		tagsMap[key] = string(jsonData) // 将切片转为 JSON 字符串
	}
	_, err := cfgdb.Hash().SetMany(devTags.DevID, tagsMap)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "New Dev Creat Fail",
			"details": fmt.Sprintf("err: '%v' ", err),
		})
		return
	}
	// 返回数据库cfgdb中App配置信息 列表
	c.JSON(http.StatusOK, gin.H{
		"message": "add devTags OK",
		"devid":   devTags.DevID,
		"instid":  devTags.InstID,
		"devTags": tagsMap,
	})
}

// @Summary 查询设备点表信息
// @Description 这是一个查询设备点表信息的接口
// @Tags DEV Manager
// @Accept json
// @Produce json
// @Param devlist body DevOpt true "del DevList"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Router /api/v1/getDevtags [post]
func GetDevTags(c *gin.Context, cfgdb *redka.DB) {
	var devOpt DevOpt
	if err := c.ShouldBindJSON(&devOpt); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	// 使用 for 循环遍历数组
	newtags := make(map[string]any)
	for i := 0; i < len(devOpt.DevList); i++ {
		//fmt.Printf("Index: %d, Value: %d\n", i, numbers[i])
		devid := devOpt.DevList[i]
		values, err3 := cfgdb.Hash().Items(devid)
		if err3 != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"message": "Get tags Fail",
				"details": fmt.Sprintf("err: '%v' ", err3),
			})
		}
		if len(values) != 0 {
			newtag := make(map[string][]string)
			for key, value := range values {
				var newValue []string
				erra := json.Unmarshal([]byte(value.String()), &newValue)
				if erra != nil {
					fmt.Println("Error unmarshalling JSON:", erra)
					return
				}
				newtag[key] = newValue
			}
			newtags[devid] = newtag
		}
	}
	if len(newtags) == 0 {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "no data, or devid is not exist",
		})
		return
	}
	// 返回数据库cfgdb中App配置信息 列表
	c.JSON(http.StatusOK, gin.H{
		"message": "Get data  OK",
		"devlist": devOpt.DevList,
		"data":    newtags,
	})
}
