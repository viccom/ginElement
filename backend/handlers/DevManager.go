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
	InstID  string `json:"instid"`
	Config  any    `json:"config"`
}

// 定义 DevOpt 结构体
type DevOpt struct {
	DevList []string `json:"devList"`
}

// 定义 DevTags 结构体
type DevTags struct {
	DevID   string              `json:"devId"`
	InstID  string              `json:"instid"`
	TagsMap map[string][]string `json:"tagsMap"`
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
	fmt.Println("body:", instinfo.InstId)
	instid := instinfo.InstId
	values, err3 := cfgdb.Hash().Items(DevAtInstKey)
	if err3 != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Get DevList Fail",
			"details": fmt.Sprintf("err: '%v' ", err3),
		})
	}
	if len(values) == 0 {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "no data",
		})
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
		fmt.Printf("键: %s, 值: %s\n", key, instid)
		if instid == "" {
			OutterMap[key] = newValue
		} else if instid == newValue.InstID {
			OutterMap[key] = newValue
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
	// 生成一个新的16位 UUID
	uuidstr := "DEV_" + gen16ID()
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
	// 返回数据库cfgdb中App配置信息 列表
	c.JSON(http.StatusOK, gin.H{
		"message": "Del Dev OK",
		"devlist": devOpt,
	})
}

// @Summary 向设备增加点表信息
// @Description 这是一个向设备增加点表信息的接口
// @Tags DEV Manager
// @Accept json
// @Produce json
// @Param devTags body DevTags true "new DevTags"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Router /api/v1/newdevtags [post]
func NewDevTags(c *gin.Context, cfgdb *redka.DB) {
	var devOpt DevOpt
	if err := c.ShouldBindJSON(&devOpt); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	// 返回数据库cfgdb中App配置信息 列表
	c.JSON(http.StatusOK, gin.H{
		"message": "Del Dev OK",
		"devlist": devOpt,
	})
}
