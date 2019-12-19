package repository

import (
	"MyCloud/cloud_server/models"
	"MyCloud/conf"
	"MyCloud/utils"
	"encoding/json"
	"github.com/garyburd/redigo/redis"
)

type IUser interface {
	SelectByUser(string) (models.UserInfo, error)
	Insert(models.UserInfo) (int64, error)
	Update(models.UserInfo) error
	Delete(string) error
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

func (u *UserManager) SelectByUser(user string) (models.UserInfo, error) {
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

func (u *UserManager) Insert(userMate models.UserInfo) (int64, error) {
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

func (u *UserManager) Update(userMate models.UserInfo) error {
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

func (u *UserManager) Delete(user string) error {
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
		_ = json.Unmarshal(jsonData, userMate)
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
