package main

import (
	"fmt"
	_ "ginElement/docs"
	"ginElement/routes"
	"github.com/gin-gonic/gin"
	"github.com/nalgeon/redka"
	swaggerFiles "github.com/swaggo/files"     // 用于提供 Swagger UI 静态文件
	ginSwagger "github.com/swaggo/gin-swagger" // 用于集成 Swagger UI 到 Gin
	"log"
)

func main() {
	// 初始化数据库连接
	cfgdb, err := redka.Open("config.db", nil)
	if err != nil {
		log.Fatal(err)
	}
	defer cfgdb.Close()

	rtdb, err := redka.Open("rt.db", nil)
	if err != nil {
		log.Fatal(err)
	}
	defer rtdb.Close()

	// 创建 Gin 引擎
	r := gin.Default()

	// 调用 routes.SetupRouter，传递数据库连接
	routes.SetupRouter(r, cfgdb, rtdb)

	// 提供 Swagger UI 静态文件
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// 启动服务
	fmt.Println("Server is running on :8880...")
	err = r.Run(":8880")
	if err != nil {
		return
	}
}
