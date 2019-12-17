package repository

import (
	"MyCloud/cloud_server/models"
	"MyCloud/conf"
	"MyCloud/utils"
	"encoding/json"
	"github.com/garyburd/redigo/redis"
)

type IUser interface {
	SelectByUser(string) (*models.UserInfo, error)
	Insert(*models.UserInfo) (int64, error)
	UpdatePassword(string, string) error
	Delete(string) error
	GetCache(string) (*models.UserInfo, error)
	SetCache(string, *models.UserInfo) error
	DelCache(string) error
}

type UserManager struct {
	table string
}

func NewUserManager() IUser {
	return &UserManager{table: "user_info"}
}

func (u *UserManager) SelectByUser(user string) (res *models.UserInfo, err error) {
	obj := models.UserInfo{}
	getSql := `
		SELECT ui_id, ui_user, ui_pwd, ui_level,
		ui_email,ui_phone, ui_recycled
		FROM user_info 
		WHERE ui_user = ?
	`
	rows := utils.Conn.QueryRow(getSql, user)
	err = rows.Scan(
		&obj.Id, &obj.User, &obj.Pwd, &obj.Level,
		&obj.Email, &obj.Phone, &obj.Recycled,
	)
	return &obj, err
}

func (u *UserManager) Insert(obj *models.UserInfo) (id int64, err error) {
	insertSql := `
		INSERT INTO user_info(
			ui_user, ui_pwd, ui_level, ui_email, ui_phone
		) 
		VALUES (?,?,?,?,?)`
	res, err := utils.Conn.Exec(insertSql, obj.User, obj.Pwd, obj.Level, obj.Email, obj.Phone)
	if err != nil{
		return -1, err
	}
	id, err = res.LastInsertId()
	if err != nil{
		return -1, err
	}
	return
}

func (u *UserManager) UpdatePassword(user string, pwd string) error {
	updateSql := `
		UPDATE user_info 
		SET ui_pwd=?
		WHERE ui_user=?
	`
	_, err := utils.Conn.Exec(updateSql, pwd, user)
	return err
}

func (u *UserManager) Delete(user string) (err error) {
	updateSql := `
		UPDATE user_info 
		SET ui_recycled = 'Y'
		WHERE ui_user=?
	`
	_, err = utils.Conn.Exec(updateSql, user)
	return
}

func (u *UserManager) GetCache(key string) (res *models.UserInfo, err error) {
	rc := utils.RedisPool.Get()
	defer rc.Close()

	jsonData, err := redis.Bytes(rc.Do("LRANGE", key, 0, -1))
	userInfo := models.UserInfo{}
	if jsonData != nil {
		_ = json.Unmarshal(jsonData, userInfo)
	}
	return &userInfo, err
}

func (u *UserManager) SetCache(user string, obj *models.UserInfo) (err error) {
	jsonData, err := json.Marshal(obj)

	rc := utils.RedisPool.Get()
	defer rc.Close()

	_, err = rc.Do("LPUSH", user, string(jsonData), "EX", string(conf.REDIS_MAXAGE))
	return
}

func (u *UserManager) DelCache(key string) (err error) {
	rc := utils.RedisPool.Get()
	defer rc.Close()
	_, err = rc.Do("DEL", key)
	return
}
