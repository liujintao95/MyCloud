package api

import (
	"MyCloud/cloud_server/handlers"
	"MyCloud/cloud_server/models"
	"MyCloud/conf"
	"MyCloud/utils"
	"database/sql"
	"encoding/json"
	"fmt"
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

	rc := utils.RedisPool.Get()
	defer rc.Close()

	jsonData, err := redis.Bytes(rc.Do("LRANGE", user, 0, -1))
	errCheck(err, "Sign:Error reading redis user information", false)
	userInfo := models.UserInfo{}
	if jsonData != nil {
		_ = json.Unmarshal(jsonData, userInfo)
	} else {
		userInfo, err = handlers.GetUserInfo(user)
		if err == sql.ErrNoRows{
			g.JSON(http.StatusOK, gin.H{
				"errmsg": "用户不存在",
				"data":   nil,
			})
			return
		}
		errCheck(err, "Sign:Error reading mysql user information", true)
	}

	if bcrypt.CompareHashAndPassword([]byte(userInfo.Pwd), []byte(pwd)) != nil {
		logging.Info(fmt.Printf("Sign:User [%s] login failed", user))
		g.JSON(http.StatusOK, gin.H{
			"errmsg": "用户名或密码错误",
			"data":   nil,
		})
	} else {
		if jsonData == nil {
			jsonData, err := json.Marshal(userInfo)
			errCheck(err, "Register:Error json marshal user information", true)
			_, err = rc.Do("LPUSH", user, string(jsonData), "EX", string(conf.REDIS_MAXAGE))
			errCheck(err, "Register:Error set redis user information", true)
		}
		token := utils.CreatToken(userInfo)
		logging.Info(fmt.Printf("Sign:User [%s] login succeeded", user))
		g.SetCookie(
			"token", token, conf.COOKIE_MAXAGE, "/",
			"localhost", false, true)
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

	if pwd != rpwd {
		g.JSON(http.StatusOK, gin.H{
			"errmsg": "两次密码输入不一致",
			"data":   nil,
		})
		return
	}

	_, err := handlers.GetUserInfo(user)
	if err != sql.ErrNoRows {
		errCheck(err, "Register:Error Query mysql UserInfo", true)
		g.JSON(http.StatusOK, gin.H{
			"errmsg": "用户已存在",
			"data":   nil,
		})
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(pwd), bcrypt.DefaultCost)
	errCheck(err, "Register:Failed to register password encryption", true)

	userInfo := models.UserInfo{
		User:  user,
		Pwd:   string(hashedPassword),
		Email: email,
		Level: "1",
	}
	uid, err := handlers.SetNewUser(userInfo)
	errCheck(err, "Register:Error Exec mysql UserInfo", true)
	userInfo.Id = uid
	token := utils.CreatToken(userInfo)

	jsonData, err := json.Marshal(userInfo)
	errCheck(err, "Register:Error json marshal user information", true)
	rc := utils.RedisPool.Get()
	defer rc.Close()
	_, err = rc.Do("LPUSH", user, string(jsonData), "EX", string(conf.REDIS_MAXAGE))
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
func Logout(g *gin.Context) {
	token, _ := g.Cookie("token")
	if utils.DelToken(token) == false {
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

// 修改密码
func PasswordChange(g *gin.Context) {
	token, _ := g.Cookie("token")
	user := g.PostForm("user")
	pwd := g.PostForm("pwd")
	new_pwd := g.PostForm("new_pwd")
	rnew_pwd := g.PostForm("rnew_pwd")

	if new_pwd != rnew_pwd {
		g.JSON(http.StatusOK, gin.H{
			"errmsg": "两次输入的新密码不相同",
			"data":   nil,
		})
		return
	}

	rc := utils.RedisPool.Get()
	defer rc.Close()
	jsonData, err := redis.Bytes(rc.Do("LRANGE", token, 0, -1))
	errCheck(err, "PasswordChange:Error reading redis token information", false)
	userInfo := models.UserInfo{}
	if jsonData != nil {
		_ = json.Unmarshal(jsonData, userInfo)
	} else {
		userInfo, err = handlers.GetUserInfo(user)
		if err == sql.ErrNoRows { // 如果没有返回结果，error的值会是sql.ErrNoRows
			g.JSON(http.StatusOK, gin.H{
				"errmsg": "用户名未被注册",
				"data":   nil,
			})
			return
		}
		errCheck(err, "PasswordChange:Error Scan mysql UserInfo", true)
	}
	if bcrypt.CompareHashAndPassword([]byte(userInfo.Pwd), []byte(pwd)) != nil {
		logging.Info(fmt.Printf("PasswordChange:User [%s] login failed", user))
		g.JSON(http.StatusOK, gin.H{
			"errmsg": "用户名或密码错误",
			"data":   nil,
		})
	} else {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(new_pwd), bcrypt.DefaultCost)
		errCheck(err, "PasswordChange:Failed to register password encryption", true)
		err = handlers.UpdateUserPwd(user, string(hashedPassword))
		errCheck(err, "PasswordChange:Error Exec mysql UserInfo", true)

		userInfo.Pwd = string(hashedPassword)
		jsonData, err := json.Marshal(userInfo)
		errCheck(err, "PasswordChange:Error json marshal user information", true)
		_, err = rc.Do("LPUSH", user, string(jsonData), "EX", string(conf.REDIS_MAXAGE))
		errCheck(err, "PasswordChange:Error set redis user information", true)
		_, err = rc.Do("LPUSH", token, string(jsonData), "EX", string(conf.REDIS_MAXAGE))
		errCheck(err, "PasswordChange:Error set redis user information", true)

		g.JSON(http.StatusOK, gin.H{
			"errmsg": "ok",
			"data":   nil,
		})
	}
}
