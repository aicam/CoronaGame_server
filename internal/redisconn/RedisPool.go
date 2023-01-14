package redisconn

import (
	"fmt"
	"sync"

	"github.com/gomodule/redigo/redis"
)

var doOnce sync.Once
var redisPool *redis.Pool

func Redis() *redis.Pool {
	doOnce.Do(func() {
		redisPool = &redis.Pool{
			Dial: func() (redis.Conn, error) {
				c, err := redis.Dial("tcp", "13.56.251.61:6379")
				if err != nil {
					panic(err.Error())
				}
				if _, err := c.Do("AUTH", "021021"); err != nil {
					c.Close()
					return nil, err
				}
				return c, err
			},
		}
		fmt.Println("Redis pool created...")
	})
	return redisPool
}
