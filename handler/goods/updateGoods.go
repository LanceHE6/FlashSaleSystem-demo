package goods

import (
	"flashSaleSystem/db/initDB"
	"flashSaleSystem/db/model"
	"flashSaleSystem/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

type updateGoodsRequest struct {
	Gid      string `json:"gid" binding:"required"`
	Name     string `json:"name"`
	Quantity int    `json:"quantity"`
}

// AddGoods 添加货品
func UpdateGoods(context *gin.Context) {
	var data updateGoodsRequest
	if err := context.ShouldBindJSON(&data); err != nil {
		context.JSON(http.StatusBadRequest, utils.ErrorResponse("Incomplete request parameters", err.Error(), 400))
		return
	}

	if data.Name == "" && data.Quantity == 0 {
		context.JSON(http.StatusBadRequest, utils.Response("Provide at least one of two parameters", gin.H{}, 401))
		return
	}

	goods := model.Goods{
		Gid:      data.Gid,
		Name:     data.Name,
		Quantity: data.Quantity,
	}

	db := initDB.Mdb
	result := db.Model(model.Goods{}).Where("gid=?", goods.Gid).Updates(&goods)
	if result.Error != nil {
		context.JSON(http.StatusInternalServerError, utils.ErrorResponse("can not update goods", result.Error.Error(), 500))
		return
	}
	// 同步变化进redis
	initDB.SyncGoodsToRedis()
	utils.PublishStock()
	context.JSON(http.StatusOK, utils.Response("update goods successfuly", gin.H{}, 200))
}
