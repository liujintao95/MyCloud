package api

import (
	"MyCloud/cloud_server/models"
	"fmt"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"net/http"
)

// 修改密码
func PasswordChange(g *gin.Context) {
	token, _ := g.Cookie("token")
	user := g.PostForm("user")
	pwd := g.PostForm("pwd")
	newPwd := g.PostForm("new_pwd")
	rNewPwd := g.PostForm("rnew_pwd")
	userInter, _ := g.Get("userInfo")
	userMate := userInter.(models.UserInfo)

	if newPwd != rNewPwd {
		g.JSON(http.StatusOK, gin.H{
			"errmsg": "两次输入的新密码不相同",
			"data":   nil,
		})
		return
	}

	if bcrypt.CompareHashAndPassword([]byte(userMate.Pwd), []byte(pwd)) != nil {
		logging.Info(fmt.Printf("PasswordChange:User [%s] login failed", user))
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
	token, _ := g.Cookie("token")
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
