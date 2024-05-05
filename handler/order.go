package handler

import (
	"encoding/json"
	"flashSaleSystem/db"
	"flashSaleSystem/utils"
	"github.com/gin-gonic/gin"
	"github.com/streadway/amqp"
	"net/http"
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
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	userId := data.UserId
	Gid := data.GoodsId
	orderTime := data.OrderTime
	orderQuantity := data.OrderQuantity
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
		cmd := db.Rdb.Eval(luaScript, []string{Gid}, orderQuantity)
		stock, err := cmd.Result()

		if err != nil {
			resultChan <- gin.H{"error": err.Error()}
			return
		}
		if stock, ok := stock.(int64); ok {
			if stock == -1 {
				resultChan <- gin.H{
					"message": "Not enough stock",
				}
				return
			}
		} else {
			resultChan <- gin.H{
				"error": "Unexpected result type from Lua script",
			}
			return
		}

		orderId := utils.GenerateUuid(8) // 生成一个唯一的订单ID
		//// 在Redis中存储用户的订单信息
		//err = db.Rdb.HMSet(orderId, map[string]interface{}{
		//	"userId":        userId,
		//	"orderTime":     orderTime,
		//	"orderQuantity": orderQuantity,
		//}).Err()
		//if err != nil {
		//	resultChan <- gin.H{"error": err.Error()}
		//	return
		//}

		// 将订单信息发送到RabbitMQ队列
		order := map[string]interface{}{
			"orderId":       orderId,
			"userId":        userId,
			"orderTime":     orderTime,
			"orderQuantity": orderQuantity,
		}
		body, err := json.Marshal(order)
		if err != nil {
			resultChan <- gin.H{"error": err.Error()}
			return
		}
		err = db.Ch.Publish(
			"",            // exchange
			"order_queue", // routing key
			false,         // mandatory
			false,         // immediate
			amqp.Publishing{
				ContentType: "application/json",
				Body:        body,
			})
		if err != nil {
			resultChan <- gin.H{"error": err.Error()}
			return
		}

		resultChan <- gin.H{
			"userId":        userId,
			"orderTime":     orderTime,
			"orderQuantity": orderQuantity,
			"stock":         stock,
		}
	}()

	result := <-resultChan
	err, ok := result["error"]

	if ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}
	msg, ok := result["message"]
	if ok {
		c.JSON(http.StatusOK, gin.H{"msg": msg})
		return
	} else {
		c.JSON(http.StatusOK, result)
	}

}
