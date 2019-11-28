package routHandlers

import (
	"encoding/json"
	"log"
	"math/rand"
	"strconv"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
)

func getGameName(gameID uint) string {
	if gameID == 6 {
		return "tower game"
	}
	if gameID == 3 {
		return "hextris"
	}
	if gameID == 4 {
		return "X-O"
	}
	return "Not Defined"
}

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

func updateWelcomeUsers(db *gorm.DB, username string, shop_id uint16) {
	newUser := Shop_Users{
		Username:    username,
		Time_Logged: time.Now(),
		ShopID:      shop_id,
		Condition:   "online",
	}
	//userDB := Users{Username:username}
	//if db.Where(&userDB).Find(&userDB).RecordNotFound() {
	//	print("1")
	//	return
	//} else {
	//	if userDB.UpdatedAt.Before(time.Now().Add(-time.Second*10)) {
	//		print("2")
	//		return
	//	}
	//}

	user := Shop_Users{}
	if err := db.Where(&Shop_Users{Username: username, ShopID: shop_id}).First(&user).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			db.Create(&newUser)
		}
	} else {
		user.Time_Logged = time.Now()
		db.Save(&user)
	}
}

type getOnlineUsers_output struct {
	Username  string `json:"username"`
	Last_seen string `json:"last_seen"`
	Sex       bool   `json:"sex"`
}

func getOnlineUsers(db *gorm.DB, username string, shop_id uint16) []byte {
	var users []Shop_Users
	var gouo []getOnlineUsers_output
	db.Where("time_logged > ? and shop_id = ?", time.Now().Add(time.Minute*(-3)), shop_id).Find(&users)
	for _, item := range users {
		user := Users{}
		db.Where(&Users{Username: item.Username}).First(&user)
		if item.Username == username {
			continue
		}
		past_seconds := strconv.Itoa(int(time.Now().Sub(item.Time_Logged).Seconds()))
		gouo = append(gouo, getOnlineUsers_output{
			Username:  item.Username,
			Last_seen: past_seconds,
			Sex:       user.Sex,
		})
	}
	data, err := json.Marshal(gouo)
	if err != nil {
		return []byte("[]")
	}
	return data
}

func SendFakeMessage(db *gorm.DB, userFrom string, userTo string, chatroomID uint) {
	db.Save(&Message{
		UserFrom:   userFrom,
		UserTo:     userTo,
		Text:       "شخص مقابل بازگشت را انتخاب کرد",
		ChatroomID: chatroomID,
	})
}

func ResetGame(db *gorm.DB, chatroomID uint) {
	chatroomGame := ChatroomGameStarted{}
	db.Where(&ChatroomGameStarted{ChatroomID: chatroomID}).First(&chatroomGame)
	chatroomGame.GameID = 0
	db.Save(&chatroomGame)
}

func SendJoinedMessage(chatroomID uint, username string, gameID uint, db *gorm.DB) {
	joinedChatroom := ChatRoom{}
	err := db.Where("id = ?", chatroomID).First(&joinedChatroom).Error
	if err != nil {
		return
	}
	userTo := ""
	if joinedChatroom.UserFrom == username {
		userTo = joinedChatroom.UserTo
	} else {
		userTo = joinedChatroom.UserFrom
	}
	msg := "رقیب شما بازی " + getGameName(gameID) + " را شروع کرد"
	newMessage := Message{
		UserFrom:   username,
		UserTo:     userTo,
		Text:       msg,
		ChatroomID: chatroomID,
	}
	db.Save(&newMessage)
}

func CreateGameCompetition(chatroomID uint, db *gorm.DB) uint {
	preGameCompetition := GameRating{
		ChatroomID: chatroomID,
	}
	preGame := GameRating{}
	if db.Where(&preGameCompetition).First(&preGame).RecordNotFound() {
		db.Create(&GameRating{
			ChatroomID:   chatroomID,
			UserFromRate: 0,
			UserToRate:   0,
		})
		return chatroomID
	}
	preGame.UserToRate = 0
	preGame.UserFromRate = 0
	db.Save(&preGame)
	return preGame.ChatroomID
}

