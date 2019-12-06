package server


type UserInfo struct {
	Id int64 `db:"ui_id"`
	User string  `db:"ui_user"`
	Pwd string `db:"ui_pwd"`
	Level string `db:"ui_level"`
	Email string `db:"ui_email"`
	Phone string `db:"ui_phone"`
}