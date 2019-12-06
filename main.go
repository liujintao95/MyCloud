package main

import (
	"MyCloud/cloud_server"
	"MyCloud/utils"
	"github.com/gin-gonic/gin"
)

func main() {
	utils.RedisInit()
	utils.LogInit()

	// Disable Console Color
	// gin.DisableConsoleColor()

	// 使用默认中间件创建一个gin路由器
	// logger and recovery (crash-free) 中间件
	router := gin.Default()
	server.UrlMap(router)
	_ = router.Run()
}
