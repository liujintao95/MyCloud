package server

import (
	"MyCloud/cloud_server/models"
	"MyCloud/conf"
	"MyCloud/utils"
	"encoding/json"
	"github.com/garyburd/redigo/redis"
	"github.com/gin-gonic/gin"
	"net/http"
)

var errCheck = utils.ErrCheck

func LoginRequired(g *gin.Context) {
	token := g.GetHeader("Authorization")

	rc := utils.RedisPool.Get()
	defer rc.Close()

	userMate := models.UserInfo{}
	jsonData, err := redis.Bytes(rc.Do("GET", "token_"+token))
	errCheck(g, err, "Failed to get token", http.StatusInternalServerError)
	if jsonData != nil {
		_ = json.Unmarshal(jsonData, &userMate)
	}

	if userMate.Pwd == "" {
		g.JSON(http.StatusUnauthorized, gin.H{
			"errmsg": "user connection timeout, please login again!",
			"data":   nil,
		})
		return
	}

	// 重置超时时间
	_, err = rc.Do("SET", "token_"+token, string(jsonData), "EX", conf.REDIS_MAXAGE)
	errCheck(g, err, "Failed to reset timeout", http.StatusInternalServerError)

	g.Set("userInfo", userMate)
	g.Next()
}
