package handlers

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/nalgeon/redka"
	"log"
	"net/http"
)

// appCode=["findmax", "periodicPrint", "modbus"]
// appType=["toSouth", "toNorth", "System", "Others"]
// channel=["tcp", "udp", "serial"]
// protocol=["rtu", "tcp", "ascii", "rtuovertcp", "rtuoverudp"]
// 定义 AppConfig 结构体
type AppConfig struct {
	AppCode   string `json:"appCode"`
	AppType   string `json:"appType"`
	InstID    string `json:"instId"`
	InstName  string `json:"instName"`
	AutoStart bool   `json:"autoStart"`
	Config    any    `json:"config"`
}

// 定义 InstInfo 结构体
type InstInfo struct {
	InstId string `json:"instid"`
}

// @Summary 查询指定App的默认配置【未实现】
// @Description 这是一个查询App默认配置信息的接口
// @Tags APP Manager
// @Accept json
// @Produce json
// @Param instid body InstInfo true "InstId"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Router /api/v1/getAppDefault [post]
func GetAppDefault(c *gin.Context) {
	// 返回数据库cfgdb中App配置信息 列表
	c.JSON(http.StatusOK, gin.H{
		"message": "get app data ok",
		"data":    "newValue",
	})
}

// @Summary 查询所有App实例配置信息
// @Description 这是一个查询App配置信息的接口
// @Tags APP Manager
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Router /api/v1/listApps [get]
func ListApps(c *gin.Context, cfgdb *redka.DB) {
	values, err := cfgdb.Hash().Items(InstListKey)
	if err != nil {
		log.Println("Error reading from database:", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Failed to read data from database",
			"details": err,
		})
		return
	}
	if len(values) == 0 {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "no data",
		})
		return
	}
	OutterMap := make(map[string]AppConfig)
	for key, value := range values {
		var newValue AppConfig
		erra := json.Unmarshal([]byte(value.String()), &newValue)
		if erra != nil {
			fmt.Println("Error unmarshalling JSON:", erra)
			return
		}
		//fmt.Printf("键: %s, 值: %s\n", key, newValue)
		OutterMap[key] = newValue
	}
	// 返回数据库cfgdb中App配置信息 列表
	c.JSON(http.StatusOK, gin.H{
		"message": "success to read data from database",
		"data":    OutterMap,
	})
}

// @Summary 新建App实例
// @Description 这是一个新建App实例的接口
// @Tags APP Manager
// @Accept json
// @Produce json
// @Param appConfig body AppConfig true "new AppConfig"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Router /api/v1/newApp [post]
func NewApp(c *gin.Context, cfgdb *redka.DB) {
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

	// 检查 iotappMap 中是否存在对应的函数
	_, exists := iotappMap[appConfig.AppCode]
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

// @Summary 删除App实例【未实现】
// @Description 这是一个删除App实例的接口
// @Tags APP Manager
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Router /api/v1/delApp [post]
func DelApp(c *gin.Context, cfgdb *redka.DB) {
	var instopt InstInfo
	if err := c.ShouldBindJSON(&instopt); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	_, err := cfgdb.Hash().Delete(InstListKey, instopt.InstId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Failed to delete app instance",
			"error":   err.Error(),
		})
		return
	}
	// 返回数据库cfgdb中App配置信息 列表
	c.JSON(http.StatusOK, gin.H{
		"message": "App instance deleted",
		"instid":  instopt.InstId,
	})
}

// @Summary 修改App实例配置【未实现】
// @Description 这是一个修改App实例配置的接口
// @Tags APP Manager
// @Accept json
// @Produce json
// @Param appConfig body AppConfig true "new AppConfig"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Router /api/v1/modApp [post]
func ModApp(c *gin.Context, cfgdb *redka.DB) {
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

	// 检查 iotappMap 中是否存在对应的函数
	_, exists := iotappMap[appConfig.AppCode]
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Function not found",
			"details": fmt.Sprintf("appcode '%s' has no associated function", appConfig.AppCode),
		})
		return
	}

	// 生成一个新的16位 UUID
	uuidstr := appConfig.InstID

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

// @Summary 查询指定App配置信息
// @Description 这是一个查询App配置信息的接口
// @Tags APP Manager
// @Accept json
// @Produce json
// @Param instid body InstInfo true "InstId"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Router /api/v1/getApp [post]
func GetApp(c *gin.Context, cfgdb *redka.DB) {
	var instopt InstInfo
	if err := c.ShouldBindJSON(&instopt); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	value, err := cfgdb.Hash().Get(InstListKey, instopt.InstId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Failed to get app data",
			"error":   err.Error(),
		})
		return
	}
	valueStr := value.String()
	var newValue AppConfig
	err = json.Unmarshal([]byte(valueStr), &newValue)
	// 返回数据库cfgdb中App配置信息 列表
	c.JSON(http.StatusOK, gin.H{
		"message": "get app data ok",
		"data":    newValue,
	})
}

// @Summary 通过实例ID启动App实例【未实现】
// @Description 这是一个通过实例ID启动App实例的接口
// @Tags APP Manager
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

// @Summary 通过实例ID停止App实例【未实现】
// @Description 这是一个通过实例ID停止App实例的接口
// @Tags APP Manager
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
