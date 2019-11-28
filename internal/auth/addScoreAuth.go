package auth

import (
	"github.com/dgrijalva/jwt-go"
	"github.com/jinzhu/gorm"
)

type AuthCodes struct {
	gorm.Model
	Username string `json:"username"`
	LastCode string `json:"last_code"`
}

func ScoreAuth(jwtCode string, username string, db *gorm.DB) bool {
	user := AuthCodes{}
	notFound := db.Where(&AuthCodes{Username:username}).First(&user).RecordNotFound()
	if notFound {
		return false
	}
	if user.LastCode == jwtCode {
		return false
	}


	token, _ := jwt.Parse(jwtCode, func(token *jwt.Token) (interface{}, error) {
		return []byte("abC123!"), nil
	})
	if _, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		user.LastCode = jwtCode
		db.Save(&user)
		return true
	}
	return false
}