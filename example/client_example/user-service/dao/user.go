package dao

import (
	"user-service/model"

	"github.com/Cleamy/uy_micro/global"
)

func GetUserByID(id uint64) (*model.User, error) {
	var user model.User
	err := global.DB.First(&user, id).Error
	return &user, err
}

func CreateUser(user *model.User) error {
	return global.DB.Create(user).Error
}

func ListUsers() ([]model.User, error) {
	var list []model.User
	err := global.DB.Find(&list).Error
	return list, err
}