func SetScoreModel(score int, gameID uint, username string, db *gorm.DB) bool {
	joinedGame := GameRating{}
	err := db.Where(&GameRating{ChatroomID: gameID}).Find(&joinedGame).Error
	if err != nil {
		return false
	}
	joinedChatRoom := ChatRoom{}
	err = db.Where("id = ?", gameID).First(&joinedChatRoom).Error
	if err != nil {
		log.Print(err)
		return false
	}
	if username == joinedChatRoom.UserFrom {
		joinedGame.UserFromRate = score
	} else {
		joinedGame.UserToRate = score
	}
	db.Save(&joinedGame)
	AddUserResult(joinedGame, joinedChatRoom, db)
	return true
}

type GetCompetitionResultModelStruct struct {
	YourScore       int `json:"your_score"`
	CompetitorScore int `json:"competitor_score"`
}

func GetCompetitionResultModel(username string, gameID uint, db *gorm.DB) ([]byte, error) {
	joinedGame := GameRating{}
	err := db.Where(GameRating{ChatroomID: gameID}).Find(&joinedGame).Error
	if err != nil {
		return json.Marshal(false)
	}
	joinedChatRoom := ChatRoom{}
	err = db.Where("id = ?", joinedGame.ChatroomID).First(&joinedChatRoom).Error
	if err != nil {
		return json.Marshal(false)
	}
	ClearCompetetion(joinedChatRoom.UserFrom)
	ClearCompetetion(joinedChatRoom.UserTo)
	if username == joinedChatRoom.UserFrom {
		return json.Marshal(GetCompetitionResultModelStruct{
			YourScore:       joinedGame.UserFromRate,
			CompetitorScore: joinedGame.UserToRate,
		})
	} else {
		return json.Marshal(GetCompetitionResultModelStruct{
			YourScore:       joinedGame.UserToRate,
			CompetitorScore: joinedGame.UserFromRate,
		})
	}
}

func CreateNewChatroom(userFrom string, userTo string, db *gorm.DB) uint {
	previous_chatroom := ChatRoom{
		UserFrom: userFrom,
		UserTo:   userTo,
	}
	previous_chatroom_reverse := ChatRoom{
		UserFrom: userTo,
		UserTo:   userFrom,
	}
	pre_chat := ChatRoom{}
	chatr := ChatRoom{}

	if db.Where(&previous_chatroom).First(&pre_chat).RecordNotFound() && db.Where(&previous_chatroom_reverse).First(&pre_chat).RecordNotFound() {
		chatr = ChatRoom{
			Model:    gorm.Model{},
			UserFrom: userFrom,
			UserTo:   userTo,
		}
		db.Create(&chatr)
		return chatr.ID
	}
	return pre_chat.ID
}

func GetChatroomChat(userFrom string, userTo string, chatroomID string, db *gorm.DB) []Message {
	messages := []Message{}
	messages_reverse := []Message{}
	db.Where("user_from = ? AND user_to = ? AND chatroom_id = ?",
		userFrom, userTo, chatroomID).Find(&messages)
	db.Where("user_from = ? AND user_to = ? AND chatroom_id = ?",
		userTo, userFrom, chatroomID).Find(&messages_reverse)
	return append(messages, messages_reverse...)
}

func InitNewXOGameModel(chatroomID uint, db *gorm.DB) string {
	xo := XO{ChatroomID: chatroomID}
	if err := db.Where(&xo).First(&xo).RecordNotFound(); err == false {
		xo.UserFromMoves = ""
		xo.UserToMoves = ""
		if xo.UpdatedAt.Before(time.Now().Add(-time.Second * 10)) {
			if rand.Intn(10) > 5 {
				xo.Starter = true
			} else {
				xo.Starter = false
			}
		}
		db.Save(&xo)
	} else {
		xo = XO{
			Starter:       false,
			ChatroomID:    chatroomID,
			UserFromMoves: "",
			UserToMoves:   "",
		}
		db.Save(&xo)
	}
	chatroom := ChatRoom{}
	db.Where("id = ? ", chatroomID).First(&chatroom)
	if xo.Starter {
		return chatroom.UserFrom
	} else {
		return chatroom.UserTo
	}

}

