package goods

import (
	"flashSaleSystem/db/initDB"
	"flashSaleSystem/db/model"
	"flashSaleSystem/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

type deleteRequest struct {
	GoodsId string `json:"goods_id" binding:"required"`
}

func DeleteGoods(context *gin.Context) {
	var data deleteRequest
	if err := context.ShouldBindJSON(&data); err != nil {
		context.JSON(http.StatusBadRequest, utils.ErrorResponse("Incomplete request parameters", err.Error(), 400))
		return
	}

	err := initDB.Mdb.Model(&model.Goods{}).Delete(model.Goods{Gid: data.GoodsId}).Error
	if err != nil {
		context.JSON(http.StatusInternalServerError, utils.ErrorResponse("can not delete goods", err.Error(), 500))
		return
	}
	// 同步变化进redis
	initDB.SyncGoodsToRedis()
	utils.PublishStock()
	context.JSON(http.StatusOK, utils.Response("delete goods successfuly", gin.H{}, 200))
}
