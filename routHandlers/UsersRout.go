package routHandlers

import (
	"crypto/md5"
	"encoding/json"
	"github.com/gorilla/mux"
	"github.com/jinzhu/gorm"
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
		for _, item := range currentHash {
			currentHashStr += strconv.Itoa(int(item))
		}
		hashCodes = append(hashCodes, currentHashStr)
		currentHashStr = ""
	}
	return hashCodes
}

func RegisterUser(w http.ResponseWriter, r *http.Request, db *gorm.DB) {
	vars := mux.Vars(r)
	username := vars["username"]
	sex := vars["sex"]
	name := vars["name"]
	phonenumber := vars["phonenumber"]
	ans := RegisterModel(db, username, sex == "1", phonenumber, name)
	js, _ := json.Marshal(ans)
	w.Write(js)
}

func UsersIdentify(w http.ResponseWriter, r *http.Request, db *gorm.DB) {
	vars := mux.Vars(r)
	token := vars["key"]
	if stringInSlice(token, generateHash()) {
		user := Users{
			Username: vars["username"],
		}
		db.Where(&user).Find(&user)
		user.UpdatedAt = time.Now()
		db.Save(&user)
		w.Write([]byte("ok"))
	}
}

func RemoveUser(w http.ResponseWriter, r *http.Request, db *gorm.DB) {
	vars := mux.Vars(r)
	user := Users{Username: vars["username"]}
	db.Where("username = ? ", vars["username"]).Delete(&user)
	w.Write([]byte("ok"))
}
