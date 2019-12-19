package repository

import "MyCloud/cloud_server/models"

type IUserFileMap interface {
	SelectByUser(string) ([]models.UserFileMap, error)
	SelectByFile(string) ([]models.UserFileMap, error)
	SelectByUserFile(string, string) (models.UserFileMap, error)
	Insert(models.UserFileMap) (int64, error)
	Update(models.UserFileMap) error
	Delete(string, string) error
	GetCache(string) (models.UserFileMap, error)
	SetCache(string, models.UserFileMap) error
	DelCache(string) error
}

type UserFileManager struct {
	table string
}

func NewUserFileManager() IUserFileMap {
	return &UserFileManager{table: "user_file_map"}
}

func (u UserFileManager) SelectByUser(string) ([]models.UserFileMap, error) {
	panic("implement me")
}

func (u UserFileManager) SelectByFile(string) ([]models.UserFileMap, error) {
	panic("implement me")
}

func (u UserFileManager) SelectByUserFile(string, string) (models.UserFileMap, error) {
	panic("implement me")
}

func (u UserFileManager) Insert(models.UserFileMap) (int64, error) {
	panic("implement me")
}

func (u UserFileManager) Update(models.UserFileMap) error {
	panic("implement me")
}

func (u UserFileManager) Delete(string, string) error {
	panic("implement me")
}

func (u UserFileManager) GetCache(string) (models.UserFileMap, error) {
	panic("implement me")
}

func (u UserFileManager) SetCache(string, models.UserFileMap) error {
	panic("implement me")
}

func (u UserFileManager) DelCache(string) error {
	panic("implement me")
}

