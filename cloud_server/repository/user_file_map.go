package repository

import (
	"MyCloud/cloud_server/models"
	"MyCloud/conf"
	"MyCloud/utils"
	"encoding/json"
	"github.com/garyburd/redigo/redis"
)

type IUserFileMap interface {
	SelectByUser(string) ([]models.UserFileMap, error)
	SelectByFile(string) ([]models.UserFileMap, error)
	SelectByUserFile(string, string) (models.UserFileMap, error)
	Insert(models.UserFileMap) (int64, error)
	Update(models.UserFileMap) error
	Delete(models.UserFileMap) error
	GetCache(string) (models.UserFileMap, error)
	SetCache(string, models.UserFileMap) error
	DelCache(string) error
}

type UserFileManager struct {
	table string
}

func NewUserFileManager() IUserFileMap {
	return &UserFileManager{table: "user_file_map"}
}

func (u UserFileManager) SelectByUser(string) ([]models.UserFileMap, error) {
	panic("implement me")
}

func (u UserFileManager) SelectByFile(fileHash string) ([]models.UserFileMap, error) {
	var mapList []models.UserFileMap

	getSql := `
		SELECT uf_id, uf_file_name, uf_star,
		uf_is_public, uf_remark, uf_state, uf_recycled,
		ui_id, ui_name, ui_user, ui_pwd,
		ui_level, ui_email, ui_phone, ui_recycled, 
		fi_id, fi_name, fi_hash, fi_size,
		fi_path, fi_remark, fi_is_public, fi_recycled
		FROM user_file_map
		INNER JOIN user_info
		ON uf_ui_id = ui_id
		INNER JOIN file_info
		ON uf_fi_id = fi_id
		WHERE fi_hash = ?
		AND uf_recycled == 'N'
		AND ui_recycled == 'N'
		AND fi_recycled == 'N'
	`
	rows,err := utils.Conn.Query(getSql, fileHash)
	if err != nil{
		return mapList, err
	}

	for rows.Next() {
		userFileMate := new(models.UserFileMap)
		_ := rows.Scan(
			userFileMate.Id, userFileMate.FileName, userFileMate.Star,
			userFileMate.IsPublic, userFileMate.Remark, userFileMate.State,
			userFileMate.Recycled, userFileMate.UserInfo.Id,
			userFileMate.UserInfo.Name, userFileMate.UserInfo.User,
			userFileMate.UserInfo.Pwd, userFileMate.UserInfo.Level,
			userFileMate.UserInfo.Email, userFileMate.UserInfo.Phone,
			userFileMate.UserInfo.Recycled, userFileMate.FileInfo.Id,
			userFileMate.FileInfo.Name, userFileMate.FileInfo.Hash,
			userFileMate.FileInfo.Size, userFileMate.FileInfo.Path,
			userFileMate.FileInfo.Remark, userFileMate.FileInfo.IsPublic,
			userFileMate.FileInfo.Recycled,
		)
		mapList = append(mapList, *userFileMate)
	}
	return mapList, err
}

func (u UserFileManager) SelectByUserFile(user string, fileHash string) (models.UserFileMap, error) {
	userFileMate := new(models.UserFileMap)
	getSql := `
		SELECT uf_id, uf_file_name, uf_star,
		uf_is_public, uf_remark, uf_state, uf_recycled,
		ui_id, ui_name, ui_user, ui_pwd,
		ui_level, ui_email, ui_phone, ui_recycled, 
		fi_id, fi_name, fi_hash, fi_size,
		fi_path, fi_remark, fi_is_public, fi_recycled
		FROM user_file_map
		INNER JOIN user_info
		ON uf_ui_id = ui_id
		INNER JOIN file_info
		ON uf_fi_id = fi_id
		WHERE ui_user = ?
		AND fi_hash = ?
		AND uf_recycled == 'N'
		AND ui_recycled == 'N'
		AND fi_recycled == 'N'
	`
	rows := utils.Conn.QueryRow(getSql, user, fileHash)
	err := rows.Scan(
		userFileMate.Id, userFileMate.FileName, userFileMate.Star,
		userFileMate.IsPublic, userFileMate.Remark, userFileMate.State,
		userFileMate.Recycled, userFileMate.UserInfo.Id,
		userFileMate.UserInfo.Name, userFileMate.UserInfo.User,
		userFileMate.UserInfo.Pwd, userFileMate.UserInfo.Level,
		userFileMate.UserInfo.Email, userFileMate.UserInfo.Phone,
		userFileMate.UserInfo.Recycled, userFileMate.FileInfo.Id,
		userFileMate.FileInfo.Name, userFileMate.FileInfo.Hash,
		userFileMate.FileInfo.Size, userFileMate.FileInfo.Path,
		userFileMate.FileInfo.Remark, userFileMate.FileInfo.IsPublic,
		userFileMate.FileInfo.Recycled,
	)
	return *userFileMate, err
}

func (u UserFileManager) Insert(userFileMate models.UserFileMap) (int64, error) {
	insertSql := `
		INSERT INTO user_file_map(
			uf_ui_id, uf_fi_id, uf_file_name,
			uf_remark
		) 
		VALUES (?,?,?,?)`
	res, err := utils.Conn.Exec(
		insertSql,
		userFileMate.UserInfo.Id, userFileMate.FileInfo.Id,
		userFileMate.FileName, userFileMate.Remark,
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

func (u UserFileManager) Update(userFileMate models.UserFileMap) error {
	updateSql := `
		UPDATE file_info 
		SET uf_file_name=?, uf_star=?,
		uf_is_public=?, uf_remark=?, uf_state=?
		WHERE uf_ui_id=?
		AND uf_fi_id=?
	`
	_, err := utils.Conn.Exec(
		updateSql,
		userFileMate.FileName, userFileMate.Star, userFileMate.IsPublic,
		userFileMate.Remark, userFileMate.State,
		userFileMate.UserInfo.Id, userFileMate.FileInfo.Id,
	)
	return err
}

func (u UserFileManager) Delete(userFileMate models.UserFileMap) error {
	updateSql := `
		UPDATE user_file_map 
		SET uf_recycled = 'Y'
		WHERE uf_ui_id=?
		AND uf_fi_id=?
	`
	_, err := utils.Conn.Exec(updateSql,
		userFileMate.UserInfo.Id, userFileMate.FileInfo.Id)
	return err
}

func (u UserFileManager) GetCache(key string) (models.UserFileMap, error) {
	rc := utils.RedisPool.Get()
	defer rc.Close()

	jsonData, err := redis.Bytes(rc.Do("LRANGE", key, 0, -1))
	userFileMate := models.UserFileMap{}
	if jsonData != nil {
		_ = json.Unmarshal(jsonData, userFileMate)
	}
	return userFileMate, err
}

func (u UserFileManager) SetCache(key string, userFileMate models.UserFileMap) error {
	jsonData, err := json.Marshal(userFileMate)

	rc := utils.RedisPool.Get()
	defer rc.Close()

	_, err = rc.Do("LPUSH", key, string(jsonData), "EX", string(conf.REDIS_MAXAGE))
	return err
}

func (u UserFileManager) DelCache(string) error {
	panic("implement me")
}
