package routHandlers

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/jinzhu/gorm"
)

func RootHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte("Asset not found\n"))
		return
	}
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("Running API v1\n"))
}

func WelcomePage(w http.ResponseWriter, r *http.Request, db *gorm.DB) {
	vars := mux.Vars(r)
	username := vars["username"]
	shop_id := vars["shop_id"]
	i, err := strconv.Atoi(shop_id)
	if err != nil {
		// handle parse error
		_, _ = w.Write([]byte("Error in parsing"))
	}
	updateWelcomeUsers(db, username, uint16(i))
	_, _ = w.Write(getOnlineUsers(db, username, uint16(i)))
}

func GetAcceptCompetitor(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userFrom := vars["user1"]
	userTo := vars["user2"]
	js, _ := json.Marshal(GetAcceptCompetitorRedis(userFrom, userTo))
	_, _ = w.Write(js)
}

func RequestToCompetition(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	user1 := vars["user1"]
	user2 := vars["user2"]
	result := newCompetition(competition{
		Username:           user1,
		Competitor:         user2,
		Time_asked:         time.Now(),
		CompetitorAccepted: false,
	})
	convertedResult := strconv.Itoa(result)
	_, _ = w.Write([]byte(convertedResult))
}

func DeclineCompetition(w http.ResponseWriter, r *http.Request, db *gorm.DB) {
	vars := mux.Vars(r)
	userFrom := vars["user_from"]
	userTo := vars["user_to"]
	chatroomID := vars["chatroom_id"]
	cID, _ := strconv.Atoi(chatroomID)
	DeclineCompetitionRedis(userFrom)
	DeclineCompetitionRedis(userTo)
	SendFakeMessage(db, userFrom, userTo, uint(cID))
	ResetGame(db, uint(cID))
	_, _ = w.Write([]byte("ok"))
}

func DeclineCompetitionStart(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userFrom := vars["user_from"]
	userTo := vars["user_to"]
	DeclineCompetitionRedis(userFrom)
	DeclineCompetitionRedis(userTo)
	_, _ = w.Write([]byte("ok"))
}

func RedirectToChatroom(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	username := vars["user"]
	js, _ := RedirectToChatroomRedis(username)
	_, _ = w.Write(js)
}

func GetUserCompetition(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	username := vars["username"]
	comp := findCompetition(username)
	js, _ := json.Marshal(comp.Username)
	_, _ = w.Write([]byte(string(js)))
}

func StartChat(w http.ResponseWriter, r *http.Request, db *gorm.DB) {
	vars := mux.Vars(r)
	userFrom := vars["user1"]
	userTo := vars["user2"]
	chatroomID := CreateNewChatroom(userFrom, userTo, db)
	js, err := json.Marshal(chatroomID)
	if err != nil {
		_, _ = w.Write([]byte("Error"))
	}
	_, _ = w.Write(js)
}

func SendMessage(w http.ResponseWriter, r *http.Request, db *gorm.DB) {
	vars := mux.Vars(r)
	reqBody, _ := ioutil.ReadAll(r.Body)
	chatroomID, _ := strconv.Atoi(vars["chatID"])
	if len(reqBody) == 0 {
		_, _ = w.Write([]byte("1"))
	} else {
		db.Create(&Message{
			Model:      gorm.Model{},
			UserFrom:   vars["user1"],
			UserTo:     vars["user2"],
			Text:       string(reqBody),
			ChatroomID: uint(chatroomID),
		})
	}
	_, _ = w.Write([]byte("1"))
}

type CheckMessages_Output struct {
	From        string `json:"from"`
	Text        string `json:"text"`
	Time_passed int    `json:"time_passed"`
}
type CheckMessages_result struct {
	Output       []CheckMessages_Output `json:"output"`
	GameRedirect uint                   `json:"game_redirect"`
}

func CheckMessages(w http.ResponseWriter, r *http.Request, db *gorm.DB) {
	vars := mux.Vars(r)
	output := CheckMessages_result{}
	userFrom := vars["user1"]
	userTo := vars["user2"]
	chatroomID := vars["chatID"]
	messages := GetChatroomChat(userFrom, userTo, chatroomID, db)
	var user_sent string
	sort.Slice(messages, func(i, j int) bool {
		return messages[i].CreatedAt.Before(messages[j].CreatedAt)
	})
	for _, item := range messages {
		if strings.Compare(userFrom, item.UserFrom) == 0 {
			user_sent = userFrom
		} else {
			user_sent = userTo
		}
		output.Output = append(output.Output, CheckMessages_Output{
			From:        user_sent,
			Text:        item.Text,
			Time_passed: int(time.Now().Sub(item.CreatedAt).Minutes()),
		})
	}
	chatroomIDInt, _ := strconv.Atoi(chatroomID)
	output.GameRedirect = GetChatroomGameCondition(uint(chatroomIDInt), db)
	js, err := json.Marshal(output)
	if err != nil {
		_, _ = w.Write([]byte("Error"))
	}
	_, _ = w.Write(js)
}

func CreateCompetition(w http.ResponseWriter, r *http.Request, db *gorm.DB) {
	vars := mux.Vars(r)
	gameID, _ := strconv.Atoi(vars["chatroomID"])
	user := vars["user"]
	gameJoined := SelectChatroomGame(uint(gameID), db)
	SendJoinedMessage(uint(gameID), user, gameJoined, db)
	CreateGameCompetition(uint(gameID), db)
	js, _ := json.Marshal(gameJoined)
	_, _ = w.Write(js)
}

func SetUsersScore(w http.ResponseWriter, r *http.Request, db *gorm.DB) {
	vars := mux.Vars(r)
	username := vars["user"]
	gameID, _ := strconv.Atoi(vars["gameID"])
	score, _ := strconv.Atoi(vars["score"])
	//clientKey := r.Header.Get("Authorization")
	//if !auth.ScoreAuth(clientKey) {
	//	w.Write([]byte("done"))
	//	return
	//}
	ClearChatroomGameCondition(uint(gameID), db)
	js, _ := json.Marshal(SetScoreModel(score, uint(gameID), username, db))
	_, _ = w.Write(js)
}

func GetCompetitionResult(w http.ResponseWriter, r *http.Request, db *gorm.DB) {
	vars := mux.Vars(r)
	username := vars["user"]
	gameID, _ := strconv.Atoi(vars["gameID"])
	js, _ := GetCompetitionResultModel(username, uint(gameID), db)
	_, _ = w.Write(js)
}

// XO game
func InitNewXOGame(w http.ResponseWriter, r *http.Request, db *gorm.DB) {
	vars := mux.Vars(r)
	chatroomID, _ := strconv.Atoi(vars["chatroom_id"])
	starter := InitNewXOGameModel(uint(chatroomID), db)
	js, _ := json.Marshal(starter)
	_, _ = w.Write(js)
}

func XOMove(w http.ResponseWriter, r *http.Request, db *gorm.DB) {
	vars := mux.Vars(r)
	username := vars["user"]
	chatroomID, _ := strconv.Atoi(vars["chatroom_id"])
	movement := vars["movement"]
	XOMoveModel(username, uint(chatroomID), movement, db)
	_, _ = w.Write([]byte("\"M\""))
}

func GetXOMoves(w http.ResponseWriter, r *http.Request, db *gorm.DB) {
	vars := mux.Vars(r)
	username := vars["user"]
	chatroomID, _ := strconv.Atoi(vars["chatroom_id"])
	_, _ = w.Write(GetXOMovesModel(username, uint(chatroomID), db))
}
