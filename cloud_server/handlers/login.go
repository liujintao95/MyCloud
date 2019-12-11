package handlers

import (
	"MyCloud/cloud_server/models"
	"MyCloud/utils"
)

var errCheck = utils.ErrCheck

func GetUserInfo(user string) (res models.UserInfo, err error) {
	getSql := `
		SELECT ui_id, ui_user, ui_pwd, ui_level,
		ui_email,ui_phone, ui_recycled
		FROM user_info 
		WHERE ui_user = ?
	`
	rows := utils.Conn.QueryRow(getSql, user)
	err = rows.Scan(
		&res.Id, &res.User, &res.Pwd, &res.Level,
		&res.Email, &res.Phone, &res.Recycled,
	)
	return
}

func SetNewUser(userInfo models.UserInfo) (uid int64, err error) {
	res, err := utils.Conn.Exec(
		"INSERT INTO user_info(ui_user,ui_pwd,ui_level,ui_email,ui_phone) VALUES (?,?,?,?,?)",
		userInfo.User, userInfo.Pwd, userInfo.Level, userInfo.Email, userInfo.Phone)
	errCheck(err, "Error insert mysql UserInfo", true)
	id, err := res.LastInsertId()
	errCheck(err, "Error get insert last id", true)
	return id, err
}

func UpdateUserPwd(user string, pwd string) error {
	_, err := utils.Conn.Exec(
		"UPDATE user_info SET ui_pwd=? WHERE ui_user=?",
		pwd, user)
	return err
}
