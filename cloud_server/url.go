package url

import (
	"MyCloud/cloud_server/api"
	"github.com/gin-gonic/gin"
)

func Register(router *gin.Engine) {
	router.POST("/sign", api.Sign)
}
