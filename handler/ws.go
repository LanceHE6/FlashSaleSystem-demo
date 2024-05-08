package handler

import (
	"flashSaleSystem/db/initDB"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

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

	pubsub := initDB.Rdb.Subscribe("stock_updates")
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
