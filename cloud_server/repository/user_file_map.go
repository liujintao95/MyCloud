package repository

import (
	"MyCloud/cloud_server/models"
	"MyCloud/conf"
	"MyCloud/utils"
	"encoding/json"

	"github.com/garyburd/redigo/redis"
)

type IUserFileMap interface {
	GetByUser(string, int) ([]models.UserFileMap, error)
	GetByFile(string, int) ([]models.UserFileMap, error)
	GetByUserFile(string, string) (models.UserFileMap, error)
	Set(models.UserFileMap) (int64, error)
	Update(models.UserFileMap) error
	DeleteByUserFile(string, string) error

	GetCountByUser(string) (int, error)
	GetSqlByUser(string) ([]models.UserFileMap, error)
	GetSqlByFile(string) ([]models.UserFileMap, error)
	GetSqlByUserFile(string, string) (models.UserFileMap, error)
	SetSql(models.UserFileMap) (int64, error)
	UpdateSql(models.UserFileMap) error
	DelSqlByUserFile(string, string) error

	GetCache(string) (models.UserFileMap, error)
	GetCacheList(string) ([]models.UserFileMap, error)
	SetCache(string, models.UserFileMap) error
	SetCacheList(string, []models.UserFileMap) error
	DelCache(string) error
}

type UserFileManager struct {
	table string
}

func NewUserFileManager() IUserFileMap {
	return &UserFileManager{table: "user_file_map"}
}

func (u UserFileManager) GetByUser(user string, page int) ([]models.UserFileMap, error) {
	userFileList, err := u.GetCacheList("UserFileList_" + user)
	if err != nil || userFileList == nil {
		userFileList, err = u.GetSqlByUser(user)
		if err == nil {
			err = u.SetCacheList("UserFileList_"+user, userFileList)
		}
	}

	if len(userFileList) >= page*10 {
		return userFileList[(page-1)*10 : page*10], err
	} else {
		return userFileList[(page-1)*10:], err
	}
}

func (u UserFileManager) GetByFile(fileHash string, page int) ([]models.UserFileMap, error) {
	userFileList, err := u.GetCacheList("UserFileList_" + fileHash)
	if err != nil || userFileList == nil {
		userFileList, err = u.GetSqlByFile(fileHash)
		if err == nil {
			err = u.SetCacheList("UserFileList_"+fileHash, userFileList)
		}
	}

	if len(userFileList) >= page*10 {
		return userFileList[(page-1)*10 : page*10], err
	} else {
		return userFileList[(page-1)*10:], err
	}
}

func (u UserFileManager) GetByUserFile(user string, fileHash string) (models.UserFileMap, error) {
	userFileMate, err := u.GetCache(user + fileHash)
	if err != nil || userFileMate.Recycled == "" {
		userFileMate, err = u.GetSqlByUserFile(user, fileHash)
		if err == nil {
			err = u.SetCache(user+fileHash, userFileMate)
		}
	}
	return userFileMate, err
}

func (u UserFileManager) Set(userFileMate models.UserFileMap) (int64, error) {
	id, err := u.SetSql(userFileMate)
	if err != nil {
		return -1, err
	}
	key := userFileMate.UserInfo.User + userFileMate.FileInfo.Hash
	err = u.SetCache(key, userFileMate)

	userFileList, err := u.GetCacheList("UserFileList_" + userFileMate.UserInfo.User)
	userFileList = append(userFileList, userFileMate)
	err = u.SetCacheList("UserFileList_"+userFileMate.UserInfo.User, userFileList)

	userFileList, err = u.GetCacheList("UserFileList_" + userFileMate.FileInfo.Hash)
	userFileList = append(userFileList, userFileMate)
	err = u.SetCacheList("UserFileList_"+userFileMate.FileInfo.Hash, userFileList)

	return id, err
}

