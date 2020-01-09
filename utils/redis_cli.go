package utils

import (
	"MyCloud/conf"
	"fmt"
	"github.com/garyburd/redigo/redis"
	"time"
)

var RedisPool *redis.Pool

func RedisInit() {
	url := fmt.Sprintf("redis://%s", conf.RedisConf["address"])
	// 建立连接池
	RedisPool = &redis.Pool{
		MaxIdle:     5,
		IdleTimeout: 300 * time.Second,
		Dial: func() (redis.Conn, error) {
			c, err := redis.DialURL(url)
			if err != nil {
				Logging.Error(err)
				return nil, fmt.Errorf("redis connection error: %s", err)
			}
			//if _, err := c.Do("AUTH", conf.RedisConf["auth"]); err != nil {
			//	_ = c.Close()
			//	Logging.Error(err)
			//	return nil, fmt.Errorf("redis auth password error: %s", err)
			//}
			return c, nil
		},
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			_, err := c.Do("PING")
			if err != nil {
				Logging.Error(err)
				return fmt.Errorf("ping redis error: %s", err)
			}
			return nil
		},
	}
}
