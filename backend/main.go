package main

import (
	"encoding/json"
	"flag"
	"fmt"
	_ "ginElement/docs"
	"ginElement/handlers"
	"ginElement/routes"
	"github.com/gin-gonic/gin"
	"github.com/nalgeon/redka"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	swaggerFiles "github.com/swaggo/files"     // 用于提供 Swagger UI 静态文件
	ginSwagger "github.com/swaggo/gin-swagger" // 用于集成 Swagger UI 到 Gin
	"log"
	"sync"
)

func main() {
	var (
		startWeb = flag.String("startweb", "0", "startWeb mode: 1, 0, Default: 1")
		port     = flag.String("port", "8880", "listen port, Default: 8880")
	)
	//flag.BoolVar(&debug.Enable, "debug", false, "enable debug logging")
	flag.Parse()

	// 获取当前可执行文件的路径
	exePath, err := os.Executable()
	if err != nil {
		fmt.Println("Failed to get executable path:", err)
		return
	}

	// 获取可执行文件所在的目录
	exeDir := filepath.Dir(exePath)

	// 切换工作路径到可执行文件所在目录
	err = os.Chdir(exeDir)
	if err != nil {
		fmt.Println("Failed to change working directory:", err)
		return
	}

	errx := handlers.EnsureDirExists("data")
	if errx != nil {
		fmt.Printf("Error: %v\n", errx)
		return
	}
	//err = handlers.CheckDBAndDelete("data/rt.db")
	// 初始化数据库连接
	opts := redka.Options{
		DriverName: "sqlite",
	}
	cfgdb, err := redka.Open("data/config.db", &opts)
	if err != nil {
		log.Fatal(err)
	}
	defer cfgdb.Close()

	//rtopts := redka.Options{
	//	DriverName: "sqlite3",
	//	Pragma: map[string]string{
	//		"temp_store": "memory",
	//	},
	//}

	//rtdb, err := redka.Open("data/rt.db", &opts)
	// All data is lost when the database is closed.
	//rtdb, err := redka.Open("file:/rt.db?vfs=memdb", nil)
	// All data is lost when the database is closed.
	//rtdb, err := redka.Open("file::memory:?cache=shared", nil)
	rtdb, err := redka.Open("file:redka?mode=memory&cache=shared", &opts)
	if err != nil {
		log.Fatal(err)
	}
	defer rtdb.Close()

	_, err2 := cfgdb.Hash().Set("system@router", "version", handlers.AppVersion)
	if err2 != nil {
		log.Println("write config.db err:", err2)
		return
	}
	rid, _ := cfgdb.Hash().Get("system@router", "routerid")
	newrid, _ := handlers.GetHardwareID()
	if rid.String() == "" || rid.String() != newrid {
		_, err3 := cfgdb.Hash().Set("system@router", "routerid", newrid)
		if err3 != nil {
			log.Println("write config.db err:", err2)
			return
		}
	}

	// 创建 Gin 引擎
	r := gin.Default()

	// 设置静态文件服务，将静态文件目录映射到URL路径
	r.Static("/html", "./html")
	r.StaticFile("/favicon.svg", "./html/favicon.svg")
	r.StaticFile("/vite.svg", "./html/vite.svg")
	r.StaticFile("/element-plus-logo-small.svg", "./html/element-plus-logo-small.svg")
	r.Static("/assets", "./html/assets")
	r.Static("/md", "./html/md")

	// 处理根URL请求，返回index.html
	r.GET("/", func(c *gin.Context) {
		c.File("./html/index.html")
	})
	// 调用 routes.SetupRouter，传递数据库连接
	routes.SetupRouter(r, cfgdb, rtdb)

	// 提供 Swagger UI 静态文件
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// 启动配置中设置为程序运行自启动的实例
	items, err := cfgdb.Hash().Items(handlers.InstListKey)
	if err != nil {
		return
	}
	for _, item := range items {
		var appconfig handlers.AppConfig
		erra := json.Unmarshal([]byte(item.String()), &appconfig)
		if erra != nil {
			fmt.Println("Error unmarshalling JSON:", erra)
			return
		}
		//fmt.Printf("hashkey: %v, InstId: %v, AppCode: %v, AutoStart: %v\n", key, appconfig.InstID, appconfig.AppCode, appconfig.AutoStart)
		if appconfig.AutoStart == true {
			startWorker(handlers.IotappMap[appconfig.AppCode], cfgdb, rtdb, appconfig.InstID)
		}
	}
	startweb := *startWeb
	if startweb == "1" {
		url := "http://localhost:" + *port
		erra := openBrowser(url)
		if erra != nil {
			fmt.Printf("Failed to open browser: %s\n", erra)
		}
	}

	// 启动WEB服务
	fmt.Println("Server is running on :8880...")
	err = r.Run(":" + *port)
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
			// 通知全局变量 Workers 删除对应的线程 ID
			workersLock.Lock()
			delete(handlers.Workers, workerName)
			workersLock.Unlock()
			fmt.Printf("子线程 %s 已从全局变量中删除\n", workerName)
		}()
		workerFunc(workerName, stopChan, cfgdb, rtdb)
	}()
	//workersLock.Lock()
	handlers.Workers[workerName] = stopChan
	//workersLock.Unlock()
}

func openBrowser(url string) error {
	var cmd string
	var args []string

	switch runtime.GOOS {
	case "windows":
		cmd = "cmd"
		args = []string{"/c", "start", url}
	case "darwin":
		cmd = "open"
		args = []string{url}
	default: // "linux", "freebsd", "openbsd", "netbsd"
		cmd = "xdg-open"
		args = []string{url}
	}

	return exec.Command(cmd, args...).Start()
}
