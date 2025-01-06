package routes

import (
	"ginElement/handlers" // 导入处理函数包
	"github.com/gin-gonic/gin"
	"github.com/nalgeon/redka"
)

func SetupRouter(r *gin.Engine, cfgdb *redka.DB, rtdb *redka.DB) {
	// 注册路由

	// 线程管理
	r.POST("/api/v1/startWorker/:appcode", func(c *gin.Context) {
		// 将数据库连接传递给 handlers.StartWorker
		handlers.StartWorker(c)
	}) // 启动子线程
	r.POST("/api/v1/stopWorker/:workerid", handlers.StopWorker) // 停止子线程
	r.GET("/api/v1/listWorkers", handlers.ListWorkers)          // 查询运行的子线程

	// App管理...
	// 查询程序支持的Appcode
	r.GET("/api/v1/listAppcode", handlers.ListAppcode)
	// 查询程序已经注册的App
	r.GET("/api/v1/listApps", func(c *gin.Context) {
		// 将数据库连接传递给 handlers.ListApps
		handlers.ListApps(c, cfgdb)
	})
	// 查询指定App实例的信息
	r.GET("/api/v1/getApp", func(c *gin.Context) {
		// 将数据库连接传递给 handlers.GetApp
		handlers.GetApp(c, cfgdb)
	})
	r.POST("/api/v1/newApp", func(c *gin.Context) {
		// 将数据库连接传递给 handlers.NewApp
		handlers.NewApp(c, cfgdb)
	}) // 新建App实例
	// 删除App实例
	r.POST("/api/v1/delApp", func(c *gin.Context) {
		// 将数据库连接传递给 handlers.NewApp
		handlers.DelApp(c, cfgdb)
	})
	// 修改App实例
	r.POST("/api/v1/modApp", func(c *gin.Context) {
		// 将数据库连接传递给 handlers.NewApp
		handlers.ModApp(c, cfgdb)
	})
	// 启动App实例
	r.POST("/api/v1/startApp/:appcode", func(c *gin.Context) {
		handlers.StartApp(c, rtdb)
	})
	// 停止App实例
	r.POST("/api/v1/stopApp", func(c *gin.Context) {
		handlers.StopApp(c)
	})

	// 设备管理
	// 查询设备列表
	//r.GET("/api/v1/listDevices", func(c *gin.Context) {
	//	// 将数据库连接传递给 handlers.ListDevices
	//	handlers.ListDevices(c, cfgdb)
	//})

	//数据管理

	// 日志管理

}
