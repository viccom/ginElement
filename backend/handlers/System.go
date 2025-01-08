package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/nalgeon/redka"
	"log"
	"net/http"
	"runtime"
)

const AppVersion = "250108"

// @Summary 查询软件的基本信息
// @Description 这是一个查询软件基本信息的接口
// @Tags SYSTEM
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Router /api/v1/getSysinfo [get]
func GetSysInfo(c *gin.Context, cfgdb *redka.DB) {
	values, err := cfgdb.Hash().Items("system@router")
	if err != nil || len(values) == 0 {
		log.Println("Error reading from database:", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Failed to read data from database",
		})
		return
	}
	OutterMap := make(map[string]string)
	os := runtime.GOOS
	arch := runtime.GOARCH
	OutterMap["os"] = os
	OutterMap["arch"] = arch
	for key, value := range values {
		OutterMap[key] = value.String()
	}
	// 返回数据库cfgdb中App配置信息 列表
	c.JSON(http.StatusOK, gin.H{
		"message": "success to read data from database",
		"data":    OutterMap,
	})
}
