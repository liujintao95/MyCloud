package server

import (
	"MyCloud/cloud_server/api"
	"github.com/gin-gonic/gin"
)

func UrlMap(router *gin.Engine) {
	router.POST("/sign", api.Sign)
	router.POST("/register", api.Register)
	router.GET("/logout", api.Logout)
	router.GET("/passwordchange", api.PasswordChange)
	router.POST("/file/upload", api.Upload)
	router.GET("/file/download", api.Download)
	router.GET("/file/update", api.Update)
	router.GET("/file/delete", api.Delete)
}
