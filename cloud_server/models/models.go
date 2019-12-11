package models

import "database/sql"

type UserInfo struct {
	Id       int64          `db:"ui_id"`
	User     string         `db:"ui_user"`
	Pwd      string         `db:"ui_pwd"`
	Level    string         `db:"ui_level"`
	Email    string         `db:"ui_email"`
	Phone    sql.NullString `db:"ui_phone"`
	Recycled string         `db:"ui_recycled"`
}
