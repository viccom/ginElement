package handlers

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/nalgeon/redka"
	"net/http"
	"sync"
)

type testFunc func(id string, stopChan chan struct{})
type iotFunc func(id string, stopChan chan struct{}, rtdb *redka.DB)

// 全局变量
var (
	Workers     = make(map[string]chan struct{}) // 存储所有子线程的停止通道
	workersLock sync.Mutex                       // 用于保护 Workers 的并发访问
	//nextID      = 1                              // 用于生成唯一的子线程 ID

	// 定义字符串数组
	funcCode = []string{"findmax", "simtodb", "periodicPrint", "modbus"}
	funcMap  = map[string]testFunc{
		"findmax":       findmax,
		"periodicPrint": PeriodicPrint,
	}
	// 定义字符串数组
	iotappCode = []string{"simtodb", "modbus"}
	iotappMap  = map[string]iotFunc{
		"simtodb": simtodb,
		"modbus":  handlermobus,
	}
)

// appCode=["findmax", "periodicPrint", "modbus"]
// appType=["toSouth", "toBorth", "System"]
// channel=["tcp", "udp", "serial"]
// protocol=["rtu", "tcp", "ascii", "rtuovertcp", "rtuoverudp"]
// 定义 AppConfig 结构体
type AppConfig struct {
	AppCode  string `json:"appCode"`
	AppType  string `json:"appType"`
	InstID   string `json:"instId"`
	InstName string `json:"instName"`
	Config   any    `json:"config"`
}

//// 初始化数据库
//var CfgDB *redka.DB
//var RtDB *redka.DB
//
//// 初始化数据库连接
//func InitDB(cfgdb *redka.DB, rtdb *redka.DB) {
//	CfgDB = cfgdb
//	RtDB = rtdb
//	// 在这里进行其他初始化操作
//}

// @Summary 启动线程运行函数接口
// @Description 这是一个启动线程的接口
// @Tags 示例
// @Accept json
// @Produce json
// @Param appcode path string true "功能appcode"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Router /api/v1/startWorker/{appcode} [post]
func StartWorker(c *gin.Context) {
	workersLock.Lock()
	defer workersLock.Unlock()

	appcode := c.Param("appcode")
	// 检查 appcode 是否有效
	if !contains(funcCode, appcode) {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Invalid appcode",
			"details": fmt.Sprintf("appcode '%s' is not supported", appcode),
		})
		return
	}

	// 检查 funcMap 中是否存在对应的函数
	fn, exists := funcMap[appcode]
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
		fn(uuidstr, stopChan) // 调用对应的函数
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

// @Summary 停止指定线程接口stopWorker
// @Description 这是一个停止线程的接口
// @Tags 示例
// @Accept json
// @Produce json
// @Param workerid path string true "线程 workerid"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Router /api/v1/stopWorker/{workerid} [post]
func StopWorker(c *gin.Context) {
	workersLock.Lock()
	defer workersLock.Unlock()

	// 获取子线程 ID
	workerid := c.Param("workerid")
	var workerID string
	_, err := fmt.Sscanf(workerid, "%v", &workerID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Invalid ID",
		})
		return
	}

	// 查找子线程的停止通道
	stopChan, exists := Workers[workerID]
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{
			"message": "Worker not found",
		})
		return
	}

	// 发送停止信号
	close(stopChan)

	// 从全局变量中移除子线程
	delete(Workers, workerID)

	// 返回成功消息
	c.JSON(http.StatusOK, gin.H{
		"message": "Worker stopped",
		"id":      workerID,
	})
}

// @Summary 查询后台运行的所有线程ID
// @Description 这是一个查询线程ID的接口
// @Tags 示例
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Router /api/v1/listWorkers [get]
func ListWorkers(c *gin.Context) {
	workersLock.Lock()
	defer workersLock.Unlock()

	// 获取所有子线程的 ID
	ids := make([]string, 0, len(Workers))
	for id := range Workers {
		ids = append(ids, id)
	}

	// 返回子线程 ID 列表
	c.JSON(http.StatusOK, gin.H{
		"message": "Workers running",
		"ids":     ids,
	})
}

// @Summary 查询软件当前支持的Appcode
// @Description 这是一个查询Appcode的接口
// @Tags IOTAPP Manage
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Router /api/v1/listAppcode [get]
func ListAppcode(c *gin.Context) {
	// 返回 funcCode	列表
	c.JSON(http.StatusOK, gin.H{
		"message": "appcode",
		"appcode": iotappCode,
	})
}

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
	_, err := cfgdb.Hash().Set(appConfig.AppType, uuidstr, jsonstr)
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
