package service

import (
	"role-service/dao"
	"role-service/model"
)

type RoleService interface {
	GetByID(id uint64) (*model.Role, error)
	Create(role *model.Role) error
	List() ([]*model.Role, error)
	BatchGet(ids []uint64) (map[uint64]*model.Role, error)
}

type roleServiceImpl struct{}

func NewRoleService() RoleService {
	return &roleServiceImpl{}
}

func (s *roleServiceImpl) GetByID(id uint64) (*model.Role, error) {
	return dao.GetRoleByID(id)
}

func (s *roleServiceImpl) Create(role *model.Role) error {
	// 可扩展业务校验：角色编码重复、名称长度校验等
	return dao.CreateRole(role)
}

func (s *roleServiceImpl) List() ([]*model.Role, error) {
	return dao.ListRoles()
}

func (s *roleServiceImpl) BatchGet(ids []uint64) (map[uint64]*model.Role, error) {
	return dao.BatchGetRoles(ids)
}
