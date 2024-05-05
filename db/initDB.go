package db

import (
	"flashSaleSystem/config"
	"flashSaleSystem/db/model"
	"fmt"
	"github.com/go-redis/redis"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"github.com/streadway/amqp"
	"os"
)

var (
	Mdb *gorm.DB
	Rdb *redis.Client
	Ch  *amqp.Channel
)

func init() {
	// 初始化数据库连接
	Mdb, _ = initDB()
	// 初始化Redis连接
	Rdb, _ = initRedis()
	// 初始化RabbitMQ连接
	Ch, _ = initRabbitMQ()

	syncGoodsToRedis() // 将mysql中的库存数据同步进redis中
}

func initDB() (*gorm.DB, error) {
	// 连接MySQL数据库
	db, err := gorm.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:%s)/?charset=utf8&parseTime=True&loc=Local",
		config.ServerConfig.MYSQL.ACCOUNT,
		config.ServerConfig.MYSQL.PASSWORD,
		config.ServerConfig.MYSQL.HOST,
		config.ServerConfig.MYSQL.PORT,
	))
	if err != nil {
		fmt.Println("connect mysql failed, err:", err)
		os.Exit(-2)
	}
	// 创建数据库
	db.Exec(fmt.Sprintf("CREATE DATABASE IF NOT EXISTS %s", config.ServerConfig.MYSQL.DBNAME))
	// 关闭连接
	err = db.Close()
	if err != nil {
		fmt.Println("can not close the database")
		os.Exit(-3)
	}
	// 连接到指定数据库
	db, err = gorm.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8&parseTime=True&loc=Local",
		config.ServerConfig.MYSQL.ACCOUNT,
		config.ServerConfig.MYSQL.PASSWORD,
		config.ServerConfig.MYSQL.HOST,
		config.ServerConfig.MYSQL.PORT,
		config.ServerConfig.MYSQL.DBNAME,
	))
	if err != nil {
		fmt.Println("connect mysql failed, err:", err)
		os.Exit(-4)
	}
	fmt.Println("connect mysql successfully")

	db.AutoMigrate(model.Order{})
	db.AutoMigrate(model.Goods{})
	return db, nil
}

func initRedis() (*redis.Client, error) {
	// 连接Redis
	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", config.ServerConfig.REDIS.HOST, config.ServerConfig.REDIS.PORT),
		Password: config.ServerConfig.REDIS.PASSWORD,
		DB:       config.ServerConfig.REDIS.DBNAME, // 默认数据库
		PoolSize: 20000,                            // 连接池大小
	})

	_, err := rdb.Ping().Result()
	if err != nil {
		fmt.Println("can not connect redis: " + err.Error())
		os.Exit(-3)
	}
	fmt.Println("connect redis successfully")

	return rdb, nil
}

func initRabbitMQ() (*amqp.Channel, error) {
	// 连接到 RabbitMQ 服务器
	conn, err := amqp.Dial(fmt.Sprintf("amqp://%s:%s@%s:%s/",
		config.ServerConfig.RABBITMQ.ACCOUNT,
		config.ServerConfig.RABBITMQ.PASSWORD,
		config.ServerConfig.RABBITMQ.HOST,
		config.ServerConfig.RABBITMQ.PORT,
	))
	if err != nil {
		fmt.Println("can not connect RabbitMQ")
		os.Exit(-30)
	}
	fmt.Println("connect RabbitMQ successfully")

	ch, err := conn.Channel()
	if err != nil {
		_ = fmt.Errorf("can not connect RabbitMQ chanel")
		os.Exit(-5)
	}

	_, err = ch.QueueDeclare(
		"order_queue", // name
		true,          // durable
		false,         // delete when unused
		false,         // exclusive
		false,         // no-wait
		nil,           // arguments
	)
	if err != nil {
		fmt.Println("can not connect RabbitMQ chanel")
		os.Exit(-6)
	}

	return ch, nil
}

func syncGoodsToRedis() {
	var goodsRow []model.Goods
	err := Mdb.Model(model.Goods{}).Select("*").Find(&goodsRow).Error
	if err != nil {
		fmt.Println("con not sync goods from mysql to redis")
		os.Exit(-5)
	}
	// 将数据写入到 Redis
	for _, good := range goodsRow {
		err = Rdb.Set(good.Gid, good.Quantity, 0).Err()
		if err != nil {
			fmt.Println("con not sync goods from mysql to redis")
			os.Exit(-6)
		}
	}

}
