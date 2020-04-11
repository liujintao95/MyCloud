package repository

import (
	"MyCloud/cloud_server/models"
	"MyCloud/conf"
	"MyCloud/utils"
	"encoding/json"

	"github.com/garyburd/redigo/redis"
)

//type IFile interface {
//	GetByHash(string) (models.FileInfo, error)
//	Set(models.FileInfo) (int64, error)
//	Update(models.FileInfo) error
//	DeleteByHash(string) error
//
//	GetSqlByHash(string) (models.FileInfo, error)
//	SetSql(models.FileInfo) (int64, error)
//	UpdateSql(models.FileInfo) error
//	DelSqlByHash(string) error
//
//	GetCache(string) (models.FileInfo, error)
//	SetCache(string, models.FileInfo) error
//	DelCache(string) error
//}

type FileManager struct {
	table string
}

func NewFileManager() *FileManager {
	return &FileManager{table: "file_info"}
}

func (f *FileManager) GetByHash(hash string) (models.FileInfo, error) {
	fileMate, err := f.GetCache(hash)
	if err != nil {
		fileMate, err = f.GetSqlByHash(hash)
		if err == nil {
			err = f.SetCache(fileMate.Hash, fileMate)
		}
	}
	return fileMate, err
}

func (f *FileManager) Set(fileMate models.FileInfo) (int64, error) {
	id, err := f.SetSql(fileMate)
	fileMate.Id = id
	if err != nil {
		return -1, err
	}
	err = f.SetCache(fileMate.Hash, fileMate)
	return id, err
}

func (f *FileManager) Update(fileMate models.FileInfo) error {
	err := f.UpdateSql(fileMate)
	if err != nil {
		return err
	}
	err = f.SetCache(fileMate.Hash, fileMate)
	return err
}

func (f *FileManager) DeleteByHash(hash string) error {
	err := f.DelSqlByHash(hash)
	if err != nil {
		return err
	}
	err = f.DelCache(hash)
	return err
}

func (f *FileManager) GetSqlByHash(hash string) (models.FileInfo, error) {
	var fileMate models.FileInfo
	getSql := `
		SELECT fi_id, fi_name, fi_hash, fi_size,
		fi_path, fi_state, fi_remark, fi_is_public,
		fi_recycled
		FROM file_info 
		WHERE fi_hash = ?
		AND fi_recycled = 'N'
	`
	rows := utils.Conn.QueryRow(getSql, hash)
	err := rows.Scan(
		&fileMate.Id, &fileMate.Name, &fileMate.Hash,
		&fileMate.Size, &fileMate.Path, &fileMate.State,
		&fileMate.Remark, &fileMate.IsPublic,
		&fileMate.Recycled,
	)
	return fileMate, err
}

func (f *FileManager) SetSql(fileMate models.FileInfo) (int64, error) {
	insertSql := `
		INSERT INTO file_info(
			fi_name, fi_hash, fi_size,
			fi_path, fi_remark
		) 
		VALUES (?,?,?,?,?)
	`
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

func (f *FileManager) UpdateSql(fileMate models.FileInfo) error {
	updateSql := `
		UPDATE file_info 
		SET fi_name=?, fi_size=?, fi_path=?, fi_state=?
		fi_remark=?, fi_is_public=?, fi_recycled=?
		WHERE fi_hash=?
	`
	_, err := utils.Conn.Exec(
		updateSql,
		fileMate.Name, fileMate.Size, fileMate.Path,
		fileMate.State, fileMate.Remark, fileMate.IsPublic,
		fileMate.Recycled, fileMate.Hash,
	)
	return err
}

func (f *FileManager) DelSqlByHash(hash string) error {
	updateSql := `
		UPDATE file_info 
		SET fi_recycled = 'Y'
		WHERE fi_hash=?
	`
	_, err := utils.Conn.Exec(updateSql, hash)
	return err
}

func (f *FileManager) GetCache(key string) (models.FileInfo, error) {
	rc := utils.RedisPool.Get()
	defer rc.Close()

	jsonData, err := redis.Bytes(rc.Do("GET", key))
	fileMate := models.FileInfo{}
	if jsonData != nil {
		_ = json.Unmarshal(jsonData, &fileMate)
	}
	return fileMate, err
}

func (f *FileManager) SetCache(key string, fileMate models.FileInfo) error {
	jsonData, err := json.Marshal(fileMate)

	rc := utils.RedisPool.Get()
	defer rc.Close()

	_, err = rc.Do("SET", key, string(jsonData), "EX", conf.REDIS_MAXAGE)
	return err
}

func (f *FileManager) DelCache(key string) error {
	rc := utils.RedisPool.Get()
	defer rc.Close()
	_, err := rc.Do("DEL", key)
	return err
}
