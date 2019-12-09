package api

import (
	"MyCloud/cloud_server/models"
	"MyCloud/conf"
	"MyCloud/utils"
	"encoding/json"
	"github.com/garyburd/redigo/redis"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"net/http"
)

var errCheck = utils.ErrCheck
var logging = utils.Logging

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

	jsonData, err := redis.Bytes(rc.Do("Get", user))
	errCheck(err, "Sign:Error reading redis user information", false)
	userInfo := models.UserInfo{}
	if jsonData == nil {
		rows, err := utils.Conn.Query("SELECT * FROM user_info where user = ?", user)
		errCheck(err, "Sign:Error Query mysql UserInfo", true)

		err = rows.Scan(
			&userInfo.Id, &userInfo.User, &userInfo.Pwd,
			&userInfo.Level, &userInfo.Email, &userInfo.Phone,
		)
		errCheck(err, "Sign:Error Scan mysql UserInfo", true)
	} else {
		_ = json.Unmarshal(jsonData, userInfo)
	}

	if bcrypt.CompareHashAndPassword([]byte(userInfo.Pwd), []byte(pwd)) != nil {
		logging.Info("Sign:User [%s] login failed", user)
		g.JSON(http.StatusOK, gin.H{
			"errmsg": "用户名或密码错误",
			"data":  nil,
		})
	} else {
		token := utils.CreatToken(userInfo)
		logging.Info("Sign:User [%s] login succeeded", user)
		g.SetCookie(
			"token", token, conf.COOKIE_MAXAGE, "/",
			"localhost", false, true)
		g.JSON(http.StatusOK, gin.H{
			"errmsg": "ok",
			"data":  nil,
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

	if pwd != rpwd {
		g.JSON(http.StatusOK, gin.H{
			"errmsg": "两次密码输入不一致",
			"data":   nil,
		})
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(pwd), bcrypt.DefaultCost)
	errCheck(err, "Register:Failed to register password encryption", true)
	res, err := utils.Conn.Exec(
		"INSERT INTO user_info(ui_user,ui_pwd,ui_email) VALUES (?,?,?)",
		user, string(hashedPassword), email)
	errCheck(err, "Register:Error Exec mysql UserInfo", true)
	id, err := res.LastInsertId()
	errCheck(err, "Register:Error get insert last id", true)

	userInfo := models.UserInfo{
		Id:    id,
		User:  user,
		Pwd:   string(hashedPassword),
		Email: email,
	}
	token := utils.CreatToken(userInfo)

	jsonData, err := json.Marshal(userInfo)
	errCheck(err, "Register:Error json marshal user information", true)
	rc := utils.RedisClient.Get()
	defer rc.Close()
	_, err = rc.Do("SET", user, jsonData, "EX", conf.REDIS_MAXAGE)
	errCheck(err, "Register:Error set redis user information", true)

	g.SetCookie(
		"token", token, conf.COOKIE_MAXAGE, "/",
		"localhost", false, true)
	g.JSON(http.StatusOK, gin.H{
		"errmsg": "ok",
		"data":   nil,
	})
}

// 登出
func Logout(g *gin.Context)  {
	token, _ := g.Cookie("token")
	if utils.DelToken(token) == false{
		g.JSON(http.StatusInternalServerError, gin.H{
			"errmsg": "服务器内部错误",
			"data":   nil,
		})
	} else {
		g.JSON(http.StatusOK, gin.H{
			"errmsg": "ok",
			"data":   nil,
		})
	}
}
