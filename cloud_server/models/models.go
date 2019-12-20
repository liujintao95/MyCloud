package models

import "database/sql"

type UserInfo struct {
	Id       int64          `db:"ui_id"`
	Name     string         `db:"ui_name"`
	User     string         `db:"ui_user"`
	Pwd      string         `db:"ui_pwd"`
	Level    string         `db:"ui_level"`
	Email    string         `db:"ui_email"`
	Phone    sql.NullString `db:"ui_phone"`
	Recycled string         `db:"ui_recycled"`
}

type FileInfo struct {
	Id       int64          `db:"fi_id"`
	Name     string         `db:"fi_name"`
	Hash     string         `db:"fi_hash"`
	Size     int64          `db:"fi_size"`
	Path     string         `db:"fi_path"`
	IsPublic int64          `db:"fi_is_public"`
	Remark   sql.NullString `db:"fi_remark"`
	Recycled string         `db:"fi_recycled"`
}

type UserFileMap struct {
	Id       int64          `db:"uf_id"`
	UserInfo UserInfo       `db:"uf_ui_id"`
	FileInfo FileInfo       `db:"uf_fi_id"`
	FileName string         `db:"uf_file_name"`
	Star     int64          `db:"uf_star"`
	IsPublic int64          `db:"uf_is_public"`
	State    sql.NullInt64  `db:"uf_remark"`
	Remark   sql.NullString `db:"uf_state"`
	Recycled string         `db:"uf_recycled"`
}
