package redis

import (
	"encoding/json"
	"flag"
	"github.com/gomodule/redigo/redis"
	"log"
	"time"
)

type competition struct {
	Username   string    `json:"username"`
	Competitor string    `json:"competitor"`
	Time_asked time.Time `json:"time_asked"`
}

func newCompetition(conn redis.Conn, c competition) bool {
	json, err := json.Marshal(c)
	if err != nil {
		return false
	}
	if findCompetition(conn, c.Username).Username != "null" {
		return false
	}
	// SET object
	_, err = conn.Do("SET", c.Username, json)
	if err != nil {
		log.Print(err)
		return false
	}

	return true
}

func findCompetition(conn redis.Conn, username string) competition {
	keys, err := redis.Strings(conn.Do("KEYS", "*"))
	if err != nil {
		return competition{Username: "null"}
	}
	for _, key := range keys {
		comp := getCompetition(conn, key)
		if comp.Competitor == username && time.Now().Sub(comp.Time_asked).Minutes() < 2.0 {
			return comp
		}
	}
	return competition{Username: "null"}
}

func getCompetition(c redis.Conn, username string) competition {
	s, err := redis.String(c.Do("GET", username))
	if err == redis.ErrNil {
		return competition{Username: "null"}
	}
	comp := competition{}
	err = json.Unmarshal([]byte(s), &comp)
	return comp

}

func New() *redis.Pool {
	return &redis.Pool{
		MaxIdle: 100,
		// Dial is an application supplied function for creating and
		// configuring a connection.
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", *flag.String("redisServer", ":6379", ""))
			if err != nil {
				panic(err.Error())
			}
			if _, err := c.Do("AUTH", "@Ali@021021"); err != nil {
				c.Close()
				return nil, err
			}
			return c, err
		}, TestOnBorrow: func(c redis.Conn, t time.Time) error {
			_, err := c.Do("PING")
			if err != nil {
				log.Println(err)
			}
			return err
		},
	}
}
