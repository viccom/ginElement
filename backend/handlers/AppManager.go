package handlers

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/nalgeon/redka"
	"net/http"
)

// @Summary 查询数据库cfgdb中App配置信息
// @Description 这是一个查询App配置信息的接口
// @Tags IOTAPP Manage
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Router /api/v1/listApps [get]
func ListApps(c *gin.Context, cfgdb *redka.DB) {

	// 返回数据库cfgdb中App配置信息 列表
	c.JSON(http.StatusOK, gin.H{
		"message": "Workers running",
		"ids":     "ids",
	})
}

// @Summary 查询数据库cfgdb中App配置信息
// @Description 这是一个查询App配置信息的接口
// @Tags IOTAPP Manage
// @Accept json
// @Produce json
// @Param appConfig body AppConfig true "new AppConfig"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Router /api/v1/newApp [post]
func NewApp(c *gin.Context, cfgdb *redka.DB) {
	// 使用 handler.CfgDB 和 handler.RtDB
	//cfgdb := handler.CfgDB
	//rtdb := handler.RtDB
	var appConfig AppConfig
	if err := c.ShouldBindJSON(&appConfig); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	// 检查 appcode 是否有效
	if !contains(iotappCode, appConfig.AppCode) {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Invalid appcode",
			"details": fmt.Sprintf("appcode '%s' is not supported", appConfig.AppCode),
		})
		return
	}

	// 检查 funcMap 中是否存在对应的函数
	_, exists := funcMap[appConfig.AppCode]
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Function not found",
			"details": fmt.Sprintf("appcode '%s' has no associated function", appConfig.AppCode),
		})
		return
	}

	// 生成一个新的16位 UUID
	uuidstr := appConfig.AppCode + "@" + gen16ID()

	appConfig.InstID = uuidstr
	jsonstr, _ := json.Marshal(appConfig)
	// 打印 anyConfig
	//fmt.Printf("anyConfig: %+v\n", jsonstr)
	_, err := cfgdb.Hash().Set(InstListKey, uuidstr, jsonstr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "New App Creat Fail",
			"details": fmt.Sprintf("err: '%v' ", err),
		})
		return
	}
	// 返回数据库cfgdb中App配置信息 列表
	c.JSON(http.StatusOK, gin.H{
		"message":   "New App Creat OK",
		"appConfig": appConfig,
	})
}

// @Summary 查询数据库cfgdb中App配置信息
// @Description 这是一个查询App配置信息的接口
// @Tags IOTAPP Manage
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Router /api/v1/delApp [post]
func DelApp(c *gin.Context, cfgdb *redka.DB) {

	// 返回数据库cfgdb中App配置信息 列表
	c.JSON(http.StatusOK, gin.H{
		"message": "Workers running",
		"ids":     "ids",
	})
}

// @Summary 查询数据库cfgdb中App配置信息
// @Description 这是一个查询App配置信息的接口
// @Tags IOTAPP Manage
// @Accept json
// @Produce json
// @Param appConfig body AppConfig true "new AppConfig"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Router /api/v1/modApp [post]
func ModApp(c *gin.Context, cfgdb *redka.DB) {

	// 返回数据库cfgdb中App配置信息 列表
	c.JSON(http.StatusOK, gin.H{
		"message": "Workers running",
		"ids":     "ids",
	})
}

// @Summary 查询数据库cfgdb中App配置信息
// @Description 这是一个查询App配置信息的接口
// @Tags IOTAPP Manage
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Router /api/v1/getApp [get]
func GetApp(c *gin.Context, cfgdb *redka.DB) {

	// 返回数据库cfgdb中App配置信息 列表
	c.JSON(http.StatusOK, gin.H{
		"message": "Workers running",
		"ids":     "ids",
	})
}

// @Summary 通过实例ID启动App实例
// @Description 这是一个通过实例ID启动App实例的接口
// @Tags IOTAPP Manage
// @Accept json
// @Produce json
// @Param appcode path string true "功能appcode"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Router /api/v1/startApp/{appcode} [post]
func StartApp(c *gin.Context, rtdb *redka.DB) {
	workersLock.Lock()
	defer workersLock.Unlock()
	appcode := c.Param("appcode")
	// 检查 appcode 是否有效
	if !contains(iotappCode, appcode) {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Invalid appcode",
			"details": fmt.Sprintf("appcode '%s' is not supported", appcode),
		})
		return
	}
	// 检查 funcMap 中是否存在对应的函数
	fn, exists := iotappMap[appcode]
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Function not found",
			"details": fmt.Sprintf("appcode '%s' has no associated function", appcode),
		})
		return
	}
	// 生成一个新的 UUID
	newUUID := uuid.New()
	uuidstr := appcode + "@" + newUUID.String()
	// 创建停止通道
	stopChan := make(chan struct{})
	// 启动子线程
	go func() {
		defer func() {
			// 使用 select 检查 channel 是否已关闭
			select {
			case <-stopChan:
				// channel 已关闭，无需再次关闭
			default:
				close(stopChan) // 关闭 channel
			}
		}()
		fn(uuidstr, stopChan, rtdb) // 调用对应的函数
	}()
	//fn(uuidstr, stopChan) // 调用对应的函数
	// 将子线程的停止通道存储到全局变量中
	Workers[uuidstr] = stopChan
	// 返回子线程 ID
	c.JSON(http.StatusOK, gin.H{
		"message": "Worker started",
		"id":      uuidstr,
	})
}

// @Summary 通过实例ID停止App实例
// @Description 这是一个通过实例ID停止App实例的接口
// @Tags IOTAPP Manage
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Router /api/v1/stopApp [post]
func StopApp(c *gin.Context) {

	// 返回数据库cfgdb中App配置信息 列表
	c.JSON(http.StatusOK, gin.H{
		"message": "Workers running",
		"ids":     "ids",
	})
}

// @Summary 通过实例ID停止App实例
// @Description 这是一个通过实例ID停止App实例的接口
// @Tags IOTAPP Manage
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Router /api/v1/stopApp [post]
