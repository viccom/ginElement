package main

import (
	"fmt"
	_ "ginElement/docs"
	"ginElement/handlers"
	"ginElement/routes"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
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

	//rtdb, err := redka.Open("data/rt.db", &opts)
	// All data is lost when the database is closed.
	//rtdb, err := redka.Open("file:/rt.db?vfs=memdb", nil)
	rtdb, err := redka.Open("file::memory:?cache=shared", nil)
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
	//autostartWorker(handlers.Simtodb, rtdb, "simtodb")
	// 启动服务
	fmt.Println("Server is running on :8880...")
	err = r.Run(":8880")
	if err != nil {
		return
	}
}

func autostartWorker(workerFunc func(string, chan struct{}, *redka.DB), rtdb *redka.DB, workerName string) {
	workersLock := &sync.Mutex{}
	workersLock.Lock()
	defer workersLock.Unlock()
	newUUID := uuid.New()
	uuidstr := workerName + "@" + newUUID.String()
	stopChan := make(chan struct{})
	go func() {
		defer func() {
			select {
			case <-stopChan:
			default:
				close(stopChan)
			}
		}()
		workerFunc(uuidstr, stopChan, rtdb)
	}()

	handlers.Workers[uuidstr] = stopChan
}
