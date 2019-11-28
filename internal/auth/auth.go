package auth

import (
	"crypto/md5"
	"encoding/hex"
	"net/http"
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
		currentHashStr += hex.EncodeToString(currentHash[:])
		hashCodes = append(hashCodes, currentHashStr)
		currentHashStr = ""
	}
	return hashCodes
}

func MiddleWareCheckAutch(r *http.Request) bool {
	clientKey := r.Header.Get("Authorization")
	if stringInSlice(clientKey, generateHash()) {
		return true
	} else {
		return false
	}
}
