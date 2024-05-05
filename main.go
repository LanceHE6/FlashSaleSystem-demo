package main

import (
	"flashSaleSystem/config"
	"flashSaleSystem/handler"
	"flashSaleSystem/handler/goods"
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

	api := r.Group("/api")

	// 秒杀接口
	api.POST("/seckill", func(ctx *gin.Context) {
		handler.OrderHandler(ctx)
	})

	// 货品相关接口
	goodsApi := api.Group("/goods")
	goodsApi.POST("/add", func(ctx *gin.Context) {
		goods.AddGoods(ctx)
	})
	goodsApi.GET("/list", func(ctx *gin.Context) {
		goods.ListGoods(ctx)
	})
	goodsApi.PUT("/update", func(ctx *gin.Context) {
		goods.UpdateGoods(ctx)
	})

	err := r.Run()
	if err != nil {
		return
	}

}
