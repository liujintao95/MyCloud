package repository

import (
	"MyCloud/cloud_server/models"
	"MyCloud/conf"
	"MyCloud/utils"
	"encoding/json"

	"github.com/garyburd/redigo/redis"
)

type IUser interface {
	GetByUser(string) (models.UserInfo, error)
	Set(string, models.UserInfo) (int64, error)
	Update(string, models.UserInfo) error
	DeleteByUser(string) error

	GetSqlByUser(string) (models.UserInfo, error)
	SetSql(models.UserInfo) (int64, error)
	UpdateSql(models.UserInfo) error
	DelSqlByUser(string) error

	GetCache(string) (models.UserInfo, error)
	SetCache(string, models.UserInfo) error
	DelCache(string) error
}

type UserManager struct {
	table string
}

func NewUserManager() IUser {
	return &UserManager{table: "user_info"}
}

func (u *UserManager) GetByUser(user string) (models.UserInfo, error) {
	userMate, err := u.GetCache(user)
	if err != nil {
		return userMate, err
	}
	if userMate.Pwd == "" {
		userMate, err = u.GetSqlByUser(user)
		if err == nil {
			err = u.SetCache(user, userMate)
		}
	}
	return userMate, err
}

func (u *UserManager) Set(key string, userMate models.UserInfo) (int64, error) {
	uid, err := u.SetSql(userMate)
	if err != nil {
		return -1, err
	}
	userMate.Id = uid
	err = u.SetCache(key, userMate)
	return uid, err
}

func (u *UserManager) Update(key string, userMate models.UserInfo) error {
	err := u.UpdateSql(userMate)
	if err != nil {
		return err
	}
	_ = u.SetCache(userMate.User, userMate)
	err = u.SetCache(key, userMate)
	return err
}

func (u *UserManager) DeleteByUser(string) error {
	panic("implement me")
}

func (u *UserManager) GetSqlByUser(user string) (models.UserInfo, error) {
	userMate := new(models.UserInfo)
	getSql := `
		SELECT ui_id, ui_name, ui_user, ui_pwd,
		ui_level, ui_email,ui_phone, ui_recycled
		FROM user_info 
		WHERE ui_user = ?
		AND ui_recycled == 'N'
	`
	rows := utils.Conn.QueryRow(getSql, user)
	err := rows.Scan(
		userMate.Id, userMate.Name, userMate.User,
		userMate.Pwd, userMate.Level, userMate.Email,
		userMate.Phone, userMate.Recycled,
	)
	return *userMate, err
}

func (u *UserManager) SetSql(userMate models.UserInfo) (int64, error) {
	insertSql := `
		INSERT INTO user_info(
			ui_user, ui_name, ui_pwd,
			ui_level, ui_email, ui_phone
		) 
		VALUES (?,?,?,?,?,?)`
	res, err := utils.Conn.Exec(
		insertSql,
		userMate.User, userMate.User, userMate.Pwd,
		userMate.Level, userMate.Email, userMate.Phone,
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

func (u *UserManager) UpdateSql(userMate models.UserInfo) error {
	updateSql := `
		UPDATE user_info 
		SET ui_name=?, ui_pwd=?, ui_level=?, ui_email=?, ui_phone=?
		WHERE ui_user=?
	`
	_, err := utils.Conn.Exec(
		updateSql,
		userMate.Name, userMate.Pwd, userMate.Level,
		userMate.Email, userMate.Phone, userMate.User,
	)
	return err
}

func (u *UserManager) DelSqlByUser(user string) error {
	updateSql := `
		UPDATE user_info 
		SET ui_recycled = 'Y'
		WHERE ui_user=?
	`
	_, err := utils.Conn.Exec(updateSql, user)
	return err
}

func (u *UserManager) GetCache(key string) (models.UserInfo, error) {
	rc := utils.RedisPool.Get()
	defer rc.Close()

	jsonData, err := redis.Bytes(rc.Do("LRANGE", key, 0, -1))
	userMate := models.UserInfo{}
	if jsonData != nil {
		_ = json.Unmarshal(jsonData, &userMate)
	}
	return userMate, err
}

func (u *UserManager) SetCache(key string, userMate models.UserInfo) error {
	jsonData, err := json.Marshal(userMate)

	rc := utils.RedisPool.Get()
	defer rc.Close()

	_, err = rc.Do("LPUSH", key, string(jsonData), "EX", string(conf.REDIS_MAXAGE))
	return err
}

func (u *UserManager) DelCache(key string) error {
	rc := utils.RedisPool.Get()
	defer rc.Close()
	_, err := rc.Do("DEL", key)
	return err
}
