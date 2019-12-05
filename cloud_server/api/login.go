package api

import (
	"MyCloud/utils"
	"github.com/garyburd/redigo/redis"
	"github.com/gin-gonic/gin"
	"net/http"
)

var logging = utils.Logging
var db = utils.Conn

// 用户登录
func Sign(g *gin.Context) {
	user := g.PostForm("user")
	pwd := g.PostForm("pwd")

	if user == "" || pwd == "" {
		g.JSON(http.StatusOK, gin.H{
			"errmsg": "用户名及密码不能为空",
			"data":   nil,
		})
		return
	}

	rc := utils.RedisClient.Get()
	defer rc.Close()

	val, err := redis.String(rc.Do("Get", user))
	if err != nil {
		logging.Error("Error reading redis user information")
	}
	if val != "" {
		if val != pwd {
			logging.Info("User [%s] login failed", user)
			g.JSON(http.StatusOK, gin.H{
				"errmsg": "用户名或密码错误",
				"data":   nil,
			})
		} else {
			logging.Info("User [%s] login succeeded", user)
			g.JSON(http.StatusOK, gin.H{
				"errmsg": "ok",
				"data":   nil,
			})
		}
		return
	}

	stmt, err := db.Prepare("SELECT pwd FROM user_info where user = ?")
	if err != nil {
		logging.Error("Error reading mysql user information")
		return
	}
	rows, err := stmt.Query(user)
	if err != nil {
		logging.Error("Error reading mysql user information")
		return
	}
	rows.Close()
}
