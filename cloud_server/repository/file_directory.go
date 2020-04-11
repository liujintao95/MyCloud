package repository

import (
	"MyCloud/cloud_server/models"
	"MyCloud/conf"
	"MyCloud/utils"
	"encoding/json"
	"strconv"
	"strings"

	"github.com/garyburd/redigo/redis"
)

//type IDir interface {
//	GetByFid(int64) ([]models.FileDirectory, error)
//	Set([]models.FileDirectory) error
//	Update(models.FileDirectory) error
//	DeleteById(int64) error
//
//	GetSqlMaxId() (int64, error)
//	GetSqlByFid(int64) ([]models.FileDirectory, error)
//	GetSqlById(int64) (models.FileDirectory, error)
//	SetSql([]models.FileDirectory) error
//	UpdateSql(models.FileDirectory) error
//	DelSqlById(int64) error
//
//	GetCache(string) ([]models.FileDirectory, error)
//	SetCache(string, []models.FileDirectory) error
//	DelCache(string) error
//}

type DirManager struct {
	table string
}

func NewDirManager() *DirManager {
	return &DirManager{table: "file_directory"}
}

func (d *DirManager) GetByFid(fid int64) ([]models.FileDirectory, error) {
	fidStr := strconv.FormatInt(fid, 64)

	dirList, err := d.GetCache("dirFid_" + fidStr)
	if err != nil {
		dirList, err = d.GetSqlByFid(fid)
		if err == nil {
			err = d.SetCache("dirFid_"+fidStr, dirList)
		}
	}
	return dirList, err
}

func (d *DirManager) Set(dirList []models.FileDirectory) error {
	err := d.SetSql(dirList)
	if err == nil {
		dirList, err = d.GetSqlByFid(dirList[0].Fid)

		fidStr := strconv.FormatInt(dirList[0].Fid, 64)
		if err == nil {
			err = d.SetCache("dirFid_"+fidStr, dirList)
		}
	}
	return err
}

func (d *DirManager) Update(dirMate models.FileDirectory) error {
	err := d.UpdateSql(dirMate)
	if err != nil {
		return err
	}
	dirList, err := d.GetSqlByFid(dirMate.Fid)
	if err == nil {
		fidStr := strconv.FormatInt(dirMate.Fid, 64)
		err = d.SetCache("dirFid_"+fidStr, dirList)
	}
	return err
}

func (d *DirManager) DeleteById(id int64) error {
	err := d.DelSqlById(id)
	if err != nil {
		return err
	}

	idStr := strconv.FormatInt(id, 64)
	err = d.DelCache("dirFid_" + idStr)

	dirMate, err := d.GetSqlById(id)
	if err != nil {
		return err
	}
	fidStr := strconv.FormatInt(dirMate.Fid, 64)
	err = d.DelCache("dirFid_" + fidStr)

	return err
}

func (d *DirManager) GetSqlMaxId() (int64, error) {
	var res int64

	getSql := `
		SELECT IF(MAX(fd_id),MAX(fd_id),0)
		FROM file_directory
	`
	row := utils.Conn.QueryRow(getSql)
	err := row.Scan(&res)
	return res, err
}

func (d *DirManager) GetSqlByFid(fid int64) ([]models.FileDirectory, error) {
	var dirList []models.FileDirectory

	getSql := `
		SELECT fd_id, fd_is_dir, fd_dir_name, fd_fid,
		fd_recycled, uf_id, uf_file_name, uf_star, 
		uf_is_public, uf_state, uf_remark, uf_recycled,
		ui_id, ui_name, ui_user, ui_pwd, ui_level, ui_email,
		ui_phone, ui_recycled, fi_id, fi_name, fi_hash,
		fi_size, fi_path, fi_is_public, fi_state, fi_remark,
		fi_recycled
		FROM file_directory
		INNER JOIN user_file_map
		ON fd_uf_id = uf_id
		INNER JOIN user_info
		ON uf_ui_id = ui_id
		INNER JOIN file_info
		ON uf_fi_id = fi_id
		WHERE fd_fid = ?
		AND fd_recycled = 'N'
	`
	rows, err := utils.Conn.Query(getSql, fid)
	if err != nil {
		return dirList, err
	}

	for rows.Next() {
		var dirMate models.FileDirectory
		_ = rows.Scan(
			&dirMate.Id, &dirMate.IsFolder, &dirMate.DirName,
			&dirMate.Fid, &dirMate.Recycled, &dirMate.UserFileMap.Id,
			&dirMate.UserFileMap.FileName, &dirMate.UserFileMap.Star,
			&dirMate.UserFileMap.IsPublic, &dirMate.UserFileMap.State,
			&dirMate.UserFileMap.Remark, &dirMate.UserFileMap.Recycled,
			&dirMate.UserFileMap.UserInfo.Id,
			&dirMate.UserFileMap.UserInfo.Name,
			&dirMate.UserFileMap.UserInfo.User,
			&dirMate.UserFileMap.UserInfo.Pwd,
			&dirMate.UserFileMap.UserInfo.Level,
			&dirMate.UserFileMap.UserInfo.Email,
			&dirMate.UserFileMap.UserInfo.Phone,
			&dirMate.UserFileMap.UserInfo.Recycled,
			&dirMate.UserFileMap.FileInfo.Id,
			&dirMate.UserFileMap.FileInfo.Name,
			&dirMate.UserFileMap.FileInfo.Hash,
			&dirMate.UserFileMap.FileInfo.Size,
			&dirMate.UserFileMap.FileInfo.Path,
			&dirMate.UserFileMap.FileInfo.IsPublic,
			&dirMate.UserFileMap.FileInfo.State,
			&dirMate.UserFileMap.FileInfo.Remark,
			&dirMate.UserFileMap.FileInfo.Recycled,
		)
		dirList = append(dirList, dirMate)
	}
	return dirList, err
}

