package repository

import (
	"MyCloud/cloud_server/models"
	"MyCloud/conf"
	"MyCloud/utils"
	"encoding/json"
	"github.com/garyburd/redigo/redis"
)

type IFile interface {
	SelectByHash(string) (models.FileInfo, error)
	Insert(models.FileInfo) (int64, error)
	Update(models.FileInfo) error
	Delete(string) error
	GetCache(string) (models.FileInfo, error)
	SetCache(string, models.FileInfo) error
	DelCache(string) error
}

type FileManager struct {
	table string
}

func NewFileManager() IFile {
	return &FileManager{table: "file_info"}
}

func (f *FileManager) SelectByHash(hash string) (models.FileInfo, error) {
	fileMate := new(models.FileInfo)
	getSql := `
		SELECT fi_id, fi_name, fi_hash, fi_size,
		fi_path, fi_remark, fi_state, fi_recycled
		FROM file_info 
		WHERE fi_hash = ?
		AND fi_recycled == 'N'
	`
	rows := utils.Conn.QueryRow(getSql, hash)
	err := rows.Scan(
		fileMate.Id, fileMate.Name, fileMate.Hash,
		fileMate.Size, fileMate.Path, fileMate.Remark,
		fileMate.State, fileMate.Recycled,
	)
	return *fileMate, err
}

func (f *FileManager) Insert(fileMate models.FileInfo) (int64, error) {
	insertSql := `
		INSERT INTO file_info(
			fi_name, fi_hash, fi_size,
			fi_path, fi_remark
		) 
		VALUES (?,?,?,?,?)`
	res, err := utils.Conn.Exec(
		insertSql,
		fileMate.Name, fileMate.Hash, fileMate.Size,
		fileMate.Path, fileMate.Remark,
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

func (f *FileManager) Update(fileMate models.FileInfo) error {
	updateSql := `
		UPDATE file_info 
		SET fi_name=?, fi_size=?, fi_path=?, fi_remark=?, fi_state=?
		WHERE fi_hash=?
	`
	_, err := utils.Conn.Exec(
		updateSql,
		fileMate.Name, fileMate.Size, fileMate.Path,
		fileMate.Remark, fileMate.State, fileMate.Hash,
	)
	return err
}

func (f *FileManager) Delete(hash string) error {
	updateSql := `
		UPDATE file_info 
		SET fi_recycled = 'Y'
		WHERE fi_hash=?
	`
	_, err := utils.Conn.Exec(updateSql, hash)
	return err
}

func (f *FileManager) GetCache(hash string) (models.FileInfo, error) {
	rc := utils.RedisPool.Get()
	defer rc.Close()

	jsonData, err := redis.Bytes(rc.Do("LRANGE", hash, 0, -1))
	fileMate := models.FileInfo{}
	if jsonData != nil {
		_ = json.Unmarshal(jsonData, fileMate)
	}
	return fileMate, err
}

func (f *FileManager) SetCache(hash string, obj models.FileInfo) error {
	jsonData, err := json.Marshal(obj)

	rc := utils.RedisPool.Get()
	defer rc.Close()

	_, err = rc.Do("LPUSH", hash, string(jsonData), "EX", string(conf.REDIS_MAXAGE))
	return err
}

func (f *FileManager) DelCache(key string) error {
	rc := utils.RedisPool.Get()
	defer rc.Close()
	_, err := rc.Do("DEL", key)
	return err
}