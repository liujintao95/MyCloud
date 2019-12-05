package conf

var RedisConf = map[string]string{
	"name":    "redis",
	"type":    "tcp",
	"address": "127.0.0.1:6379",
	"auth":    "123456",
}

var MySqlConf = map[string]string{
	"user":     "root",
	"pwd":      "123456",
	"type":     "tcp",
	"address":  "127.0.0.1",
	"port":     "3306",
	"database": "blog",
}
