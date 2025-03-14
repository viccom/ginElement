package handlers

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/nalgeon/redka"
	"log"
	"net/http"
	"time"
)

// appCode=["simulator", "modbus", "opcada"]
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

// 定义 AppInfo 结构体
type AppInfo struct {
	AppCode string `json:"appCode"`
}

// @Summary 查询指定App的默认配置
// @Description 这是一个查询App默认配置信息的接口
// @Tags APP Manager
// @Accept json
// @Produce json
// @Param appCode body AppInfo true "AppCode"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Router /api/v1/getAppDefault [post]
func GetAppDefault(c *gin.Context) {
	var appinfo AppInfo
	if err := c.ShouldBindJSON(&appinfo); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	//fmt.Printf("%+v\n", appinfo.AppCode)
	appcode := appinfo.AppCode
	var newValue = make(map[string]any)
	newValue["appConfig"] = app_default[appcode]
	newValue["devTags"] = tags_default[appcode]
	if isEmptyMap(newValue["appConfig"]) {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "appCode is not exist",
		})
		return
	}

	fmt.Printf("newValue: %v\n", newValue["appConfig"])
	// 返回数据库cfgdb中App配置信息 列表
	c.JSON(http.StatusOK, gin.H{
		"message": "get app_default ok",
		"data":    newValue,
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
	workersLock.Lock()
	defer workersLock.Unlock()

	// 获取所有子线程的 ID
	ids := make([]string, 0, len(Workers))
	for id := range Workers {
		ids = append(ids, id)
	}
	type NewConfig struct {
		AppConfig      // 嵌入 AppConfig 结构体
		IsRunning bool `json:"isRunning"` // 新增字段
	}
	OutterMap := make(map[string]NewConfig)
	for key, value := range values {
		var newValue AppConfig
		erra := json.Unmarshal([]byte(value.String()), &newValue)
		if erra != nil {
			fmt.Println("Error unmarshalling JSON:", erra)
			return
		}

		var isRunning bool
		if contains(ids, key) {
			isRunning = true
		}
		config := NewConfig{
			AppConfig: newValue,
			IsRunning: isRunning, // 设置新增的 Status 字段
		}
		//fmt.Printf("键: %s, 值: %s\n", key, newValue)
		OutterMap[key] = config
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
// @Param appConfig body AppConfig true "AppConfig"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Router /api/v1/newApp [post]
func NewApp(c *gin.Context, cfgdb *redka.DB) {
	var appConfig AppConfig
	if err := c.ShouldBindJSON(&appConfig); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	// 检查 appCode 是否有效
	fmt.Printf("%+v,appConfig: %+v\n", iotappCode, appConfig)
	if !contains(iotappCode, appConfig.AppCode) {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Invalid appCode",
			"details": fmt.Sprintf("appCode '%s' is not supported", appConfig.AppCode),
		})
		return
	}

	// 检查 iotappMap 中是否存在对应的函数
	_, exists := IotappMap[appConfig.AppCode]
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Function not found",
			"details": fmt.Sprintf("appCode '%s' has no associated function", appConfig.AppCode),
		})
		return
	}

	// 生成一个新的16位 UUID
	uuidstr := appConfig.AppCode + "@" + Gen16ID()

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

// @Summary 删除App实例
// @Description 这是一个删除App实例的接口
// @Tags APP Manager
// @Accept json
// @Produce json
// @Param instid body InstInfo true "InstId"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Router /api/v1/delApp [post]
func DelApp(c *gin.Context, cfgdb *redka.DB) {
	var instopt InstInfo
	if err := c.ShouldBindJSON(&instopt); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	workersLock.Lock()
	defer workersLock.Unlock()
	instid := instopt.InstId
	// 查找子线程的停止通道
	stopChan, exists := Workers[instid]
	if exists {
		// 发送停止信号
		close(stopChan)
		// 从全局变量中移除子线程
		delete(Workers, instid)
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
		"data":    instopt,
	})
}

