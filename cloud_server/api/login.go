package api

import (
	"MyCloud/cloud_server"
	"MyCloud/utils"
	"github.com/garyburd/redigo/redis"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"net/http"
)

var errCheck = utils.ErrCheck
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
	errCheck(err, "Sign:Error reading redis user information", false)

	if val == "" {
		user_info := new(server.UserInfo)

		rows, err := db.Query("SELECT pwd FROM user_info where user = ?", user)
		errCheck(err, "Sign:Error Query mysql UserInfo", true)

		err = rows.Scan(&user_info.Pwd)
		errCheck(err, "Sign:Error Scan mysql UserInfo", true)

		val = user_info.Pwd
	}

	if bcrypt.CompareHashAndPassword([]byte(val), []byte(pwd)) != nil {
		logging.Info("Sign:User [%s] login failed", user)
		g.JSON(http.StatusOK, gin.H{
			"errmsg": "用户名或密码错误",
			"data":   nil,
		})
	} else {
		logging.Info("Sign:User [%s] login succeeded", user)
		g.JSON(http.StatusOK, gin.H{
			"errmsg": "ok",
			"data":   nil,
		})
	}
}


// 注册
func Register(g *gin.Context) {
	email := g.PostForm("email")
	user := g.PostForm("user")
	pwd := g.PostForm("pwd")
	rpwd := g.PostForm("rpwd")

	if user == "" || pwd == "" || rpwd == "" || email == "" {
		g.JSON(http.StatusOK, gin.H{
			"errmsg": "用户名、密码及邮箱不能为空",
			"data":   nil,
		})
		return
	}

	if pwd != rpwd{
		g.JSON(http.StatusOK, gin.H{
			"errmsg": "两次密码输入不一致",
			"data":   nil,
		})
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(pwd), bcrypt.DefaultCost)
	errCheck(err, "Register:Failed to register password encryption", true)

	_, err = db.Exec(
		"INSERT INTO users(ui_user,ui_pwd,ui_email) values(?,?,?)",
		user, string(hashedPassword), email)
	errCheck(err, "Register:Error Exec mysql UserInfo", true)

	rc := utils.RedisClient.Get()
	defer rc.Close()
	_, err = rc.Do("SET", user, string(hashedPassword))
	errCheck(err, "Register:Error set redis user information", true)

	g.JSON(http.StatusOK, gin.H{
		"errmsg": "ok",
		"data":   nil,
	})
}
