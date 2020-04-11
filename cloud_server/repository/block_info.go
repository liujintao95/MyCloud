package repository

import (
	"MyCloud/cloud_server/models"
	"MyCloud/conf"
	"MyCloud/utils"
	"encoding/json"
	"github.com/garyburd/redigo/redis"
)

//type IBlock interface {
//	GetByUploadId(string) ([]models.BlockInfo, error)
//	GetByUploadIdIndex(string, int) (models.BlockInfo, error)
//	Set([]models.BlockInfo) error
//	Update(models.BlockInfo) error
//	DeleteByUploadId(string) error
//
//	GetSqlByUploadId(string) ([]models.BlockInfo, error)
//	GetSqlByUploadIdIndex(string, int) (models.BlockInfo, error)
//	SetSql([]models.BlockInfo) error
//	UpdateSql(models.BlockInfo) error
//	DelSqlByUploadId(string) error
//
//	GetCache(string) ([]models.BlockInfo, error)
//	SetCache(string, []models.BlockInfo) error
//	DelCache(string) error
//}

type BlockManager struct {
	table string
}

func NewBlockManager() *BlockManager {
	return &BlockManager{table: "block_info"}
}

func (b *BlockManager) GetByUploadIdIndex(uploadID string, index int) (models.BlockInfo, error) {
	blockList, err := b.GetCache("bl_" + uploadID)

	if err != nil {
		for _, blockMate := range blockList {
			if blockMate.Index == index {
				return blockMate, err
			}
		}
	}
	blockMate, err := b.GetSqlByUploadIdIndex(uploadID, index)
	return blockMate, err
}

func (b *BlockManager) GetByUploadId(uploadID string) ([]models.BlockInfo, error) {
	blockList, err := b.GetCache("bl_" + uploadID)
	if err != nil {
		blockList, err = b.GetByUploadId(uploadID)
		if err == nil {
			err = b.SetCache("bl_"+uploadID, blockList)
		}
	}
	return blockList, err
}

func (b *BlockManager) Set(blockList []models.BlockInfo) error {
	err := b.SetSql(blockList)
	if err == nil {
		blockList, err = b.GetByUploadId(blockList[0].FileBlockInfo.UploadID)
		if err == nil {
			err = b.SetCache("bl_"+blockList[0].FileBlockInfo.UploadID, blockList)
		}
	}
	return err
}

func (b *BlockManager) Update(blockMate models.BlockInfo) error {
	err := b.UpdateSql(blockMate)
	if err != nil {
		return err
	}
	blockList, err := b.GetByUploadId(blockMate.FileBlockInfo.UploadID)
	if err == nil {
		err = b.SetCache("bl_"+blockMate.FileBlockInfo.UploadID, blockList)
	}
	return err
}

func (b *BlockManager) DeleteByUploadId(uploadID string) error {
	err := b.DeleteByUploadId(uploadID)
	if err != nil {
		return err
	}
	err = b.DelCache("bl_" + uploadID)
	return err
}

func (b *BlockManager) GetSqlByUploadIdIndex(uploadId string, index int) (models.BlockInfo, error) {
	var blockMate models.BlockInfo

	getSql := `
		SELECT bi_id, bi_index, bi_path, bi_size, bi_state,
		bi_recycled, fbi_id, fbi_hash, fbi_file_name, fbi_path,
		fbi_upload_id, fbi_file_size, fbi_block_size, 
		fbi_block_count, fbi_recycled, ui_id, ui_name, 
		ui_user, ui_pwd, ui_level, ui_email, ui_phone,
		ui_recycled
		FROM block_info
		INNER JOIN file_block_info
		ON bi_upload_id = fbi_upload_id
		INNER JOIN user_info
		ON fbi_ui_id = ui_id
		WHERE bi_upload_id = ?
		AND bi_index = ?
		AND bi_recycled = 'N'
	`
	rows := utils.Conn.QueryRow(getSql, uploadId, index)
	err := rows.Scan(
		&blockMate.Id, &blockMate.Index, &blockMate.Path,
		&blockMate.Size, &blockMate.State, &blockMate.Recycled,
		&blockMate.FileBlockInfo.Id, &blockMate.FileBlockInfo.Hash,
		&blockMate.FileBlockInfo.FileName, &blockMate.FileBlockInfo.Path,
		&blockMate.FileBlockInfo.UploadID, &blockMate.FileBlockInfo.FileSize,
		&blockMate.FileBlockInfo.BlockSize, &blockMate.FileBlockInfo.BlockCount,
		&blockMate.FileBlockInfo.Recycled,
		&blockMate.FileBlockInfo.UserInfo.Id,
		&blockMate.FileBlockInfo.UserInfo.Name,
		&blockMate.FileBlockInfo.UserInfo.User,
		&blockMate.FileBlockInfo.UserInfo.Pwd,
		&blockMate.FileBlockInfo.UserInfo.Level,
		&blockMate.FileBlockInfo.UserInfo.Email,
		&blockMate.FileBlockInfo.UserInfo.Phone,
		&blockMate.FileBlockInfo.UserInfo.Recycled,
	)
	return blockMate, err
}

