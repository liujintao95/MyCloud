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
	IsPublic int            `db:"fi_is_public"`
	State    int            `db:"fi_state"`
	Remark   sql.NullString `db:"fi_remark"`
	Recycled string         `db:"fi_recycled"`
}

type UserFileMap struct {
	Id       int64          `db:"uf_id"`
	UserInfo UserInfo       `db:"uf_ui_id"`
	FileInfo FileInfo       `db:"uf_fi_id"`
	FileName string         `db:"uf_file_name"`
	Star     int            `db:"uf_star"`
	IsPublic int            `db:"uf_is_public"`
	State    int            `db:"uf_state"`
	Remark   sql.NullString `db:"uf_remark"`
	Recycled string         `db:"uf_recycled"`
}

type FileBlockInfo struct {
	Id         int64    `db:"fbi_id"`
	Hash       string   `db:"fbi_hash"`
	FileName   string   `db:"fbi_file_name"`
	UserInfo   UserInfo `db:"fbi_ui_id"`
	UploadID   string   `db:"fbi_upload_id"`
	FileSize   int64    `db:"fbi_file_size"`
	BlockSize  int64    `db:"fbi_block_size"`
	BlockCount int      `db:"fbi_block_count"`
	State      int      `db:"fbi_state"`
	Recycled   string   `db:"fbi_recycled"`
}

type BlockInfo struct {
	Id            int64         `db:"bi_id"`
	FileBlockInfo FileBlockInfo `db:"bi_upload_id"`
	Index         int           `db:"bi_index"`
	Path          string        `db:"bi_path"`
	Size          int64         `db:"bi_size"`
	State         int           `db:"bi_state"`
	Recycled      string        `db:"bi_recycled"`
}

type FileDirectory struct {
	Id          int64          `db:"fd_id"`
	UserFileMap UserFileMap    `db:"fd_uf_id"`
	IsDir       int            `db:"fd_is_dir"`
	DirName     string `db:"fd_dir_name"`
	Fid         int64          `db:"fd_fid"`
	Recycled    string         `db:"fd_recycled"`
}
