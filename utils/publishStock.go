package utils

import (
	"encoding/json"
	"flashSaleSystem/db/initDB"
)

func PublishStock() {
	result := make(map[string]string)

	var cursor uint64
	for {
		var keys []string
		var err error
		keys, cursor, err = initDB.Rdb.Scan(cursor, "*", 10).Result()
		if err != nil {
			return
		}

		for _, key := range keys {
			value, err := initDB.Rdb.Get(key).Result()
			if err != nil {
				return
			}
			result[key] = value
		}

		if cursor == 0 {
			break
		}
	}
	// 将库存信息转换为 JSON
	stockInfoJson, _ := json.Marshal(result)
	initDB.Rdb.Publish("stock_updates", stockInfoJson)

}
