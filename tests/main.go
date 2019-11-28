package tests

import (
	_ "testing"
	"crypto/md5"
	"github.com/dgrijalva/jwt-go"
	"log"
	"strconv"
	"time"
)

func stringInSlice(str string, list []string) bool {
	for _, v := range list {
		if v == str {
			return true
		}
	}
	return false
}

func generateHash() []string {
	var hashCodes []string
	var currentHash [16]byte
	var currentHashStr string
	var min int
	var sec int
	var minString string
	var secString string
	for i := 0; i < 4; i++ {
		min = time.Now().Add(-time.Second * time.Duration(i)).Minute()
		sec = time.Now().Add(-time.Second * time.Duration(i)).Second()
		if min < 10 {
			minString = "0" + strconv.Itoa(min)
		} else {
			minString = strconv.Itoa(min)
		}
		if sec < 10 {
			secString = "0" + strconv.Itoa(sec)
		} else {
			secString = strconv.Itoa(sec)
		}
		currentHash = md5.Sum([]byte(minString + secString))
		for _, item := range currentHash {
			currentHashStr += strconv.Itoa(int(item))
		}
		hashCodes = append(hashCodes, currentHashStr)
		currentHashStr = ""
	}
	return hashCodes
}

func main() {
	//hashCodes := generateHash()
	//log.Print(hashCodes)
	//log.Print(time.Now().Add(-10*time.Second))
	token, _ := jwt.Parse("eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJ1c2VyX2lkIjoxMjN9.NYlecdiqVuRg0XkWvjFvpLvglmfR1ZT7f8HeDDEoSx8", func(token *jwt.Token) (interface{}, error) {
		return []byte("abC123!"), nil
	})
	if _, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		log.Print("valid")
	}
}
