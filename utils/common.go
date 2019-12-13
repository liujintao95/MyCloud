package utils

import (
	"github.com/dgrijalva/jwt-go"
	"time"
)

func ErrCheck(err error, msg string, risk bool) {
	if err != nil {
		Logging.Error(msg)
		if risk {
			panic(msg)
		}
	}
}

func CreatToken() (string, error) {
	token := jwt.New(jwt.SigningMethodHS256)
	claims := make(jwt.MapClaims)
	claims["exp"] = time.Now().Add(time.Hour * time.Duration(1)).Unix()
	claims["iat"] = time.Now().Unix()
	token.Claims = claims
	tokenString, err := token.SignedString([]byte(""))
	return tokenString, err
}
