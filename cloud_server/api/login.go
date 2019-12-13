package api

import (
	"MyCloud/cloud_server/models"
	"MyCloud/cloud_server/repository"
	"MyCloud/conf"
	"MyCloud/utils"
	"database/sql"
	"fmt"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"net/http"
)

var errCheck = utils.ErrCheck
var logging = utils.Logging
var userManager = repository.NewUserManager()

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

	// 查看redis中是否有用户登录信息
	userInfo, err := userManager.GetCache(user)
	errCheck(err, "Sign:Error reading redis user information", false)
	haveCache := true
	if userInfo == nil {
		// 如果没有则取mysql中的数据
		haveCache = false
		userInfo, err = userManager.Select(user)
		if err == sql.ErrNoRows {
			g.JSON(http.StatusOK, gin.H{
				"errmsg": "用户不存在",
				"data":   nil,
			})
			return
		}
		errCheck(err, "Sign:Error reading mysql user information", true)
	}

	// 判断密码是否正确
	if bcrypt.CompareHashAndPassword([]byte(userInfo.Pwd), []byte(pwd)) != nil {
		logging.Info(fmt.Printf("Sign:User [%s] login failed", user))
		g.JSON(http.StatusOK, gin.H{
			"errmsg": "用户名或密码错误",
			"data":   nil,
		})
	} else {
		if haveCache == false {
			// 如果redis没有用户信息则添加
			err = userManager.SetCache(user, userInfo)
			errCheck(err, "Register:Error set redis user information", false)
		}
		// 生成token，并存入redis
		token, err := utils.CreatToken()
		errCheck(err, "Register:Error creat token", true)
		err = userManager.SetCache("token_"+token, userInfo)
		errCheck(err, "Register:Error set token", true)

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

	// 判断用户名是否已存在
	_, err := userManager.Select(user)
	if err != sql.ErrNoRows {
		errCheck(err, "Register:Error Query mysql UserInfo", true)
		g.JSON(http.StatusOK, gin.H{
			"errmsg": "用户已存在",
			"data":   nil,
		})
		return
	}

	// 密码加密
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(pwd), bcrypt.DefaultCost)
	errCheck(err, "Register:Failed to register password encryption", true)

	// 保存新密码
	userInfo := models.UserInfo{
		User:  user,
		Pwd:   string(hashedPassword),
		Email: email,
		Level: "1",
		Phone: nil,
	}
	uid, err := userManager.Insert(&userInfo)
	errCheck(err, "Register:Error Exec mysql UserInfo", true)
	userInfo.Id = uid

	// 保存token信息
	token, err := utils.CreatToken()
	errCheck(err, "Register:Error creat token", true)
	err = userManager.SetCache("token_"+token, &userInfo)
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
	err := userManager.DelCache(token)
	errCheck(err, "Logout:Error del token", true)

	g.JSON(http.StatusOK, gin.H{
		"errmsg": "ok",
		"data":   nil,
	})

}