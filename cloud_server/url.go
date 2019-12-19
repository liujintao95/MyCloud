package server

import (
	"MyCloud/cloud_server/api"
	"github.com/gin-gonic/gin"
)

func UrlMap(router *gin.Engine) {
	router.POST("/sign", api.Sign)
	router.POST("/register", api.Register)
	router.GET("/logout", api.Logout)

	authorized := router.Group("/auth", loginRequired)
	authorized.GET("/user/change/password", api.PasswordChange)
	authorized.GET("/user/change/username", api.UsernameChange)
	authorized.POST("/file/upload", api.Upload)
	authorized.GET("/file/download", api.Download)
	authorized.GET("/file/update", api.UpdateFileName)
	authorized.GET("/file/delete", api.Delete)
}
