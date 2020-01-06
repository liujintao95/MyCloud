package server

import (
	"MyCloud/cloud_server/api"

	"github.com/gin-gonic/gin"
)

func UrlMap(router *gin.Engine) {
	router.POST("/sign", api.Sign)
	router.POST("/register", api.Register)
	router.GET("/logout", api.Logout)

	authorized := router.Group("/auth", LoginRequired)
	authorized.POST("/user/change/password", api.PasswordChange)
	authorized.POST("/user/change/username", api.UsernameChange)
	authorized.POST("/file/upload", api.Upload)
	authorized.POST("/file/rapid_upload", api.RapidUpload)
	authorized.GET("/file/download", api.Download)
	authorized.GET("/file/public_download", api.PublicDownload)
	authorized.POST("/file/update", api.UpdateFileName)
	authorized.POST("/file/delete", api.Delete)
	authorized.POST("/block/init", api.InitBlockUpload)
	authorized.GET("/block/resume", api.ResumeFromBreakPoint)
	authorized.POST("/block/upload", api.BlockUpload)
	authorized.GET("/block/progress", api.UploadProgress)
	authorized.POST("/block/merge", api.BlockMerge)
	authorized.POST("/block/remove", api.RemoveBlock)
}
