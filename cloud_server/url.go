package server

import (
	"MyCloud/cloud_server/api"
	"github.com/gin-gonic/gin"
)

func UrlMap(router *gin.Engine) {
	router.POST("/sign", api.Sign)
	router.POST("/register", api.Register)
	router.POST("/logout", api.Logout)
}
