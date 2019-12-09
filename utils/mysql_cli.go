package utils

import (
	"MyCloud/conf"
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
)

var Conn *sql.DB

func MySqlInit() {
	dsn := fmt.Sprintf("%s:%s@%s(%s:%s)/%s",
		conf.MySqlConf["user"],
		conf.MySqlConf["pwd"],
		conf.MySqlConf["type"],
		conf.MySqlConf["address"],
		conf.MySqlConf["port"],
		conf.MySqlConf["database"],
	)
	Conn, _ = sql.Open("mysql", dsn)
	Conn.SetMaxOpenConns(50)
	Conn.SetMaxIdleConns(10)
	_ = Conn.Ping()
}