func XOMoveModel(username string, chatroomID uint, movement string, db *gorm.DB) {
	xo := XO{ChatroomID: chatroomID}
	chatroom := ChatRoom{}
	err := db.Where(&xo).First(&xo).Error
	if err != nil {
		return
	}
	err = db.Where(" id = ? ", chatroomID).First(&chatroom).Error
	if err != nil {
		return
	}
	userFromMoves := strings.Split(xo.UserFromMoves, ",")
	userToMoves := strings.Split(xo.UserToMoves, ",")
	if chatroom.UserFrom == username && xo.Starter && len(userFromMoves) == len(userToMoves) {
		xo.UserFromMoves += movement + ","
	}
	if chatroom.UserFrom == username && !xo.Starter && len(userFromMoves) < len(userToMoves) {
		xo.UserFromMoves += movement + ","
	}
	if chatroom.UserTo == username && !xo.Starter && len(userFromMoves) == len(userToMoves) {
		xo.UserToMoves += movement + ","
	}
	if chatroom.UserTo == username && xo.Starter && len(userFromMoves) > len(userToMoves) {
		xo.UserToMoves += movement + ","
	}
	db.Save(&xo)
}

func GetXOMovesModel(username string, chatroomID uint, db *gorm.DB) []byte {
	var board [9]string
	xo := XO{ChatroomID: chatroomID}
	chatroom := ChatRoom{}
	err := db.Where(&xo).First(&xo).Error
	if err != nil {
		return nil
	}
	err = db.Where(" id = ? ", chatroomID).First(&chatroom).Error
	if err != nil {
		return nil
	}

	userFromMoves := strings.Split(xo.UserFromMoves, ",")
	userToMoves := strings.Split(xo.UserToMoves, ",")
	if username == chatroom.UserFrom && len(userFromMoves) > 0 && len(userToMoves) > 0 {
		for _, item := range userFromMoves {
			if item != "" {
				index, _ := strconv.Atoi(item)
				board[index] = "x"
			}
		}
		for _, item := range userToMoves {
			if item != "" {
				index, _ := strconv.Atoi(item)
				board[index] = "o"
			}
		}
	}
	if username == chatroom.UserTo && len(userFromMoves) > 0 && len(userToMoves) > 0 {
		for _, item := range userFromMoves {
			if item != "" {
				index, _ := strconv.Atoi(item)
				board[index] = "o"
			}
		}
		for _, item := range userToMoves {
			if item != "" {
				index, _ := strconv.Atoi(item)
				board[index] = "x"
			}
		}
	}
	js, _ := json.Marshal(board)
	return (js)
}

func SelectChatroomGame(chatroomID uint, db *gorm.DB) uint {
	games := []uint{3, 4, 6}
	rnd := rand.Intn(4)
	chatroomGame := ChatroomGameStarted{ChatroomID: chatroomID}
	db.Where(chatroomGame).Assign(ChatroomGameStarted{GameID: games[rnd]}).FirstOrCreate(&chatroomGame)
	return games[rnd]
}

func GetChatroomGameCondition(chatroomID uint, db *gorm.DB) uint {
	chatroomGame := ChatroomGameStarted{ChatroomID: chatroomID}
	err := db.Where(&chatroomGame).First(&chatroomGame).RecordNotFound()
	if err {
		return 0
	}
	return chatroomGame.GameID
}

func ClearChatroomGameCondition(chatroomID uint, db *gorm.DB) {
	chatroomGame := ChatroomGameStarted{ChatroomID: chatroomID}
	db.Where(chatroomGame).First(&chatroomGame)
	chatroomGame.GameID = 0
	db.Save(&chatroomGame)
}
