package worker

import (
	"encoding/json"
	"flashSaleSystem/db"
	"flashSaleSystem/db/model"
	"fmt"
	"log"
	"time"
)

func failOnError(err error, msg string) {
	if err != nil {
		//log.Fatalf("%s: %s", msg, err)
		panic(fmt.Sprintf("%s: %s", msg, err))
	}
}

func SyncOrderWorker() {

	q, err := db.Ch.QueueDeclare(
		"order_queue", // name
		true,          // durable
		false,         // delete when unused
		false,         // exclusive
		false,         // no-wait
		nil,           // arguments
	)
	failOnError(err, "Failed to declare a queue")

	msgs, err := db.Ch.Consume(
		q.Name, // queue
		"",     // consumer
		true,   // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	failOnError(err, "Failed to register a consumer")
	fmt.Println("started sync order RabbitMQ worker")
	for d := range msgs {
		log.Printf("Received a message: %s", d.Body)
		order := model.Order{}
		err := json.Unmarshal(d.Body, &order)
		if err != nil {
			log.Printf("Error decoding JSON: %s", err)
		}

		// 将订单信息写入到 MySQL 数据库
		db.Mdb.Create(&order)
	}
}

func SyncGoods() {
	fmt.Println("started sync goods worker")
	// 创建一个 Ticker，每隔10秒钟触发一次
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()
	// 无限循环，每次 Ticker 触发时同步一次 Redis 和 MySQL
	for {
		select {
		case <-ticker.C:
			err := syncGoodsToMySQL()
			if err != nil {
				log.Printf("Failed to sync Redis and MySQL: %v", err)
			}
		}
	}
}

func syncGoodsToMySQL() error {
	// 获取 Redis 中所有的商品库存
	keys, err := db.Rdb.Keys("g*").Result()
	if err != nil {
		return err
	}

	// 逐个更新 MySQL 中的商品库存
	for _, key := range keys {
		stock, err := db.Rdb.Get(key).Int()
		if err != nil {
			return err
		}

		var good model.Goods
		good.Quantity = stock
		db.Mdb.Model(&model.Goods{}).Where("gid=?", key).Updates(&good)
		good.Quantity = stock
		fmt.Println("sync goods to mysql")
	}

	return nil
}