// @Summary 修改App实例配置
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
			"message": "Invalid appCode",
			"details": fmt.Sprintf("appCode '%s' is not supported", appConfig.AppCode),
		})
		return
	}

	// 检查 iotappMap 中是否存在对应的函数
	_, exists := IotappMap[appConfig.AppCode]
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Function not found",
			"details": fmt.Sprintf("appCode '%s' has no associated function", appConfig.AppCode),
		})
		return
	}

	// 获取inst UUID
	uuidstr := appConfig.InstID
	if uuidstr == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "instId is nil",
			"details": fmt.Sprintf("instId '%s' is nil", uuidstr),
		})
		return
	}
	isExist, _ := cfgdb.Hash().Exists(InstListKey, uuidstr)
	if !isExist {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Invalid instId",
			"details": fmt.Sprintf("instId '%s' is not exist", uuidstr),
		})
		return
	}

	jsonstr, _ := json.Marshal(appConfig)
	// 打印 anyConfig
	fmt.Printf("appConfig: %+v\n", jsonstr)
	_, errb := cfgdb.Hash().Set(InstListKey, uuidstr, jsonstr)
	if errb != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "New App Creat Fail",
			"details": fmt.Sprintf("err: '%v' ", errb),
		})
		return
	}
	// 返回数据库cfgdb中App配置信息 列表
	c.JSON(http.StatusOK, gin.H{
		"message": "New App Mod OK",
		"data":    appConfig,
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

// @Summary 通过实例ID启动App实例
// @Description 这是一个通过实例ID启动App实例的接口
// @Tags APP Manager
// @Accept json
// @Produce json
// @Param instid body InstInfo true "InstId"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Router /api/v1/startApp [post]
func StartApp(c *gin.Context, cfgdb *redka.DB, rtdb *redka.DB) {
	var instopt InstInfo
	if err := c.ShouldBindJSON(&instopt); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	workersLock.Lock()
	defer workersLock.Unlock()
	instid := instopt.InstId
	appcode, _ := extractChar(instid)
	isSupport, msg := appCheck(appcode)
	//如果appcode=="opcda"时，ostype!="Windows"返回错误
	if !isSupport {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": appcode + " not support",
			"details": fmt.Sprintf(msg),
		})
		return
	}

	//如果应用实例是南向应用且中没有任何便签点
	//hasTag, msg := appHasTag(appcode, cfgdb)
	//if !hasTag {
	//	c.JSON(http.StatusBadRequest, gin.H{
	//		"message": instid + " no tag",
	//		"details": fmt.Sprintf(msg),
	//	})
	//	return
	//}
	now := time.Now()
	formattedDate := now.Format("2006-01-02 15:04:05")
	// 检查 funcMap 中是否存在对应的函数
	fn, exists := IotappMap[appcode]
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Function not found",
			"details": fmt.Sprintf("appCode '%s' has no associated function", appcode),
		})
		return
	}
	// 检查子线程是否已经在运行
	if _, cexists := Workers[instid]; cexists {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Worker " + instid + " is running",
			"data":    instopt,
		})
		return
	}
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
			// 通知全局变量 Workers 删除对应的线程 ID
			workersLock.Lock()
			delete(Workers, instid)
			workersLock.Unlock()
			fmt.Printf("StartApp提示：Worker退出,线程ID: %s 已从全局变量中删除\n", instid)
		}()
		fn(instid, stopChan, cfgdb, rtdb) // 调用对应的函数
	}()
	// 将子线程的停止通道存储到全局变量中
	Workers[instid] = stopChan
	// 返回子线程 ID
	fmt.Printf("%v Worker %v started\n", formattedDate, instid)
	c.JSON(http.StatusOK, gin.H{
		"message": "Worker started",
		"data":    instopt,
	})
}

// @Summary 通过实例ID停止App实例
// @Description 这是一个通过实例ID停止App实例的接口
// @Tags APP Manager
// @Accept json
// @Produce json
// @Param instid body InstInfo true "InstId"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Router /api/v1/stopApp [post]
func StopApp(c *gin.Context) {
	var instopt InstInfo
	if err := c.ShouldBindJSON(&instopt); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	instid := instopt.InstId

	// 查找子线程的停止通道
	stopChan, exists := Workers[instid]
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{
			"message": "Worker not found",
		})
		return
	}
	// 发送停止信号
	close(stopChan)
	// 从全局变量中移除子线程
	delete(Workers, instid)
	// 等待子线程退出
	//wg.Wait()

	// 检查全局变量 Workers 是否已删除对应的线程 ID
	workersLock.Lock()
	if _, cexists := Workers[instid]; !cexists {
		fmt.Printf("StopApp提示：Worker %s 已成功从全局变量中删除\n", instid)
	}
	workersLock.Unlock()
	// 返回成功消息
	c.JSON(http.StatusOK, gin.H{
		"message": "Worker stopped",
		"data":    instopt,
	})
}
