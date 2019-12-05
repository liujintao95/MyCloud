package utils

import (
	"MyCloud/conf"
	"github.com/garyburd/redigo/redis"
	"time"
)

var RedisClient *redis.Pool

func RedisInit() {
	// 建立连接池
	RedisClient = &redis.Pool{
		MaxIdle:     5,
		MaxActive:   0,
		IdleTimeout: 300 * time.Second,
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial(conf.RedisConf["type"], conf.RedisConf["address"])
			if err != nil {
				return nil, err
			}
			if _, err := c.Do("AUTH", conf.RedisConf["auth"]); err != nil {
				_ = c.Close()
				return nil, err
			}
			return c, nil
		},
	}
}
