package api

import (
	"MyCloud/cloud_server/models"
	"MyCloud/cloud_server/repository"
	"MyCloud/conf"
	"MyCloud/utils"
	"database/sql"
	"fmt"
	"net/http"
	"regexp"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

var errCheck = utils.ErrCheck
var logging = utils.Logging
var userManager = repository.NewUserManager()
var fileManager = repository.NewFileManager()
var userFileManager = repository.NewUserFileManager()
var fileBlockManager = repository.NewFileBlockManager()
var blockManager = repository.NewBlockManager()
var dirManager = repository.NewDirManager()

// 用户登录
func Sign(g *gin.Context) {
	user := g.Query("user")
	pwd := g.Query("pwd")

	if user == "" || pwd == "" {
		g.JSON(http.StatusOK, gin.H{
			"errmsg": "用户名及密码不能为空",
			"data":   nil,
		})
		return
	}

	// 查看redis中是否有用户登录信息
	userMate, err := userManager.GetByUser(user)
	if err == sql.ErrNoRows {
		g.JSON(http.StatusOK, gin.H{
			"errmsg": "用户不存在",
			"data":   nil,
		})
		return
	}
	errCheck(g, err, "Sign:Failed to read user information", http.StatusInternalServerError)

	// 判断密码是否正确
	if bcrypt.CompareHashAndPassword([]byte(userMate.Pwd), []byte(pwd)) != nil {
		logging.Info(fmt.Printf("Sign:User [%s] login failed", user))
		g.JSON(http.StatusOK, gin.H{
			"errmsg": "用户名或密码错误",
			"data":   nil,
		})
	} else {
		// 生成token，并存入redis
		token, err := utils.CreatToken()
		errCheck(g, err, "Register:Failed to creat token", http.StatusInternalServerError)
		err = userManager.SetCache("token_"+token, userMate)
		errCheck(g, err, "Register:Failed to set token", http.StatusInternalServerError)

		logging.Info(fmt.Printf("Sign:User [%s] login succeeded", user))
		g.SetCookie(
			"token", token, conf.COOKIE_MAXAGE, "/",
			"localhost", false, true)

		data := make(map[string]interface{})
		data["level"] = userMate.Level
		data["token"] = token
		g.JSON(http.StatusOK, gin.H{
			"errmsg": "ok",
			"data":   data,
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

	if len(pwd) < 6 {
		g.JSON(http.StatusOK, gin.H{
			"errmsg": "密码长度过短",
			"data":   nil,
		})
		return
	}

	// 判断用户名是否已存在
	_, err := userManager.GetSqlByUser(user)
	if err != sql.ErrNoRows {
		errCheck(g, err, "Register:Failed to Query mysql UserInfo", http.StatusInternalServerError)
		g.JSON(http.StatusOK, gin.H{
			"errmsg": "用户已存在",
			"data":   nil,
		})
		return
	}

	// 密码加密
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(pwd), bcrypt.DefaultCost)
	errCheck(g, err, "Register:Failed to register password encryption", http.StatusInternalServerError)

	// 保存新密码
	userMate := models.UserInfo{
		User:  user,
		Pwd:   string(hashedPassword),
		Email: email,
		Level: "1",
	}

	haveUppercase, _ := regexp.MatchString("[A-Z].*", pwd)
	haveLowercase, _ := regexp.MatchString("[a-z].*", pwd)
	haveNumber, _ := regexp.MatchString("\\d.*", pwd)
	if haveUppercase && haveLowercase && haveNumber {
		userMate.PwdStrength = "强"
	} else if haveUppercase || haveLowercase && haveNumber {
		userMate.PwdStrength = "普通"
	} else {
		userMate.PwdStrength = "弱"
	}

	// 保存token信息
	token, err := utils.CreatToken()
	errCheck(g, err, "Register:Failed to creat token", http.StatusInternalServerError)
	_, err = userManager.Set("token_"+token, userMate)
	errCheck(g, err, "Register:Failed to set user info", http.StatusInternalServerError)

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
	token := g.GetHeader("Authorization")
	err := userManager.DelCache(token)
	errCheck(g, err, "Logout:Error del token", http.StatusInternalServerError)

	g.JSON(http.StatusOK, gin.H{
		"errmsg": "ok",
		"data":   nil,
	})
}
