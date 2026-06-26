package service

import (
	rolepb "common_pkg/proto/role/v1"
	"user-service/client"
	"user-service/dao"
	"user-service/model"
)

// UserService 用户业务服务接口
type UserService interface {
	GetByID(id uint64) (*model.User, error)
	Create(user *model.User) error
	List() ([]model.User, error)
	GetWithRole(id uint64) (*model.User, *rolepb.Role, error)
}

type userServiceImpl struct{}

// NewUserService 构造函数
func NewUserService() UserService {
	return &userServiceImpl{}
}

// GetByID 纯查询用户信息
func (s *userServiceImpl) GetByID(id uint64) (*model.User, error) {

	return dao.GetUserByID(id)
}

// Create 创建用户
func (s *userServiceImpl) Create(user *model.User) error {
	// 这里可以加业务校验：用户名重复、角色合法性校验等
	return dao.CreateUser(user)
}

// List 查询用户列表
func (s *userServiceImpl) List() ([]model.User, error) {
	return dao.ListUsers()
}

// GetWithRole 组合查询：本地用户信息 + gRPC 调用角色服务
func (s *userServiceImpl) GetWithRole(id uint64) (*model.User, *rolepb.Role, error) {
	// 1. 本地查用户
	user, err := dao.GetUserByID(id)
	if err != nil {
		return nil, nil, err
	}

	// 2. 跨服务调用角色信息，失败降级不中断主流程
	role, err := client.GetRoleByID(user.RoleID)
	if err != nil {
		return user, nil, nil
	}

	return user, role, nil
}
