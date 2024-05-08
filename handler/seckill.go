package handler

import (
	"encoding/json"
	"flashSaleSystem/db/initDB"
	"flashSaleSystem/utils"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/streadway/amqp"
)

type orderRequest struct {
	UserId        string `json:"user_id" binding:"required"`
	GoodsId       string `json:"goods_id" binding:"required"`
	OrderTime     string `json:"order_time" binding:"required"`
	OrderQuantity int    `json:"order_num" binding:"required"`
}

func OrderHandler(c *gin.Context) {
	// 创建一个用于控制并发的Channel
	concurrentLimit := make(chan bool, 10000) // 限制并发数为10000
	var data orderRequest
	if err := c.ShouldBindJSON(&data); err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse("Incomplete request parameters", err.Error(), 400))
		return
	}
	userId := data.UserId
	Gid := data.GoodsId
	orderQuantity := data.OrderQuantity
	orderTime := data.OrderTime
	// 创建一个Channel用于接收处理结果
	resultChan := make(chan gin.H)

	// 使用Goroutine处理秒杀请求
	go func() {
		concurrentLimit <- true
		defer func() { <-concurrentLimit }()
		// 使用Lua脚本保证原子性
		luaScript := `
				local stock = tonumber(redis.call('get', KEYS[1]))
				local quantity = tonumber(ARGV[1])
				if stock >= quantity then
				  redis.call('decrby', KEYS[1], quantity)
				  return stock - quantity
				else
				  return -1
				end
				`
		cmd := initDB.Rdb.Eval(luaScript, []string{Gid}, orderQuantity)
		stock, err := cmd.Result()

		if err != nil {
			resultChan <- utils.ErrorResponse("Script execution error", err.Error(), 500)
			return
		}
		if stock, ok := stock.(int64); ok {
			if stock == -1 {
				resultChan <- utils.Response("out of stock", gin.H{}, 200)
				return
			}
		} else {
			resultChan <- utils.ErrorResponse("Unexpected result type from Lua script", "Unexpected result type from Lua script", 501)
			return
		}

		orderId := utils.GenerateUuid(8) // 生成一个唯一的订单ID

		// 将订单信息发送到RabbitMQ队列
		order := map[string]interface{}{
			"orderId":       orderId,
			"userId":        userId,
			"orderQuantity": orderQuantity,
			"orderTime":     orderTime,
		}
		body, err := json.Marshal(order)
		if err != nil {
			resultChan <- utils.ErrorResponse("MQ error", err.Error(), 502)
			return
		}
		err = initDB.Ch.Publish(
			"",            // exchange
			"order_queue", // routing key
			false,         // mandatory
			false,         // immediate
			amqp.Publishing{
				ContentType: "application/json",
				Body:        body,
			})
		if err != nil {
			resultChan <- utils.ErrorResponse("MQ error", err.Error(), 503)
			return
		}

		resultChan <- utils.Response("send MQ successfully", gin.H{
			"userId":        userId,
			"orderQuantity": orderQuantity,
			"stock":         stock,
		}, 201)
	}()

	result := <-resultChan
	_, ok := result["error"]

	if ok {
		c.JSON(http.StatusInternalServerError, result)
		return
	} else {
		utils.PublishStock()
		c.JSON(http.StatusOK, result)
	}

}
