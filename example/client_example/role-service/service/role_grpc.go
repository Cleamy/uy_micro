package service

import (
	rolepb "common_pkg/proto/role/v1"
	"context"
)

/*
	没用继承


	组合




*/

type RoleGrpcServiceImpl struct {
	rolepb.UnimplementedRoleServiceServer
	roleSvc RoleService
}

func NewRoleGrpcService(roleSvc RoleService) *RoleGrpcServiceImpl {
	return &RoleGrpcServiceImpl{roleSvc: roleSvc}
}

func (s *RoleGrpcServiceImpl) GetRoleByID(ctx context.Context, req *rolepb.GetRoleByIDReq) (*rolepb.GetRoleByIDResp, error) {
	role, err := s.roleSvc.GetByID(req.Id)
	if err != nil {
		return &rolepb.GetRoleByIDResp{
			Code: 500,
			Msg:  "query role failed: " + err.Error(),
		}, nil
	}
	

	return &rolepb.GetRoleByIDResp{
		Code: 200,
		Msg:  "ok",
		Data: &rolepb.Role{
			Id:        role.ID,
			Name:      role.Name,
			Code:      role.Code,
			Remark:    role.Remark,
			CreatedAt: role.CreatedAt.Unix(),
		},
	}, nil
}

func (s *RoleGrpcServiceImpl) BatchGetRoles(ctx context.Context, req *rolepb.BatchGetRolesReq) (*rolepb.BatchGetRolesResp, error) {
	roleMap, err := s.roleSvc.BatchGet(req.Ids)
	if err != nil {
		return &rolepb.BatchGetRolesResp{
			Code: 500,
			Msg:  "batch query failed",
		}, nil
	}

	data := make(map[uint64]*rolepb.Role, len(roleMap))
	for id, role := range roleMap {
		data[id] = &rolepb.Role{
			Id:        role.ID,
			Name:      role.Name,
			Code:      role.Code,
			Remark:    role.Remark,
			CreatedAt: role.CreatedAt.Unix(),
		}
	}

	return &rolepb.BatchGetRolesResp{
		Code: 200,
		Msg:  "ok",
		Data: data,
	}, nil
}
