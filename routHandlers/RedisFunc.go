package routHandlers

import (
	"encoding/json"
	"log"
	"time"

	"github.com/aicam/game_server/internal/redisconn"
	"github.com/gomodule/redigo/redis"
)

type competition struct {
	Username           string    `json:"username"`
	Competitor         string    `json:"competitor"`
	Time_asked         time.Time `json:"time_asked"`
	CompetitorAccepted bool      `json:"competitor_accepted"`
}

func newCompetition(c competition) int {
	conn := redisconn.Redis().Get()
	json, err := json.Marshal(c)
	if err != nil {
		return 0
	}
	comp := getCompetition(c.Username)
	if comp.Username != "null" {
		return 1 // user has requested under 10s ago
	}
	comp = findCompetition(c.Competitor)
	if comp.CompetitorAccepted {
		return 2 // busy
	}
	comp = getCompetition(c.Competitor)
	if comp.CompetitorAccepted {
		return 2
	}
	// SET object
	conn.Send("SET", c.Username, json)
	conn.Flush()
	conn.Receive()
	if err != nil {
		return 0
	}
	conn.Send("EXPIRE", c.Username, 15)
	conn.Flush()
	conn.Receive()
	return 4
}

func findCompetition(username string) competition {
	conn := redisconn.Redis().Get()
	conn.Send("KEYS", "*")
	conn.Flush()
	re, erre := conn.Receive()
	keys, err := redis.Strings(re, erre)
	if err != nil {
		log.Print(err)
		return competition{Username: "null"}
	}
	for _, key := range keys {
		comp := getCompetition(key)
		if comp.Competitor == username {
			return comp
		}
	}
	return competition{Username: "null"}
}

func getCompetition(username string) competition {
	c := redisconn.Redis().Get()
	c.Send("GET", username)
	c.Flush()
	rv, er := c.Receive()
	s, err := redis.String(rv, er)
	if er != nil {
		log.Print(err)
	}
	if err == redis.ErrNil {
		return competition{Username: "null"}
	}
	comp := competition{}
	err = json.Unmarshal([]byte(s), &comp)
	return comp
}

func GetAcceptCompetitorRedis(userFrom string, userTo string) bool {
	conn := redisconn.Redis().Get()
	conn.Send("KEYS", "*")
	conn.Flush()
	comp := competition{}
	re, erre := conn.Receive()
	keys, err := redis.Strings(re, erre)
	if err != nil {
		return false
	}
	for _, key := range keys {
		comp = getCompetition(key)
		if comp.Username != userFrom || comp.Competitor != userTo {
			comp = competition{
				Username: "null",
			}
		} else {
			break
		}
	}
	if comp.Username == "null" {
		return false
	}
	comp.CompetitorAccepted = true
	js, _ := json.Marshal(comp)
	conn.Send("SET", comp.Username, js)
	conn.Flush()
	conn.Receive()
	conn.Send("EXPIRE", comp.Username, 240)
	conn.Flush()
	conn.Receive()
	return true
}

type RedirectToChatroomRedisOutput struct {
	UserFrom string `json:"user_from"`
	UserTo   string `json:"user_to"`
}

func RedirectToChatroomRedis(username string) ([]byte, error) {
	comp := getCompetition(username)
	if comp.CompetitorAccepted {
		return json.Marshal(RedirectToChatroomRedisOutput{
			UserFrom: comp.Username,
			UserTo:   comp.Competitor,
		})
	}
	comp = findCompetition(username)
	if comp.CompetitorAccepted {
		return json.Marshal(RedirectToChatroomRedisOutput{
			UserFrom: comp.Username,
			UserTo:   comp.Competitor,
		})
	}
	return json.Marshal(RedirectToChatroomRedisOutput{
		UserFrom: "null",
		UserTo:   "null",
	})
}

func DeclineCompetitionRedis(username string) {
	comp := findCompetition(username)
	ClearCompetetion(comp.Username)
}

func ClearCompetetion(username string) bool {
	conn := redisconn.Redis().Get()
	conn.Send("DEL", username)
	conn.Flush()
	conn.Receive()
	return true
}
