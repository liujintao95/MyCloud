package utils

import (
	"MyCloud/cloud_server/models"
	"MyCloud/conf"
	"encoding/json"
	"github.com/dgrijalva/jwt-go"
	"time"
)

func ErrCheck(err error, msg string, risk bool) {
	if err != nil {
		Logging.Error(msg)
		if risk {
			panic("")
		}
	}
}

func CreatToken(userInfo models.UserInfo) string {
	token := jwt.New(jwt.SigningMethodHS256)
	claims := make(jwt.MapClaims)
	claims["exp"] = time.Now().Add(time.Hour * time.Duration(1)).Unix()
	claims["iat"] = time.Now().Unix()
	token.Claims = claims
	tokenString, err := token.SignedString([]byte(""))
	ErrCheck(err, "Error while signing the token", true)

	jsonData, err := json.Marshal(userInfo) // 序列化结构体
	ErrCheck(err, "Error json marshal user information", true)
	rc := RedisClient.Get()
	defer rc.Close()
	_, err = rc.Do(
		"SET", "token_"+tokenString, jsonData, "EX", conf.REDIS_MAXAGE)
	ErrCheck(err, "Error set redis token information", true)
	return tokenString
}

func DelToken(key string) (res bool) {
	rc := RedisClient.Get()
	defer rc.Close()
	_, err := rc.Do("DEL", key)
	ErrCheck(err, "Error del redis token information", true)
	res = true
	return
}