func (u UserFileManager) Update(userFileMate models.UserFileMap) error {
	err := u.UpdateSql(userFileMate)
	if err != nil {
		return err
	}
	key := userFileMate.UserInfo.User + userFileMate.FileInfo.Hash
	err = u.SetCache(key, userFileMate)

	userFileList, err := u.GetCacheList("UserFileList_" + userFileMate.UserInfo.User)
	for index, val := range userFileList {
		if val.Id == userFileMate.Id {
			userFileList = append(userFileList[:index], userFileMate)
			userFileList = append(userFileList[:index], userFileList[index+1:]...)
		}
	}
	err = u.SetCacheList("UserFileList_"+userFileMate.UserInfo.User, userFileList)

	userFileList, err = u.GetCacheList("UserFileList_" + userFileMate.FileInfo.Hash)
	for index, val := range userFileList {
		if val.Id == userFileMate.Id {
			userFileList = append(userFileList[:index], userFileMate)
			userFileList = append(userFileList[:index], userFileList[index+1:]...)
		}
	}
	err = u.SetCacheList("UserFileList_"+userFileMate.FileInfo.Hash, userFileList)
	return err
}

func (u UserFileManager) DeleteByUserFile(user string, fileHash string) error {
	err := u.DelSqlByUserFile(user, fileHash)
	if err != nil {
		return err
	}
	err = u.DelCache(user + fileHash)

	userFileList, err := u.GetCacheList("UserFileList_" + user)
	for index, val := range userFileList {
		if val.UserInfo.User == user && val.FileInfo.Hash == fileHash {
			userFileList = append(userFileList[:index], userFileList[index+1:]...)
		}
	}
	err = u.SetCacheList("UserFileList_"+user, userFileList)

	userFileList, err = u.GetCacheList("UserFileList_" + fileHash)
	for index, val := range userFileList {
		if val.UserInfo.User == user && val.FileInfo.Hash == fileHash {
			userFileList = append(userFileList[:index], userFileList[index+1:]...)
		}
	}
	err = u.SetCacheList("UserFileList_"+fileHash, userFileList)
	return err
}

func (u UserFileManager) GetCountByUser(user string) (int, error) {
	var res int
	getSql := `
		SELECT COUNT(uf_id)
		FROM user_file_map
		INNER JOIN user_info
		ON uf_ui_id = ui_id
		WHERE ui_user = ?
		AND uf_recycled = 'N'
	`
	rows := utils.Conn.QueryRow(getSql, user)
	err := rows.Scan(&res)
	return res, err
}

func (u UserFileManager) GetSqlByUser(user string) ([]models.UserFileMap, error) {
	var mapList []models.UserFileMap

	getSql := `
		SELECT uf_id, uf_file_name, uf_star,
		uf_is_public, uf_remark, uf_state, uf_recycled,
		ui_id, ui_name, ui_user, ui_pwd, ui_pwd_strength,
		ui_level, ui_email, ui_phone, ui_recycled, 
		fi_id, fi_name, fi_hash, fi_size, fi_state,
		fi_path, fi_remark, fi_is_public, fi_recycled
		FROM user_file_map
		INNER JOIN user_info
		ON uf_ui_id = ui_id
		INNER JOIN file_info
		ON uf_fi_id = fi_id
		WHERE ui_user = ?
		AND uf_recycled = 'N'
		AND ui_recycled = 'N'
		AND fi_recycled = 'N'
	`
	rows, err := utils.Conn.Query(getSql, user)
	if err != nil {
		return mapList, err
	}

	for rows.Next() {
		var userFileMate models.UserFileMap
		_ = rows.Scan(
			&userFileMate.Id, &userFileMate.FileName, &userFileMate.Star,
			&userFileMate.IsPublic, &userFileMate.Remark, &userFileMate.State,
			&userFileMate.Recycled, &userFileMate.UserInfo.Id,
			&userFileMate.UserInfo.Name, &userFileMate.UserInfo.User,
			&userFileMate.UserInfo.Pwd, &userFileMate.UserInfo.PwdStrength,
			&userFileMate.UserInfo.Level, &userFileMate.UserInfo.Email,
			&userFileMate.UserInfo.Phone, &userFileMate.UserInfo.Recycled,
			&userFileMate.FileInfo.Id, &userFileMate.FileInfo.Name,
			&userFileMate.FileInfo.Hash, &userFileMate.FileInfo.Size,
			&userFileMate.FileInfo.State, &userFileMate.FileInfo.Path,
			&userFileMate.FileInfo.Remark, &userFileMate.FileInfo.IsPublic,
			&userFileMate.FileInfo.Recycled,
		)
		mapList = append(mapList, userFileMate)
	}
	return mapList, err
}

