package models

import (
	_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
	"log"
	"time"
)

type ChatRoom struct {
	gorm.Model
	UserFrom      string `json:"user_from"`
	UserTo        string `json:"user_to"`
	UserFrom_rate uint   `json:"userFrom_rate"`
	UserTo_rate   uint   `json:"userTo_rate"`
}

type Message struct {
	gorm.Model
	UserFrom   string `json:"user_from"`
	UserTo     string `json:"user_to"`
	Text       string `json:"text"`
	ChatroomID uint   `json:"chatroom_id"`
}

type Shop_Users struct {
	gorm.Model
	Username    string    `json:"username"`
	Time_Logged time.Time `json:"time_logged"`
	ShopID      uint16    `json:"shop_id"`
	Condition   string    `json:"condition"`
}
type GameRating struct {
	gorm.Model
	ChatroomID   uint `json:"chatroom_id"`
	UserFromRate int  `json:"user_from_rate"`
	UserToRate   int  `json:"user_to_rate"`
}

type Users struct {
	gorm.Model
	Username    string `json:"username"`
	PhoneNumber string `json:"phone_number"`
	Name        string `json:"name"`
	Sex         bool   `json:"sex"`
	HashCode    string `json:"hash_code"`
}

type XO struct {
	gorm.Model
	Starter       bool   `json:"starter"`
	ChatroomID    uint   `json:"chatroom_id"`
	UserFromMoves string `json:"user_from_moves"`
	UserToMoves   string `json:"user_to_moves"`
}

type ChatroomGameStarted struct {
	gorm.Model
	ChatroomID uint `json:"chatroom_id"`
	GameID     uint `json:"game_id"`
}

type PlayersResult struct {
	gorm.Model
	Username string `json:"username"`
	Score    uint   `json:"score"`
	WinRate  uint   `json:"win_rate"`
	LoseRate uint   `json:"lose_rate"`
}

type GameRatingLog struct {
	gorm.Model
	ChatroomID   uint `json:"chatroom_id"`
	UserFromRate int  `json:"user_from_rate"`
	UserToRate   int  `json:"user_to_rate"`
}

type UsersAward struct {
	gorm.Model
	Username    string `json:"username"`
	PhoneNumber string `json:"phone_number"`
	Name        string `json:"name"`
}

type AuthCodes struct {
	gorm.Model
	Username string `json:"username"`
	LastCode string `json:"last_code"`
}

func DbSqlMigration(url string) *gorm.DB {
	var db *gorm.DB
	var err error
	for i := 1; i < 5; i++ {
		db, err = gorm.Open("mysql", url)
		if err != nil {
			log.Println(err)
			time.Sleep(time.Duration(4) * time.Second)
		}
	}
	db = db.Set("gorm:table_options", "ENGINE=InnoDB CHARSET=utf8mb4 auto_increment=1")
	db.AutoMigrate(&ChatRoom{})
	db.AutoMigrate(&Message{})
	db.AutoMigrate(&Shop_Users{})
	db.AutoMigrate(&GameRating{})
	db.AutoMigrate(&Users{})
	db.AutoMigrate(&XO{})
	db.AutoMigrate(&ChatroomGameStarted{})
	db.AutoMigrate(&PlayersResult{})
	db.AutoMigrate(&GameRatingLog{})
	db.AutoMigrate(&UsersAward{})
	db.AutoMigrate(&AuthCodes{})
	return db
}