func (d *DirManager) GetSqlById(id int64) (models.FileDirectory, error) {
	var dirMate models.FileDirectory

	getSql := `
		SELECT fd_id, fd_is_dir, fd_dir_name, fd_fid,
		fd_recycled, uf_id, uf_file_name, uf_star, 
		uf_is_public, uf_state, uf_remark, uf_recycled,
		ui_id, ui_name, ui_user, ui_pwd, ui_level, ui_email,
		ui_phone, ui_recycled, fi_id, fi_name, fi_hash,
		fi_size, fi_path, fi_is_public, fi_state, fi_remark,
		fi_recycled
		FROM file_directory
		INNER JOIN user_file_map
		ON fd_uf_id = uf_id
		INNER JOIN user_info
		ON uf_ui_id = ui_id
		INNER JOIN file_info
		ON uf_fi_id = fi_id
		WHERE fd_id = ?
		AND fd_recycled = 'N'
	`
	rows := utils.Conn.QueryRow(getSql, id)
	err := rows.Scan(
		&dirMate.Id, &dirMate.IsFolder, &dirMate.DirName,
		&dirMate.Fid, &dirMate.Recycled, &dirMate.UserFileMap.Id,
		&dirMate.UserFileMap.FileName, &dirMate.UserFileMap.Star,
		&dirMate.UserFileMap.IsPublic, &dirMate.UserFileMap.State,
		&dirMate.UserFileMap.Remark, &dirMate.UserFileMap.Recycled,
		&dirMate.UserFileMap.UserInfo.Id,
		&dirMate.UserFileMap.UserInfo.Name,
		&dirMate.UserFileMap.UserInfo.User,
		&dirMate.UserFileMap.UserInfo.Pwd,
		&dirMate.UserFileMap.UserInfo.Level,
		&dirMate.UserFileMap.UserInfo.Email,
		&dirMate.UserFileMap.UserInfo.Phone,
		&dirMate.UserFileMap.UserInfo.Recycled,
		&dirMate.UserFileMap.FileInfo.Id,
		&dirMate.UserFileMap.FileInfo.Name,
		&dirMate.UserFileMap.FileInfo.Hash,
		&dirMate.UserFileMap.FileInfo.Size,
		&dirMate.UserFileMap.FileInfo.Path,
		&dirMate.UserFileMap.FileInfo.IsPublic,
		&dirMate.UserFileMap.FileInfo.State,
		&dirMate.UserFileMap.FileInfo.Remark,
		&dirMate.UserFileMap.FileInfo.Recycled,
	)
	return dirMate, err
}

func (d *DirManager) SetSql(dirList []models.FileDirectory) error {
	var setlist []interface{}

	insertSql := `
		INSERT INTO file_directory(
			fd_id, fd_uf_id, fd_is_dir,
			fd_dir_name, fd_fid
		) 
		VALUES
	`
	var insertList []string
	for _, dirMate := range dirList {
		insertList = append(insertList, "(?,?,?,?,?)")
		setlist = append(setlist, dirMate.Id)
		setlist = append(setlist, dirMate.UserFileMap.Id)
		setlist = append(setlist, dirMate.IsFolder)
		setlist = append(setlist, dirMate.DirName)
		setlist = append(setlist, dirMate.Fid)
	}
	insertSql += strings.Join(insertList, ",")

	_, err := utils.Conn.Exec(insertSql, setlist...)
	return err
}

func (d *DirManager) UpdateSql(dirMate models.FileDirectory) error {
	updateSql := `
		UPDATE file_directory 
		SET fd_uf_id=?, fd_is_dir=?, fd_dir_name=?,
		fd_fid=?, fd_recycled=?
		WHERE fd_id=?
	`
	_, err := utils.Conn.Exec(
		updateSql,
		dirMate.UserFileMap.Id, dirMate.IsFolder, dirMate.DirName,
		dirMate.Fid, dirMate.Recycled, dirMate.Id,
	)
	return err
}

func (d *DirManager) DelSqlById(id int64) error {
	updateSql := `
		UPDATE file_directory 
		SET fd_recycled = 'Y'
		WHERE fd_id=?
	`
	_, err := utils.Conn.Exec(updateSql, id)
	return err
}

func (d *DirManager) GetCache(key string) ([]models.FileDirectory, error) {
	rc := utils.RedisPool.Get()
	defer rc.Close()

	jsonData, err := redis.Bytes(rc.Do("GET", key))
	var dirList []models.FileDirectory
	if jsonData != nil {
		_ = json.Unmarshal(jsonData, &dirList)
	}
	return dirList, err
}

func (d *DirManager) SetCache(key string, dirList []models.FileDirectory) error {
	jsonData, err := json.Marshal(dirList)

	rc := utils.RedisPool.Get()
	defer rc.Close()

	_, err = rc.Do("SET", key, string(jsonData), "EX", conf.REDIS_MAXAGE)
	return err
}

func (d *DirManager) DelCache(key string) error {
	rc := utils.RedisPool.Get()
	defer rc.Close()
	_, err := rc.Do("DEL", key)
	return err
}