func (u UserFileManager) GetSqlByFile(fileHash string) ([]models.UserFileMap, error) {
	var mapList []models.UserFileMap

	getSql := `
		SELECT uf_id, uf_file_name, uf_star,
		uf_is_public, uf_remark, uf_state, uf_recycled,
		ui_id, ui_name, ui_user, ui_pwd, ui_pwd_strength,
		ui_level, ui_email, ui_phone, ui_recycled, 
		fi_id, fi_name, fi_hash, fi_size, fi_state,
		fi_path, fi_remark, fi_is_public, fi_recycled
		FROM user_file_map
		INNER JOIN user_info
		ON uf_ui_id = ui_id
		INNER JOIN file_info
		ON uf_fi_id = fi_id
		WHERE fi_hash = ?
		AND uf_recycled = 'N'
		AND ui_recycled = 'N'
		AND fi_recycled = 'N'
	`
	rows, err := utils.Conn.Query(getSql, fileHash)
	if err != nil {
		return mapList, err
	}

	for rows.Next() {
		var userFileMate models.UserFileMap
		_ = rows.Scan(
			&userFileMate.Id, &userFileMate.FileName, &userFileMate.Star,
			&userFileMate.IsPublic, &userFileMate.Remark, &userFileMate.State,
			&userFileMate.Recycled, &userFileMate.UserInfo.Id,
			&userFileMate.UserInfo.Name, &userFileMate.UserInfo.User,
			&userFileMate.UserInfo.Pwd, &userFileMate.UserInfo.PwdStrength,
			&userFileMate.UserInfo.Level, &userFileMate.UserInfo.Email,
			&userFileMate.UserInfo.Phone, &userFileMate.UserInfo.Recycled,
			&userFileMate.FileInfo.Id, &userFileMate.FileInfo.Name,
			&userFileMate.FileInfo.Hash, &userFileMate.FileInfo.Size,
			&userFileMate.FileInfo.State, &userFileMate.FileInfo.Path,
			&userFileMate.FileInfo.Remark, &userFileMate.FileInfo.IsPublic,
			&userFileMate.FileInfo.Recycled,
		)
		mapList = append(mapList, userFileMate)
	}
	return mapList, err
}

