package utils

import (
	"MyCloud/conf"
	"database/sql"
	"fmt"
)

var Conn *sql.DB

func init() {
	dsn := fmt.Sprintf("%s:%s@%s(%s:%d)/%s",
		conf.MySqlConf["user"],
		conf.MySqlConf["pwd"],
		conf.MySqlConf["type"],
		conf.MySqlConf["address"],
		conf.MySqlConf["port"],
		conf.MySqlConf["database"],
	)
	Conn, _ = sql.Open("mysql", dsn)
	Conn.SetMaxOpenConns(15)
	Conn.SetMaxIdleConns(5)
	_ = Conn.Ping()
}