func (b *BlockManager) GetSqlByUploadId(uploadId string) ([]models.BlockInfo, error) {
	var blockList []models.BlockInfo

	getSql := `
		SELECT bi_id, bi_index, bi_path, bi_size, bi_state,
		bi_recycled, fbi_id, fbi_hash, fbi_file_name, fbi_path,
		fbi_upload_id, fbi_file_size, fbi_block_size, 
		fbi_block_count, fbi_recycled, ui_id, ui_name, 
		ui_user, ui_pwd, ui_level, ui_email, ui_phone,
		ui_recycled
		FROM block_info
		INNER JOIN file_block_info
		ON bi_upload_id = fbi_upload_id
		INNER JOIN user_info
		ON fbi_ui_id = ui_id
		WHERE bi_upload_id = ?
		AND bi_recycled = 'N'
	`
	rows, err := utils.Conn.Query(getSql, uploadId)
	if err != nil {
		return blockList, err
	}

	for rows.Next() {
		var blockMate models.BlockInfo
		_ = rows.Scan(
			&blockMate.Id, &blockMate.Index, &blockMate.Path,
			&blockMate.Size, &blockMate.State, &blockMate.Recycled,
			&blockMate.FileBlockInfo.Id, &blockMate.FileBlockInfo.Hash,
			&blockMate.FileBlockInfo.FileName, &blockMate.FileBlockInfo.Path,
			&blockMate.FileBlockInfo.UploadID, &blockMate.FileBlockInfo.FileSize,
			&blockMate.FileBlockInfo.BlockSize, &blockMate.FileBlockInfo.BlockCount,
			&blockMate.FileBlockInfo.Recycled,
			&blockMate.FileBlockInfo.UserInfo.Id,
			&blockMate.FileBlockInfo.UserInfo.Name,
			&blockMate.FileBlockInfo.UserInfo.User,
			&blockMate.FileBlockInfo.UserInfo.Pwd,
			&blockMate.FileBlockInfo.UserInfo.Level,
			&blockMate.FileBlockInfo.UserInfo.Email,
			&blockMate.FileBlockInfo.UserInfo.Phone,
			&blockMate.FileBlockInfo.UserInfo.Recycled,
		)
		blockList = append(blockList, blockMate)
	}
	return blockList, err
}

func (b *BlockManager) SetSql(blockList []models.BlockInfo) error {
	var setlist []interface{}
	insertSql := `
		INSERT INTO block_info(
			bi_upload_id, bi_index, bi_path,
			bi_size, bi_state
		) 
		VALUES (?,?,?,?,?)
	`

	for _, blockMate := range blockList {
		insertSql += ",(?,?,?,?,?)"
		setlist = append(setlist, blockMate.FileBlockInfo.UploadID)
		setlist = append(setlist, blockMate.Index)
		setlist = append(setlist, blockMate.Path)
		setlist = append(setlist, blockMate.Size)
		setlist = append(setlist, blockMate.State)
	}

	_, err := utils.Conn.Exec(insertSql, setlist)
	return err
}

func (b *BlockManager) UpdateSql(blockMate models.BlockInfo) error {
	updateSql := `
		UPDATE block_info 
		SET bi_path=?, bi_size=?, bi_state=?, bi_recycled=?
		WHERE bi_upload_id=?
		AND bi_index = ?
	`
	_, err := utils.Conn.Exec(
		updateSql,
		blockMate.Path, blockMate.Size, blockMate.State,
		blockMate.Recycled, blockMate.FileBlockInfo.UploadID,
		blockMate.Index,
	)
	return err
}

func (b *BlockManager) DelSqlByUploadId(uploadId string) error {
	updateSql := `
		UPDATE block_info 
		SET bi_recycled='Y'
		WHERE bi_upload_id=?
	`
	_, err := utils.Conn.Exec(updateSql, uploadId)
	return err
}

func (b *BlockManager) GetCache(key string) ([]models.BlockInfo, error) {
	rc := utils.RedisPool.Get()
	defer rc.Close()

	jsonData, err := redis.Bytes(rc.Do("GET", key))
	var blockList []models.BlockInfo
	if jsonData != nil {
		_ = json.Unmarshal(jsonData, &blockList)
	}
	return blockList, err
}

func (b *BlockManager) SetCache(key string, fileBlockList []models.BlockInfo) error {
	jsonData, err := json.Marshal(fileBlockList)

	rc := utils.RedisPool.Get()
	defer rc.Close()

	_, err = rc.Do("SET", key, string(jsonData), "EX", conf.REDIS_MAXAGE)
	return err
}

func (b *BlockManager) DelCache(key string) error {
	rc := utils.RedisPool.Get()
	defer rc.Close()
	_, err := rc.Do("DEL", key)
	return err
}
