package main

import (
	"flashSaleSystem/config"
	"flashSaleSystem/handler"
	"flashSaleSystem/handler/goods"
	"flashSaleSystem/worker"

	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()

	// 解决跨域问题
	r.Use(cors.New(cors.Config{
		//准许跨域请求网站,多个使用,分开,限制使用*
		AllowOrigins: []string{"*"},
		//准许使用的请求方式
		AllowMethods: []string{"PUT", "PATCH", "POST", "GET", "DELETE"},
		//准许使用的请求表头
		AllowHeaders: []string{"Origin", "Authorization", "Content-Type", "Access-Token"},
		//显示的请求表头
		ExposeHeaders: []string{"Content-Type"},
		//凭证共享,确定共享
		AllowCredentials: true,
		//容许跨域的原点网站,可以直接return true就万事大吉了
		AllowOriginFunc: func(origin string) bool {
			return true
		},
		//超时时间设定
		MaxAge: 24 * time.Hour,
	}))

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
	api.GET("/ws", func(ctx *gin.Context) {
		handler.WsHandler(ctx)
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
	goodsApi.POST("/del", func(ctx *gin.Context) {
		goods.DeleteGoods(ctx)
	})

	err := r.Run()
	if err != nil {
		return
	}

}
