package goods

import (
	"flashSaleSystem/db/initDB"
	"flashSaleSystem/db/model"
	"flashSaleSystem/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

// ListGoods 获取商品信息列表
func ListGoods(context *gin.Context) {
	db := initDB.Mdb

	goods := []model.Goods{}

	err := db.Select([]string{"gid", "name", "quantity", "created_at", "updated_at"}).Find(&goods).Error
	if err != nil {
		context.JSON(http.StatusInternalServerError, utils.ErrorResponse("can not select goods data", err.Error(), 400))
		return
	}

	rows := []gin.H{}

	for _, g := range goods {
		rows = append(rows, gin.H{
			"gid":        g.Gid,
			"name":       g.Name,
			"quantity":   g.Quantity,
			"created_at": g.CreatedAt,
			"updated_at": g.UpdatedAt,
		})
	}
	context.JSON(http.StatusOK, utils.Response("get goods list successfully", gin.H{"rows": rows}, 200))
}
