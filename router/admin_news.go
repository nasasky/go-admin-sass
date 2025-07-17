package router

import (
	"nasa-go-admin/controllers/admin"

	"github.com/gin-gonic/gin"
)

func RegisterNewsRoutes(rg *gin.RouterGroup) {
	rg.POST("/news/add", admin.AddNews)
	rg.PUT("/news/update", admin.UpdateNews)
	rg.GET("/news/list", admin.GetNewsList)
	rg.GET("/news/detail", admin.GetNewsDetail)
	rg.DELETE("/news/delete", admin.DeleteNews)
}
