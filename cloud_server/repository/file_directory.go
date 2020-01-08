package repository

import "MyCloud/cloud_server/models"

type IDir interface {
	GetByFid(int) ([]models.FileDirectory, error)
	GetById(int) (models.FileDirectory, error)
	Set([]models.FileDirectory) error
	Update(models.FileDirectory) error
	DeleteById(int) error

	GetSqlMaxId() (int64, error)
	GetSqlByFid(int) ([]models.FileDirectory, error)
	GetSqlById(int) (models.FileDirectory, error)
	SetSql([]models.FileDirectory) error
	UpdateSql(models.FileDirectory) error
	DelSqlByUploadId(int) error

	GetCache(string) ([]models.FileDirectory, error)
	SetCache(string, []models.FileDirectory) error
	DelCache(string) error
}

type DirManager struct {
	table string
}

func NewDirManager() IDir {
	return &DirManager{table: "file_directory"}
}

func (d DirManager) GetByFid(int) ([]models.FileDirectory, error) {
	panic("implement me")
}

func (d DirManager) GetById(int) (models.FileDirectory, error) {
	panic("implement me")
}

func (d DirManager) Set([]models.FileDirectory) error {
	panic("implement me")
}

func (d DirManager) Update(models.FileDirectory) error {
	panic("implement me")
}

func (d DirManager) DeleteById(int) error {
	panic("implement me")
}

func (d DirManager) GetSqlMaxId() (int64, error) {
	panic("implement me")
}

func (d DirManager) GetSqlByFid(int) ([]models.FileDirectory, error) {
	panic("implement me")
}

func (d DirManager) GetSqlById(int) (models.FileDirectory, error) {
	panic("implement me")
}

func (d DirManager) SetSql([]models.FileDirectory) error {
	panic("implement me")
}

func (d DirManager) UpdateSql(models.FileDirectory) error {
	panic("implement me")
}

func (d DirManager) DelSqlByUploadId(int) error {
	panic("implement me")
}

func (d DirManager) GetCache(string) ([]models.FileDirectory, error) {
	panic("implement me")
}

func (d DirManager) SetCache(string, []models.FileDirectory) error {
	panic("implement me")
}

func (d DirManager) DelCache(string) error {
	panic("implement me")
}
