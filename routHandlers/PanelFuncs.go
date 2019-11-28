package routHandlers

import (
	"encoding/json"
	"github.com/ghiac/go-commons/log"
	"github.com/gorilla/mux"
	"github.com/jinzhu/gorm"
	"net/http"
	"time"
)

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

func AddUserResult(joinedGame GameRating, joinedChatRoom ChatRoom, db *gorm.DB) {
	winnerUsername := ""
	loserUsername := ""
	var winnerScore int
	var losserScore int
	if joinedGame.UserToRate != 0 && joinedGame.UserFromRate != 0 {
		if joinedGame.UpdatedAt.After(time.Now().Add(-time.Second * 10)) {
			preGameLog := GameRatingLog{}
			noLogFound := db.Where(&GameRatingLog{ChatroomID: joinedGame.ChatroomID}).Last(&preGameLog).RecordNotFound()
			if noLogFound || preGameLog.CreatedAt.Before(time.Now().Add(-time.Minute*2)) {
				db.Save(&GameRatingLog{
					ChatroomID:   joinedGame.ChatroomID,
					UserFromRate: joinedGame.UserFromRate,
					UserToRate:   joinedGame.UserToRate,
				})
			}
		}
		if joinedGame.UserFromRate > joinedGame.UserToRate {
			winnerUsername = joinedChatRoom.UserFrom
			loserUsername = joinedChatRoom.UserTo
			winnerScore = joinedGame.UserFromRate
			losserScore = joinedGame.UserToRate
		} else {
			winnerUsername = joinedChatRoom.UserTo
			loserUsername = joinedChatRoom.UserFrom
			winnerScore = joinedGame.UserToRate
			losserScore = joinedGame.UserFromRate
		}
		player := PlayersResult{}
		notFound := db.Where(PlayersResult{Username: winnerUsername}).First(&player).RecordNotFound()
		if notFound == true {
			db.Save(&PlayersResult{
				Username: winnerUsername,
				Score:    uint(winnerScore),
				WinRate:  1,
				LoseRate: 0,
			})
		} else {
			if player.UpdatedAt.Before(time.Now().Add(-time.Minute * 3)) {
				player.WinRate += 1
				player.Score += uint(winnerScore)
				db.Save(&player)
			}
		}
		playerLoser := PlayersResult{}
		notFoundLoser := db.Where(PlayersResult{Username: loserUsername}).First(&playerLoser).RecordNotFound()
		if notFoundLoser == true {
			log.Logger.Print("loser not found")
			db.Save(&PlayersResult{
				Username: loserUsername,
				Score:    uint(losserScore),
				WinRate:  0,
				LoseRate: 1,
			})
		} else {
			if playerLoser.UpdatedAt.Before(time.Now().Add(-time.Minute * 3)) {
				playerLoser.LoseRate += 1
				playerLoser.Score += uint(losserScore)
				db.Save(&playerLoser)
			}
		}
	}
}

type GetPlayersGamesOutput struct {
	Username    string `json:"username"`
	Name        string `json:"name"`
	PhoneNumber string `json:"phone_number"`
	Score       uint   `json:"score"`
	WinRate     uint   `json:"win_rate"`
	LoseRate    uint   `json:"lose_rate"`
}
type GetPlayersGamesOutputArray struct {
	UsersCount uint                    `json:"users_count"`
	Output     []GetPlayersGamesOutput `json:"output"`
}

func GetPlayersGames(writer http.ResponseWriter, r *http.Request, db *gorm.DB) {
	output := GetPlayersGamesOutputArray{}
	db.Raw("select users.username,users.phone_number,users.name,pr.score,pr.win_rate,pr.lose_rate from users, players_results as pr where pr.username=users.username ").Scan(&output.Output)
	db.Model(&Users{}).Count(&output.UsersCount)
	js, _ := json.Marshal(&output)
	writer.Write(js)
}

type GetPlayersMatchesOutputResponse struct {
	UserFrom            string    `json:"user_from"`
	UserTo              string    `json:"user_to"`
	UserFromRate        uint      `json:"user_from_rate"`
	UserToRate          uint      `json:"user_to_rate"`
	UserFromPhonenumber string    `json:"user_from_phonenumber"`
	UserToPhonenumber   string    `json:"user_to_phonenumber"`
	CreatedAt           time.Time `json:"created_at"`
	Timestamp           int64     `json:"timestamp"`
}

