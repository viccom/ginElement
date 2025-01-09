package main

import (
	"encoding/json"
	"fmt"
	_ "ginElement/docs"
	"ginElement/handlers"
	"ginElement/routes"
	"github.com/gin-gonic/gin"
	"github.com/nalgeon/redka"

	swaggerFiles "github.com/swaggo/files"     // 用于提供 Swagger UI 静态文件
	ginSwagger "github.com/swaggo/gin-swagger" // 用于集成 Swagger UI 到 Gin
	"log"
	"sync"
)

func main() {

	err := handlers.EnsureDirExists("data")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	// 初始化数据库连接
	cfgdb, err := redka.Open("data/config.db", nil)
	if err != nil {
		log.Fatal(err)
	}
	defer cfgdb.Close()

	//opts := redka.Options{
	//	DriverName: "sqlite3",
	//	Pragma: map[string]string{
	//		"temp_store": "memory",
	//	},
	//}

	rtdb, err := redka.Open("data/rt.db", nil)
	// All data is lost when the database is closed.
	//rtdb, err := redka.Open("file:/rt.db?vfs=memdb", nil)
	// All data is lost when the database is closed.
	//rtdb, err := redka.Open("file::memory:?cache=shared", nil)
	//rtdb, err := redka.Open("file:redka?mode=memory&cache=shared", nil)
	if err != nil {
		log.Fatal(err)
	}
	defer rtdb.Close()

	_, err2 := cfgdb.Hash().Set("system@router", "version", handlers.AppVersion)
	if err2 != nil {
		log.Println("write config.db err:", err2)
		return
	}

	// 创建 Gin 引擎
	r := gin.Default()

	// 调用 routes.SetupRouter，传递数据库连接
	routes.SetupRouter(r, cfgdb, rtdb)

	// 提供 Swagger UI 静态文件
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	// 启动一个子线程运行 periodicPrint 函数
	items, err := cfgdb.Hash().Items(handlers.InstListKey)
	if err != nil {
		return
	}
	for key, item := range items {
		var appconfig handlers.AppConfig
		erra := json.Unmarshal([]byte(item.String()), &appconfig)
		if erra != nil {
			fmt.Println("Error unmarshalling JSON:", erra)
			return
		}
		fmt.Printf("hashkey: %v, InstId: %v, AppCode: %v, AutoStart: %v\n", key, appconfig.InstID, appconfig.AppCode, appconfig.AutoStart)
		if appconfig.AutoStart == true {
			startWorker(handlers.IotappMap[appconfig.AppCode], cfgdb, rtdb, appconfig.InstID)
		}
	}

	// 启动服务
	fmt.Println("Server is running on :8880...")
	err = r.Run(":8880")
	if err != nil {
		return
	}
}

func startWorker(workerFunc func(string, chan struct{}, *redka.DB, *redka.DB), cfgdb *redka.DB, rtdb *redka.DB, workerName string) {
	if workerFunc == nil {
		log.Printf("Error: workerFunc is nil for worker %s", workerName)
		return
	}
	workersLock := &sync.Mutex{}
	workersLock.Lock()
	defer workersLock.Unlock()
	//newUUID := uuid.New()
	//uuidstr := workerName + "@" + newUUID.String()
	stopChan := make(chan struct{})
	go func() {
		defer func() {
			select {
			case <-stopChan:
			default:
				close(stopChan)
			}
		}()
		workerFunc(workerName, stopChan, cfgdb, rtdb)
	}()

	handlers.Workers[workerName] = stopChan
}
