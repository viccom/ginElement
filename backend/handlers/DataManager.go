package handlers

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/nalgeon/redka"
	"log"
	"net/http"
)

// 定义 DevInfo 结构体
type DevInfo struct {
	DevId string `json:"devid"`
}

// 定义 tagsInfo 结构体
type TagsInfo struct {
	DevId  string   `json:"devid"`
	TagsId []string `json:"tagsid"`
}

// @Summary 查询软件的基本信息
// @Description 这是一个查询软件基本信息的接口
// @Tags Data Manager
// @Accept json
// @Produce json
// @Param devid body DevInfo true "DevId"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Router /api/v1/getDevvalues [post]
func GetDevValues(c *gin.Context, rtdb *redka.DB) {
	var devInfo DevInfo
	if err := c.ShouldBindJSON(&devInfo); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	//fmt.Printf("devId: %v\n", devInfo.DevId)
	values, err := rtdb.Hash().Items(devInfo.DevId)

	if err != nil {
		log.Println("Error reading from database:", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Failed to read data from database",
		})
		return
	}
	//fmt.Printf("values: %v\n", values)
	if len(values) == 0 {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "no data",
		})
		return
	}
	OutterMap := make(map[string][]any)
	for key, value := range values {
		var newValue []any
		//fmt.Printf("newValue: %v\n", newValue)
		erra := json.Unmarshal([]byte(value.String()), &newValue)
		if erra != nil {
			fmt.Println("Error unmarshalling JSON:", erra)
			return
		}
		OutterMap[key] = newValue
	}
	// 返回数据库cfgdb中App配置信息 列表
	c.JSON(http.StatusOK, gin.H{
		"message": "success to read data from database",
		"data":    OutterMap,
	})
}

// @Summary 查询软件的基本信息
// @Description 这是一个查询软件基本信息的接口
// @Tags Data Manager
// @Accept json
// @Produce json
// @Param tagsinfo body TagsInfo true "tagsinfo"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Router /api/v1/getTagvalues [post]
func GetTagValues(c *gin.Context, rtdb *redka.DB) {
	var tagsInfo TagsInfo
	if err := c.ShouldBindJSON(&tagsInfo); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	values, err := rtdb.Hash().Items(tagsInfo.DevId)
	if err != nil {
		log.Println("Error reading from database:", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Failed to read data from database",
		})
		return
	}
	if len(values) == 0 {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "no data",
		})
	}
	tagslist := tagsInfo.TagsId
	OutterMap := make(map[string][]any)
	for key, value := range values {
		if contains(tagslist, key) {
			var newValue []any
			erra := json.Unmarshal([]byte(value.String()), &newValue)
			if erra != nil {
				fmt.Println("Error unmarshalling JSON:", erra)
				return
			}
			OutterMap[key] = newValue
		}

	}
	// 返回数据库cfgdb中App配置信息 列表
	c.JSON(http.StatusOK, gin.H{
		"message": "success to read data from database",
		"data":    OutterMap,
	})
}