func (u UserFileManager) GetSqlByUserFile(user string, fileHash string) (models.UserFileMap, error) {
	var userFileMate models.UserFileMap
	getSql := `
		SELECT uf_id, uf_file_name, uf_star,
		uf_is_public, uf_remark, uf_state, uf_recycled,
		ui_id, ui_name, ui_user, ui_pwd, ui_pwd_strength,
		ui_level, ui_email, ui_phone, ui_recycled, 
		fi_id, fi_name, fi_hash, fi_size, fi_state,
		fi_path, fi_remark, fi_is_public, fi_recycled
		FROM user_file_map
		INNER JOIN user_info
		ON uf_ui_id = ui_id
		INNER JOIN file_info
		ON uf_fi_id = fi_id
		WHERE ui_user = ?
		AND fi_hash = ?
		AND uf_recycled = 'N'
		AND ui_recycled = 'N'
		AND fi_recycled = 'N'
	`
	rows := utils.Conn.QueryRow(getSql, user, fileHash)
	err := rows.Scan(
		&userFileMate.Id, &userFileMate.FileName, &userFileMate.Star,
		&userFileMate.IsPublic, &userFileMate.Remark, &userFileMate.State,
		&userFileMate.Recycled, &userFileMate.UserInfo.Id,
		&userFileMate.UserInfo.Name, &userFileMate.UserInfo.User,
		&userFileMate.UserInfo.Pwd, &userFileMate.UserInfo.PwdStrength,
		&userFileMate.UserInfo.Level, &userFileMate.UserInfo.Email,
		&userFileMate.UserInfo.Phone, &userFileMate.UserInfo.Recycled,
		&userFileMate.FileInfo.Id, &userFileMate.FileInfo.Name,
		&userFileMate.FileInfo.Hash, &userFileMate.FileInfo.Size,
		&userFileMate.FileInfo.State, &userFileMate.FileInfo.Path,
		&userFileMate.FileInfo.Remark, &userFileMate.FileInfo.IsPublic,
		&userFileMate.FileInfo.Recycled,
	)
	return userFileMate, err
}

func (u UserFileManager) SetSql(userFileMate models.UserFileMap) (int64, error) {
	insertSql := `
		INSERT INTO user_file_map(
			uf_ui_id, uf_fi_id, uf_file_name
		) 
		VALUES (?,?,?)
	`
	res, err := utils.Conn.Exec(
		insertSql,
		userFileMate.UserInfo.Id, userFileMate.FileInfo.Id,
		userFileMate.FileName,
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

func (u UserFileManager) UpdateSql(userFileMate models.UserFileMap) error {
	updateSql := `
		UPDATE user_file_map 
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

func (u UserFileManager) DelSqlByUserFile(user string, fileHash string) error {
	updateSql := `
		UPDATE user_file_map,file_info,user_info
		SET uf_recycled = 'Y'
		WHERE uf_fi_id = fi_id
		AND uf_ui_id = ui_id
		AND fi_hash=?
		AND ui_user=?
	`
	_, err := utils.Conn.Exec(updateSql, user, fileHash)
	return err
}

func (u UserFileManager) GetCache(key string) (models.UserFileMap, error) {
	rc := utils.RedisPool.Get()
	defer rc.Close()

	jsonData, err := redis.Bytes(rc.Do("GET", key))
	userFileMate := models.UserFileMap{}
	if jsonData != nil {
		_ = json.Unmarshal(jsonData, &userFileMate)
	}
	return userFileMate, err
}

func (u UserFileManager) GetCacheList(key string) ([]models.UserFileMap, error) {
	rc := utils.RedisPool.Get()
	defer rc.Close()

	jsonData, err := redis.Bytes(rc.Do("GET", key))
	userFileList := []models.UserFileMap{}
	if jsonData != nil {
		_ = json.Unmarshal(jsonData, &userFileList)
	}
	return userFileList, err
}

func (u UserFileManager) SetCache(key string, userFileMate models.UserFileMap) error {
	jsonData, err := json.Marshal(userFileMate)

	rc := utils.RedisPool.Get()
	defer rc.Close()

	_, err = rc.Do("SET", key, string(jsonData), "EX", conf.REDIS_MAXAGE)
	return err
}

func (u UserFileManager) SetCacheList(key string, userFileMate []models.UserFileMap) error {
	jsonData, err := json.Marshal(userFileMate)

	rc := utils.RedisPool.Get()
	defer rc.Close()

	_, err = rc.Do("SET", key, string(jsonData), "EX", conf.REDIS_MAXAGE)
	return err
}

func (u UserFileManager) DelCache(key string) error {
	rc := utils.RedisPool.Get()
	defer rc.Close()
	_, err := rc.Do("DEL", key)
	return err
}
