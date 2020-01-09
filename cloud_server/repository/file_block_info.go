package repository

import (
	"MyCloud/cloud_server/models"
	"MyCloud/conf"
	"MyCloud/utils"
	"encoding/json"
	"github.com/garyburd/redigo/redis"
)

type IFileBlock interface {
	GetByUploadId(string) (models.FileBlockInfo, error)
	Set(models.FileBlockInfo) (int64, error)
	Update(models.FileBlockInfo) error
	DeleteByUploadId(string) error

	GetSqlByUploadId(string) (models.FileBlockInfo, error)
	SetSql(models.FileBlockInfo) (int64, error)
	UpdateSql(models.FileBlockInfo) error
	DelSqlByUploadId(string) error

	GetCache(string) (models.FileBlockInfo, error)
	SetCache(string, models.FileBlockInfo) error
	DelCache(string) error
}

type FileBlockManager struct {
	table string
}

func NewFileBlockManager() IFileBlock {
	return &FileBlockManager{table: "file_block_info"}
}

func (f FileBlockManager) GetByUploadId(uploadId string) (models.FileBlockInfo, error) {
	fileBlockMate, err := f.GetCache(uploadId)
	if err != nil{
		fileBlockMate, err = f.GetSqlByUploadId(uploadId)
		if err == nil {
			err = f.SetCache("fb_"+uploadId, fileBlockMate)
		}
	}
	return fileBlockMate, err
}

func (f FileBlockManager) Set(fileBlockMate models.FileBlockInfo) (int64, error) {
	id, err := f.SetSql(fileBlockMate)
	if err != nil {
		return -1, err
	}
	fileBlockMate.Id = id

	err = f.SetCache("fb_"+fileBlockMate.UploadID, fileBlockMate)
	return id, err
}

func (f FileBlockManager) Update(fileBlockMate models.FileBlockInfo) error {
	err := f.UpdateSql(fileBlockMate)
	if err != nil {
		return err
	}
	key := "fb_" + fileBlockMate.UploadID
	err = f.SetCache(key, fileBlockMate)
	return err
}

func (f FileBlockManager) DeleteByUploadId(uploadId string) error {
	err := f.DelSqlByUploadId(uploadId)
	if err != nil {
		return err
	}
	err = f.DelCache("fb_" + uploadId)
	return err
}

func (f FileBlockManager) GetSqlByUploadId(uploadId string) (models.FileBlockInfo, error) {
	var fileBlockMate models.FileBlockInfo

	getSql := `
		SELECT fbi_id, fbi_hash, fbi_file_name, 
		fbi_upload_id, fbi_file_size, fbi_block_size, 
		fbi_block_count, fbi_state, fbi_recycled,
		ui_id, ui_name, ui_user, ui_pwd, ui_level,
		ui_email, ui_phone, ui_recycled
		FROM file_block_info
		INNER JOIN user_info
		ON fbi_ui_id = ui_id
		WHERE fbi_upload_id = ?
		AND fbi_recycled == 'N'
	`
	rows := utils.Conn.QueryRow(getSql, uploadId)
	err := rows.Scan(
		&fileBlockMate.Id, &fileBlockMate.Hash, &fileBlockMate.FileName,
		&fileBlockMate.UploadID, &fileBlockMate.FileSize, &fileBlockMate.BlockSize,
		&fileBlockMate.BlockCount, &fileBlockMate.State, &fileBlockMate.Recycled,
		&fileBlockMate.UserInfo.Id, &fileBlockMate.UserInfo.Name,
		&fileBlockMate.UserInfo.User, &fileBlockMate.UserInfo.Pwd,
		&fileBlockMate.UserInfo.Level, &fileBlockMate.UserInfo.Email,
		&fileBlockMate.UserInfo.Phone, &fileBlockMate.UserInfo.Recycled,
	)
	return fileBlockMate, err
}

func (f FileBlockManager) SetSql(fileBlockMate models.FileBlockInfo) (int64, error) {
	insertSql := `
		INSERT INTO file_block_info(
			fbi_hash, fbi_file_name, fbi_ui_id, fbi_upload_id,
			fbi_file_size, fbi_block_size, fbi_block_count
		) 
		VALUES (?,?,?,?,?,?,?)
	`
	res, err := utils.Conn.Exec(
		insertSql,
		fileBlockMate.Hash, fileBlockMate.FileName,
		fileBlockMate.UserInfo.Id, fileBlockMate.UploadID,
		fileBlockMate.FileSize, fileBlockMate.BlockSize,
		fileBlockMate.BlockCount,
	)
	if err != nil {
		return -1, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return -1, err
	}
	return id, err
}

func (f FileBlockManager) UpdateSql(fileBlockMate models.FileBlockInfo) error {
	updateSql := `
		UPDATE file_block_info 
		SET fbi_file_name=?, fbi_file_size=?, fbi_block_size=?,
		fbi_block_count=?, fbi_state=?, fbi_recycled=?
		WHERE fbi_upload_id=?
	`
	_, err := utils.Conn.Exec(
		updateSql,
		fileBlockMate.FileName, fileBlockMate.FileSize,
		fileBlockMate.BlockSize, fileBlockMate.BlockCount,
		fileBlockMate.State, fileBlockMate.Recycled,
		fileBlockMate.UploadID,
	)
	return err
}

func (f FileBlockManager) DelSqlByUploadId(uploadId string) error {
	updateSql := `
		UPDATE file_block_info 
		SET fbi_recycled = 'Y'
		WHERE fbi_upload_id=?
	`
	_, err := utils.Conn.Exec(updateSql, uploadId)
	return err
}

func (f FileBlockManager) GetCache(key string) (models.FileBlockInfo, error) {
	rc := utils.RedisPool.Get()
	defer rc.Close()

	jsonData, err := redis.Bytes(rc.Do("GET", key))
	fileBlockMate := models.FileBlockInfo{}
	if jsonData != nil {
		_ = json.Unmarshal(jsonData, &fileBlockMate)
	}
	return fileBlockMate, err
}

func (f FileBlockManager) SetCache(key string, fileBlockMate models.FileBlockInfo) error {
	jsonData, err := json.Marshal(fileBlockMate)

	rc := utils.RedisPool.Get()
	defer rc.Close()

	_, err = rc.Do("SET", key, string(jsonData), "EX", conf.REDIS_MAXAGE)
	return err
}

func (f FileBlockManager) DelCache(key string) error {
	rc := utils.RedisPool.Get()
	defer rc.Close()
	_, err := rc.Do("DEL", key)
	return err
}
