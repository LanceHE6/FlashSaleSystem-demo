<p align="center">
	<img alt="logo" src="./imgs/buy.256x256.png">
</p>
<h1 align="center" style="margin: 30px 0 30px; font-weight: bold;">FlashSaleSystem-Demo</h1>
<h4 align="center">基于gin+Vue3前后端分离的秒杀演示系统</h4>

<div align="center">

![Static Badge](https://img.shields.io/badge/Licence-MIT-blue)
![Static Badge](https://img.shields.io/badge/前端-vue-orange)
![Static Badge](https://img.shields.io/badge/后端-gin-green)
![Static Badge](https://img.shields.io/badge/Database-MySQL-red)
![Static Badge](https://img.shields.io/badge/Cache-Redis-yellow)
![Static Badge](https://img.shields.io/badge/MQ-RabbitMQ-purple)

</div>






## 技术栈介绍

* 前端技术栈 [Vue3](https://v3.cn.vuejs.org) + [Naive UI](https://www.naiveui.com/zh-CN/os-theme) + [Vite](https://cn.vitejs.dev) 。
* 后端[gin](https://gin-gonic.com/zh-cn/)+[gorm](https://gorm.io/zh_CN/docs/index.html)。
* 数据库[mysql]([MySQL](https://www.mysql.com/cn/))+[redis](https://redis.io/)。
* 消息队列[RabbitMQ](https://www.rabbitmq.com/)

## 前端运行

```bash
# 克隆项目
git clone https://github.com/LanceHE6/FlashSaleSystem-demo.git

# 进入项目目录
cd vue-web

# 安装依赖
npm install

# 启动服务
npm run dev

# 前端访问地址 http://localhost:5173
```

## 后端运行

```bash
# 进入项目目录
cd go-server

# 下载依赖
go mod download

# 直接运行
go run main.go

# 构建项目
go build -o main .
```



## 内置功能

* 商品添加
* 商品修改
* 商品更新
* 商品删除
* 商品秒杀
* 库存较实时显示

## 技术点介绍

### 高并发实现

```go
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

```



1. **并发限制**：通过创建一个带缓存的通道 `concurrentLimit`，并将其容量设置为 10000，函数限制了同时处理的并发请求的数量。每次处理请求时，函数都会尝试向这个通道发送一个值，如果通道已满（即正在处理的请求已达到 10000），这个发送操作就会阻塞，从而阻止处理更多的并发请求。当一个请求处理完成时，函数会从通道接收一个值，从而为其他请求腾出空间。
2. **异步处理**：使用 Go 协程来异步处理每个请求。这意味着函数不会等待一个请求处理完成就开始处理下一个请求，而是将每个请求的处理操作放在一个单独的协程中，并立即开始处理下一个请求。这样可以大大提高函数处理并发请求的能力。
3. **原子操作**：使用 Lua 脚本在 Redis 中执行库存检查和更新操作。由于 Redis 可以原子地执行 Lua 脚本，这可以防止在高并发环境下出现数据竞态条件。
4. **消息队列**：使用 RabbitMQ 消息队列来处理订单信息。当一个请求需要处理的订单信息被生成后，将这个信息发送到 RabbitMQ 队列，然后立即开始处理下一个请求。这样可以避免函数在处理请求时等待订单信息的处理操作完成。
5. **实时更新库存信息**：在一个请求被处理完成后，调用 `utils.PublishStock()` 函数来更新库存信息。这样可以确保在高并发环境下，库存信息始终是最新的。

### 库存实时显示

```go
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func WsHandler(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		// handle error
		fmt.Println("upgrade protocol failed")
		return
	}
	defer conn.Close()

	pubsub := initDB.Rdb.Subscribe("stock_updates") // initDB.Rdb.Publish("stock_updates", stockInfoJson) 		file:utils.publishStock.go
	defer pubsub.Close()

	// 使用一个 goroutine 来处理来自 Redis 的消息
	go func() {
		for {
			msg, err := pubsub.ReceiveMessage()
			if err != nil {
				// handle error
				fmt.Println(err.Error())
				break
			}

			// 将消息发送给前端
			if err := conn.WriteMessage(websocket.TextMessage, []byte(msg.Payload)); err != nil {
				// handle error
				fmt.Println(err.Error())
				break
			}
		}
	}()

	// 使用一个无限循环来保持 WebSocket 连接打开
	for {
		// 等待 1 秒
		time.Sleep(1 * time.Second)
	}
}
```

```vue
// 创建 WebSocket 连接
const socket = new WebSocket('ws://127.0.0.1:8080/api/ws')

// WebSocket 连接打开时触发
socket.onopen = function(event) {
	console.log('WebSocket is open now.')
}

// WebSocket 连接关闭时触发
socket.onclose = function(event) {
	console.log('WebSocket is closed now.')
}

// WebSocket 接收到消息时触发
socket.onmessage = function(event) {
	let stockInfo = JSON.parse(event.data)
// 更新商品的库存信息
	for (let item of items.value) {
		if (stockInfo.hasOwnProperty(item.gid)) {
			item.quantity = stockInfo[item.gid]
		}
	}
}

// WebSocket 出错时触发
socket.onerror = function(error) {
	console.log(`WebSocket error: ${error}`)
}
```



1. **使用 WebSocket 连接**：首先使用 `upgrader.Upgrade` 方法将 HTTP 连接升级为 WebSocket 连接。WebSocket 是一种在单个 TCP 连接上进行全双工通信的协议，它允许服务器主动向客户端发送消息。使用一个无限循环来保持 WebSocket 连接打开。每次循环都会等待 1 秒，然后继续下一次循环。这样可以防止函数在处理完所有当前的消息后立即返回，从而关闭 WebSocket 连接。
2. **订阅 Redis 频道**：使用 `initDB.Rdb.Subscribe` 方法订阅了名为 "stock_updates" 的 Redis 频道。Redis 的发布-订阅模型允许客户端订阅一个或多个频道，并接收其他客户端向这些频道发布的消息。
3. **处理 Redis 消息**：创建一个 Go 协程来处理从 Redis 频道接收的消息。每当接收到一个消息时，函数就会将这个消息的内容作为 WebSocket 消息发送给客户端。这样，每当 "stock_updates" 频道收到一个新的库存更新消息时，这个消息就会被实时地发送给客户端。
