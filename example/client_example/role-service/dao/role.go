package dao

import (
	"role-service/model"
	"github.com/Cleamy/uy_micro/global"
)

func GetRoleByID(id uint64) (*model.Role, error) {
	var role model.Role
	err := global.DB.First(&role, id).Error
	return &role, err
}

func CreateRole(role *model.Role) error {
	return global.DB.Create(role).Error
}

func ListRoles() ([]*model.Role, error) {
	var list []*model.Role
	err := global.DB.Find(&list).Error
	return list, err
}

func BatchGetRoles(ids []uint64) (map[uint64]*model.Role, error) {
	var roles []*model.Role
	err := global.DB.Where("id IN ?", ids).Find(&roles).Error
	if err != nil {
		return nil, err
	}

	result := make(map[uint64]*model.Role, len(roles))
	for _, r := range roles {
		result[r.ID] = r
	}
	return result, nil
}
