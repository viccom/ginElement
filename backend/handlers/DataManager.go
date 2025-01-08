package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/nalgeon/redka"
	"log"
	"net/http"
)

// 定义 DevInfo 结构体
type DevInfo struct {
	DevId []string `json:"devid"`
}

// @Summary 查询软件的基本信息
// @Description 这是一个查询软件基本信息的接口
// @Tags Data Manager
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Router /api/v1/getRtvalues [get]
func GetDevValues(c *gin.Context, rtdb *redka.DB) {
	devid := c.Param("devid")
	values, err := rtdb.Hash().Items(devid)
	if err != nil || len(values) == 0 {
		log.Println("Error reading from database:", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Failed to read data from database",
		})
		return
	}
	OutterMap := make(map[string]string)
	for key, value := range values {
		OutterMap[key] = value.String()
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
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Router /api/v1/getRtvalues [get]
func GetTagValues(c *gin.Context, rtdb *redka.DB) {
	devid := c.Param("devid")
	values, err := rtdb.Hash().Items(devid)
	if err != nil || len(values) == 0 {
		log.Println("Error reading from database:", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Failed to read data from database",
		})
		return
	}
	OutterMap := make(map[string]string)
	for key, value := range values {
		OutterMap[key] = value.String()
	}
	// 返回数据库cfgdb中App配置信息 列表
	c.JSON(http.StatusOK, gin.H{
		"message": "success to read data from database",
		"data":    OutterMap,
	})
}
