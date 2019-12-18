package api

import (
	"database/sql"
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
	new_pwd := g.PostForm("new_pwd")
	rnew_pwd := g.PostForm("rnew_pwd")

	if new_pwd != rnew_pwd {
		g.JSON(http.StatusOK, gin.H{
			"errmsg": "两次输入的新密码不相同",
			"data":   nil,
		})
		return
	}

	userInfo, err := userManager.GetCache(token)
	errCheck(g, err, "PasswordChange:Error reading redis token information", 0)
	if userInfo == nil {
		userInfo, err = userManager.SelectByUser(user)
		if err == sql.ErrNoRows { // 如果没有返回结果，error的值会是sql.ErrNoRows
			g.JSON(http.StatusOK, gin.H{
				"errmsg": "用户名未被注册",
				"data":   nil,
			})
			return
		}
		errCheck(g, err, "PasswordChange:Error Scan mysql UserInfo", http.StatusInternalServerError)
	}

	if bcrypt.CompareHashAndPassword([]byte(userInfo.Pwd), []byte(pwd)) != nil {
		logging.Info(fmt.Printf("PasswordChange:User [%s] login failed", user))
		g.JSON(http.StatusOK, gin.H{
			"errmsg": "用户名或密码错误",
			"data":   nil,
		})
	} else {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(new_pwd), bcrypt.DefaultCost)
		errCheck(g, err, "PasswordChange:Failed to register password encryption", http.StatusInternalServerError)
		err = userManager.UpdatePassword(user, string(hashedPassword))
		errCheck(g, err, "PasswordChange:Error Exec mysql UserInfo", http.StatusInternalServerError)

		userInfo.Pwd = string(hashedPassword)
		err = userManager.SetCache(user, userInfo)
		errCheck(g, err, "PasswordChange:Error set redis user information", 0)
		err = userManager.SetCache("token_"+token, userInfo)
		errCheck(g, err, "PasswordChange:Error set token", http.StatusInternalServerError)

		g.JSON(http.StatusOK, gin.H{
			"errmsg": "ok",
			"data":   nil,
		})
	}
}
