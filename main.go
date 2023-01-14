package main

import (
	"fmt"
	"github.com/aicam/game_server/internal/auth"
	"github.com/aicam/game_server/internal/models"
	redis2 "github.com/aicam/game_server/internal/redis"
	"github.com/ghiac/go-commons/log"
	"net/http"
	"time"

	"github.com/aicam/game_server/routHandlers"
	"github.com/gomodule/redigo/redis"
	"github.com/gorilla/mux"
)

func enableCors(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
}

var Pool *redis.Pool

func main() {
	log.Initialize("info")
	log.Logger.Info("Started")

	router := mux.NewRouter()

	router.Methods("OPTIONS").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// fmt.Printf("OPTIONS")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, Access-Control-Request-Headers, Access-Control-Request-Method, Connection, Host, Origin, User-Agent, Referer, Cache-Control, X-header")
		w.WriteHeader(http.StatusNoContent)
		return
	})
	router.StrictSlash(true)
	// migration
	db := models.DbSqlMigration("root:021021ali@tcp(goapp_mysql:3306)/messenger_api?charset=utf8mb4&parseTime=True")

	log.Logger.Info("DB connected")
	// http handler
	router.HandleFunc("/welcome/{username}/{shop_id}", func(w http.ResponseWriter, r *http.Request) {
		enableCors(w)
		routHandlers.WelcomePage(w, r, db)
	}).Methods("GET")
	router.HandleFunc("/request_competition/{user1}/{user2}", func(writer http.ResponseWriter, request *http.Request) {
		enableCors(writer)
		routHandlers.RequestToCompetition(writer, request)
	}).Methods("GET")
	router.HandleFunc("/find_competition/{username}", func(writer http.ResponseWriter, request *http.Request) {
		enableCors(writer)
		routHandlers.GetUserCompetition(writer, request)
	}).Methods("GET")
	router.HandleFunc("/decline_competition/{user_from}/{user_to}/{chatroom_id}", func(w http.ResponseWriter, r *http.Request) {
		enableCors(w)
		routHandlers.DeclineCompetition(w, r, db)
	}).Methods("GET")
	router.HandleFunc("/decline_competition_start/{user_from}/{user_to}", func(w http.ResponseWriter, r *http.Request) {
		enableCors(w)
		routHandlers.DeclineCompetitionStart(w, r)
	}).Methods("GET")
	router.HandleFunc("/accept_competition/{user1}/{user2}", func(writer http.ResponseWriter, request *http.Request) {
		enableCors(writer)
		routHandlers.GetAcceptCompetitor(writer, request)
	})
	router.HandleFunc("/redirect/{user}", func(writer http.ResponseWriter, request *http.Request) {
		enableCors(writer)
		routHandlers.RedirectToChatroom(writer, request)
	})
	router.HandleFunc("/new_chatroom/{user1}/{user2}", func(writer http.ResponseWriter, request *http.Request) {
		enableCors(writer)
		routHandlers.StartChat(writer, request, db)
	}).Methods("GET")
	router.HandleFunc("/send_message/{user1}/{user2}/{chatID}", func(writer http.ResponseWriter, request *http.Request) {
		enableCors(writer)
		routHandlers.SendMessage(writer, request, db)
	})
	router.HandleFunc("/receive_chat/{user1}/{user2}/{chatID}", func(writer http.ResponseWriter, request *http.Request) {
		enableCors(writer)
		routHandlers.CheckMessages(writer, request, db)
	})
	router.HandleFunc("/new_competition/{chatroomID}/{user}", func(writer http.ResponseWriter, request *http.Request) {
		enableCors(writer)
		routHandlers.CreateCompetition(writer, request, db)
	})
	router.HandleFunc("/add_score/{gameID}/{user}/{score}", func(writer http.ResponseWriter, request *http.Request) {
		enableCors(writer)
		routHandlers.SetUsersScore(writer, request, db)
	})
	router.HandleFunc("/get_result/{gameID}/{user}", func(writer http.ResponseWriter, request *http.Request) {
		enableCors(writer)
		routHandlers.GetCompetitionResult(writer, request, db)
	})
	// Users Links
	router.HandleFunc("/users/register/{username}/{name}/{sex}/{phonenumber}", func(writer http.ResponseWriter, request *http.Request) {
		enableCors(writer)
		if auth.MiddleWareCheckAutch(request) {
			routHandlers.RegisterUser(writer, request, db)
		} else {
			writer.Write([]byte("Auth failed"))
		}
	})
	router.HandleFunc("/users/identify/{username}/{key}", func(writer http.ResponseWriter, request *http.Request) {
		enableCors(writer)
		routHandlers.UsersIdentify(writer, request, db)
	})
	router.HandleFunc("/users/remove/{username}", func(writer http.ResponseWriter, request *http.Request) {
		enableCors(writer)
		if auth.MiddleWareCheckAutch(request) {
			routHandlers.RemoveUser(writer, request, db)
		} else {
			writer.Write([]byte("Auth failed"))
		}
	})
	// XO game
	router.HandleFunc("/xo/init_game/{chatroom_id}", func(writer http.ResponseWriter, request *http.Request) {
		enableCors(writer)
		routHandlers.InitNewXOGame(writer, request, db)
	})
	router.HandleFunc("/xo/move/{user}/{chatroom_id}/{movement}", func(writer http.ResponseWriter, request *http.Request) {
		enableCors(writer)
		routHandlers.XOMove(writer, request, db)
	})
	router.HandleFunc("/xo/get_moves/{user}/{chatroom_id}", func(writer http.ResponseWriter, request *http.Request) {
		enableCors(writer)
		routHandlers.GetXOMoves(writer, request, db)
	})
	// XO END
	// Panel
	router.HandleFunc("/panel/get_all/{offset}", func(writer http.ResponseWriter, request *http.Request) {
		enableCors(writer)
		routHandlers.GetPlayersGames(writer, request, db)
	})
	router.HandleFunc("/panel/get_matches/{offset}", func(writer http.ResponseWriter, request *http.Request) {
		enableCors(writer)
		routHandlers.GetPlayersMatches(writer, request, db)
	})
	router.HandleFunc("/panel/search_games/{phonenumber}", func(writer http.ResponseWriter, request *http.Request) {
		enableCors(writer)
		routHandlers.SearchUsersByPhonenumberInGames(writer, request, db)
	})
	router.HandleFunc("/panel/search_matches/{phonenumber}", func(writer http.ResponseWriter, request *http.Request) {
		enableCors(writer)
		routHandlers.SearchUserByPhonenumberInMatches(writer, request, db)
	})
	router.HandleFunc("/panel/give_award/{username}", func(writer http.ResponseWriter, request *http.Request) {
		enableCors(writer)
		routHandlers.GiveAward(writer, request, db)
	})
	router.HandleFunc("/panel/rewarded_players/", func(writer http.ResponseWriter, request *http.Request) {
		enableCors(writer)
		routHandlers.GetRewardedPlayers(writer, request, db)
	})
	// Panel END

	// ssl
	router.HandleFunc("/.well-known/acme-challenge/E4N0bAE-zp3oi9GvD5xTXYGhz61uEgt7-ycfCLpbaPk", func(writer http.ResponseWriter, request *http.Request) {
		_, _ = writer.Write([]byte("E4N0bAE-zp3oi9GvD5xTXYGhz61uEgt7-ycfCLpbaPk.Ua65auf1IP-HXX6NdYhTPQHZagX7G-wA9NhQsq-Sw74"))
	})
	// ssl END
	log.Logger.Info("Server started successfully!")
	err := http.ListenAndServe("0.0.0.0:4500", router)
	if err != nil {
		fmt.Println(err)
	}
}

func checkRedigoConnection(conn redis.Conn) {
	for i := 0; i < 1; i = i {
		_, err := conn.Do("PING")
		if err != nil {
			conn = redis2.New().Get()
		}
		time.Sleep(time.Second * 3)
	}
}
