package routHandlers

import (
	_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
	"hash/fnv"
)

type Users struct {
	gorm.Model
	Username    string `json:"username"`
	PhoneNumber string `json:"phone_number"`
	Name        string `json:"name"`
	Sex         bool   `json:"sex"`
	HashCode    string `json:"hash_code"`
}

func hash(s string) string {
	h := fnv.New32a()
	h.Write([]byte(s))
	return string(h.Sum32())
}
func RegisterModel(db *gorm.DB, username string, sex bool, phonenumber string, name string) string {
	user := Users{
		Username:    username,
		PhoneNumber: phonenumber,
		Name:        name,
		Sex:         sex,
	}
	pre_user := Users{Username: username}
	if db.Where(&pre_user).Find(&pre_user).RecordNotFound() {
		db.Save(&user)
		return "a"
	} else {
		return "r"
	}
}
