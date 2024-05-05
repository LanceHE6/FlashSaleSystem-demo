package main

import (
	"flashSaleSystem/config"
	"flashSaleSystem/handler"
	"flashSaleSystem/worker"
	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()
	// 启动后台数据库同步进程
	go worker.SyncOrderWorker()
	go worker.SyncGoods()

	if config.ServerConfig.SERVER.MODE == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	r.POST("/seckill", func(c *gin.Context) {
		handler.OrderHandler(c)
	})

	err := r.Run()
	if err != nil {
		return
	}

}
