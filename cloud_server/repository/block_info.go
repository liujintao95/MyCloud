package repository

import "MyCloud/cloud_server/models"

type IBlock interface {
	GetByUploadId(string) ([]models.BlockInfo, error)
	GetByUploadIdIndex(string, string) (models.BlockInfo, error)
	Set(string, models.BlockInfo) (int64, error)
	Update(string, models.BlockInfo) error
	DeleteByUploadId(string) error

	GetSqlByUploadId(string) ([]models.BlockInfo, error)
	GetSqlByUploadIdIndex(string, string) (models.BlockInfo, error)
	SetSql(models.BlockInfo) (int64, error)
	UpdateSql(models.BlockInfo) error
	DelSqlByUploadId(string) error

	GetCache(string) (models.BlockInfo, error)
	SetCache(string, models.BlockInfo) error
	DelCache(string) error
}

type BlockManager struct {
	table string
}

func NewBlockManager() IBlock {
	return &BlockManager{table: "block_info"}
}

func (b BlockManager) GetByUploadId(string) ([]models.BlockInfo, error) {
	panic("implement me")
}

func (b BlockManager) GetByUploadIdIndex(string, string) (models.BlockInfo, error) {
	panic("implement me")
}

func (b BlockManager) Set(string, models.BlockInfo) (int64, error) {
	panic("implement me")
}

func (b BlockManager) Update(string, models.BlockInfo) error {
	panic("implement me")
}

func (b BlockManager) DeleteByUploadId(string) error {
	panic("implement me")
}

func (b BlockManager) GetSqlByUploadId(string) ([]models.BlockInfo, error) {
	panic("implement me")
}

func (b BlockManager) GetSqlByUploadIdIndex(string, string) (models.BlockInfo, error) {
	panic("implement me")
}

func (b BlockManager) SetSql(models.BlockInfo) (int64, error) {
	panic("implement me")
}

func (b BlockManager) UpdateSql(models.BlockInfo) error {
	panic("implement me")
}

func (b BlockManager) DelSqlByUploadId(string) error {
	panic("implement me")
}

func (b BlockManager) GetCache(string) (models.BlockInfo, error) {
	panic("implement me")
}

func (b BlockManager) SetCache(string, models.BlockInfo) error {
	panic("implement me")
}

func (b BlockManager) DelCache(string) error {
	panic("implement me")
}
