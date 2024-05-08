package goods

import (
	"flashSaleSystem/db/initDB"
	"flashSaleSystem/db/model"
	"flashSaleSystem/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

type AddGoodsRequest struct {
	Name     string `json:"name" binding:"required"`
	Quantity int    `json:"quantity" binding:"required"`
}

// AddGoods 添加货品
func AddGoods(context *gin.Context) {
	var data AddGoodsRequest
	if err := context.ShouldBindJSON(&data); err != nil {
		context.JSON(http.StatusBadRequest, utils.ErrorResponse("Incomplete request parameters", err.Error(), 400))
		return
	}

	gid := "g" + utils.GenerateUuid(8)
	goods := model.Goods{
		Gid:      gid,
		Name:     data.Name,
		Quantity: data.Quantity,
	}

	db := initDB.Mdb
	result := db.Model(model.Goods{}).Create(&goods)
	if result.Error != nil {
		context.JSON(http.StatusInternalServerError, utils.ErrorResponse("can not insert new goods", result.Error.Error(), 500))
		return
	}
	// 同步变化进redis
	initDB.SyncGoodsToRedis()
	utils.PublishStock()
	context.JSON(http.StatusOK, utils.Response("insert new goods successfuly", gin.H{}, 200))
}
