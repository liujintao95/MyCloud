package repository

import "MyCloud/cloud_server/models"

type IFileBlock interface {
	GetByHash(string) ([]models.FileBlockInfo, error)
	GetByUserHash(string, string) (models.FileBlockInfo, error)
	GetByUploadId(string) (models.FileBlockInfo, error)
	Set(string, models.FileBlockInfo) (int64, error)
	Update(string, models.FileBlockInfo) error
	DeleteByHash(string) error
	DeleteByUploadId(string) error

	GetSqlByHash(string) (models.FileBlockInfo, error)
	GetSqlByUploadId(string) (models.FileBlockInfo, error)
	SetSql(models.FileBlockInfo) (int64, error)
	UpdateSql(models.FileBlockInfo) error
	DelSqlByHash(string) error
	DelSqlByUploadId(string) error

	GetCache(string) (models.FileBlockInfo, error)
	SetCache(string, models.FileBlockInfo) error
	DelCache(string) error
}

type FileBlockManager struct {
	table string
}

func NewFileBlockManager() IFileBlock {
	return &FileBlockManager{table: "file_block_info"}
}

func (f FileBlockManager) GetByHash(string) ([]models.FileBlockInfo, error) {
	panic("implement me")
}

func (f FileBlockManager) GetByUserHash(string, string) (models.FileBlockInfo, error) {
	panic("implement me")
}

func (f FileBlockManager) GetByUploadId(string) (models.FileBlockInfo, error) {
	panic("implement me")
}

func (f FileBlockManager) Set(string, models.FileBlockInfo) (int64, error) {
	panic("implement me")
}

func (f FileBlockManager) Update(string, models.FileBlockInfo) error {
	panic("implement me")
}

func (f FileBlockManager) DeleteByHash(string) error {
	panic("implement me")
}

func (f FileBlockManager) DeleteByUploadId(string) error {
	panic("implement me")
}

func (f FileBlockManager) GetSqlByHash(string) (models.FileBlockInfo, error) {
	panic("implement me")
}

func (f FileBlockManager) GetSqlByUploadId(string) (models.FileBlockInfo, error) {
	panic("implement me")
}

func (f FileBlockManager) SetSql(models.FileBlockInfo) (int64, error) {
	panic("implement me")
}

func (f FileBlockManager) UpdateSql(models.FileBlockInfo) error {
	panic("implement me")
}

func (f FileBlockManager) DelSqlByHash(string) error {
	panic("implement me")
}

func (f FileBlockManager) DelSqlByUploadId(string) error {
	panic("implement me")
}

func (f FileBlockManager) GetCache(string) (models.FileBlockInfo, error) {
	panic("implement me")
}

func (f FileBlockManager) SetCache(string, models.FileBlockInfo) error {
	panic("implement me")
}

func (f FileBlockManager) DelCache(string) error {
	panic("implement me")
}
