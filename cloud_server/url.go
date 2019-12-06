package server

import (
	"MyCloud/cloud_server/api"
	"github.com/gin-gonic/gin"
)

func UrlMap(router *gin.Engine) {
	router.POST("/sign", api.Sign)
}