func GetPhonenumberByUsers(username string, db *gorm.DB) string {
	user := Users{}
	db.Where(&Users{Username: username}).Find(&user)
	return user.PhoneNumber
}
func GetPlayersMatches(w http.ResponseWriter, r *http.Request, db *gorm.DB) {
	var output []GetPlayersMatchesOutputResponse
	db.Raw("select cr.user_from,cr.user_to,gr.user_from_rate,gr.user_to_rate,gr.created_at from game_rating_logs as gr, chat_rooms as cr where gr.chatroom_id = cr.id order by gr.created_at ").Scan(&output)
	for i := 0; i < len(output); i++ {
		output[i].UserFromPhonenumber = GetPhonenumberByUsers(output[i].UserFrom, db)
		output[i].UserToPhonenumber = GetPhonenumberByUsers(output[i].UserTo, db)
		output[i].Timestamp = output[i].CreatedAt.Unix()
	}
	js, _ := json.Marshal(&output)
	w.Write(js)
}
func SearchUsersByPhonenumberInGames(w http.ResponseWriter, r *http.Request, db *gorm.DB) {
	vars := mux.Vars(r)
	phonenumber := vars["phonenumber"]
	output := []GetPlayersGamesOutput{}
	db.Raw("select users.username,users.phone_number,users.name,pr.score,pr.win_rate,pr.lose_rate from users, players_results as pr where pr.username=users.username and users.phone_number=" + phonenumber).Scan(&output)
	js, _ := json.Marshal(&output)
	w.Write(js)
}

func GetUsernameByPhonenumber(phonenumber string, db *gorm.DB) string {
	user := Users{}
	db.Where(&Users{PhoneNumber: phonenumber}).Find(&user)
	return user.Username
}
func SearchUserByPhonenumberInMatches(w http.ResponseWriter, r *http.Request, db *gorm.DB) {
	vars := mux.Vars(r)
	phonenumber := vars["phonenumber"]
	var output []GetPlayersMatchesOutputResponse
	username := GetUsernameByPhonenumber(phonenumber, db)
	db.Raw("select cr.user_from,cr.user_to,gr.user_from_rate,gr.user_to_rate,gr.created_at from game_rating_logs as gr, chat_rooms as cr where gr.chatroom_id = cr.id and (cr.user_from = \"" + username + "\" or cr.user_to = \"" + username + "\") order by gr.created_at ").Scan(&output)
	for i := 0; i < len(output); i++ {
		output[i].UserFromPhonenumber = GetPhonenumberByUsers(output[i].UserFrom, db)
		output[i].UserToPhonenumber = GetPhonenumberByUsers(output[i].UserTo, db)
	}
	js, _ := json.Marshal(&output)
	w.Write(js)
}

func GiveAward(w http.ResponseWriter, r *http.Request, db *gorm.DB) {
	vars := mux.Vars(r)
	username := vars["username"]
	userAward := UsersAward{}
	user := Users{}
	db.Where(&Users{Username: username}).First(&user)
	notFound := db.Where(&UsersAward{
		Username:    username,
		PhoneNumber: user.PhoneNumber,
		Name:        user.Name,
	}).First(&userAward).RecordNotFound()
	if notFound {
		db.Save(&UsersAward{
			Username:    username,
			PhoneNumber: user.PhoneNumber,
			Name:        user.Name,
		})
		_, _ = w.Write([]byte("\"d\""))
	} else {
		if userAward.UpdatedAt.Before(time.Now().Add(-time.Hour * 24)) {
			userAward.UpdatedAt = time.Now()
			db.Save(&userAward)
			_, _ = w.Write([]byte("\"d\""))
		} else {
			_, _ = w.Write([]byte("\"e\""))
		}
	}
}


func GetRewardedPlayers(w http.ResponseWriter, r *http.Request, db *gorm.DB) {
	users := []Users{}
	db.Find(&users)
	js, _ := json.Marshal(users)
	w.Write(js)
}