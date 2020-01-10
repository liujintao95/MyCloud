package api

import (
	"MyCloud/cloud_server/models"
	"fmt"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"net/http"
	"regexp"
)

func ShowUser(g *gin.Context) {
	userInter, _ := g.Get("userInfo")
	userMate := userInter.(models.UserInfo)

	g.JSON(http.StatusOK, gin.H{
		"errmsg": "ok",
		"data":   userMate,
	})
}

// 修改密码
func PasswordChange(g *gin.Context) {
	token := g.GetHeader("Authorization")
	pwd := g.PostForm("pwd")
	newPwd := g.PostForm("newPwd")
	rNewPwd := g.PostForm("rNewPwd")
	userInter, _ := g.Get("userInfo")
	userMate := userInter.(models.UserInfo)

	if newPwd != rNewPwd {
		g.JSON(http.StatusOK, gin.H{
			"errmsg": "两次输入的新密码不相同",
			"data":   nil,
		})
		return
	}

	if len(newPwd) < 6 {
		g.JSON(http.StatusOK, gin.H{
			"errmsg": "密码长度过短",
			"data":   nil,
		})
		return
	}

	haveUppercase, _ := regexp.MatchString("[A-Z].*", newPwd)
	haveLowercase, _ := regexp.MatchString("[a-z].*", newPwd)
	haveNumber, _ := regexp.MatchString("\\d.*", newPwd)
	if haveUppercase && haveLowercase && haveNumber {
		userMate.PwdStrength = "强"
	} else if haveUppercase || haveLowercase && haveNumber {
		userMate.PwdStrength = "普通"
	} else {
		userMate.PwdStrength = "弱"
	}

	if bcrypt.CompareHashAndPassword([]byte(userMate.Pwd), []byte(pwd)) != nil {
		logging.Info(fmt.Printf("PasswordChange:User [%s] login failed", userMate.User))
		g.JSON(http.StatusOK, gin.H{
			"errmsg": "用户名或密码错误",
			"data":   nil,
		})
	} else {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPwd), bcrypt.DefaultCost)
		errCheck(g, err, "PasswordChange:Failed to register password encryption", http.StatusInternalServerError)
		userMate.Pwd = string(hashedPassword)

		err = userManager.Update("token_"+token, userMate)
		errCheck(g, err, "PasswordChange:Error update password", http.StatusInternalServerError)

		g.JSON(http.StatusOK, gin.H{
			"errmsg": "ok",
			"data":   nil,
		})
	}
}

func UsernameChange(g *gin.Context) {
	token := g.GetHeader("Authorization")
	newName := g.PostForm("name")
	userInter, _ := g.Get("userInfo")
	userMate := userInter.(models.UserInfo)

	userMate.Name = newName
	err := userManager.Update("token_"+token, userMate)
	errCheck(g, err, "UsernameChange:Error update user name", http.StatusInternalServerError)

	g.JSON(http.StatusOK, gin.H{
		"errmsg": "ok",
		"data":   nil,
	})
}

func PhoneChange(g *gin.Context) {
	token := g.GetHeader("Authorization")
	phoneStr := g.PostForm("phone")
	userInter, _ := g.Get("userInfo")
	userMate := userInter.(models.UserInfo)

	userMate.Phone.String = phoneStr
	userMate.Phone.Valid = true
	err := userManager.Update("token_"+token, userMate)
	errCheck(g, err, "PhoneChange:Error update user name", http.StatusInternalServerError)

	g.JSON(http.StatusOK, gin.H{
		"errmsg": "ok",
		"data":   nil,
	})
}

func EmailChange(g *gin.Context) {
	token := g.GetHeader("Authorization")
	email := g.PostForm("email")
	userInter, _ := g.Get("userInfo")
	userMate := userInter.(models.UserInfo)

	userMate.Email = email
	err := userManager.Update("token_"+token, userMate)
	errCheck(g, err, "UsernameChange:Error update user name", http.StatusInternalServerError)

	g.JSON(http.StatusOK, gin.H{
		"errmsg": "ok",
		"data":   nil,
	})
}
