package client

import (
	"context"
	"fmt"

	rolepb "common_pkg/proto/role/v1"

	"github.com/Cleamy/uy_micro/global"
)

// GetRoleClient 通过全局工厂获取 role 服务 gRPC 客户端
func GetRoleClient() (rolepb.RoleServiceClient, func(), error) {
	// 建立grpc 连接
	conn, clean, err := global.RpcFactory.GetConn("role-service")
	if err != nil {
		return nil, nil, err
	}
	// 返回 role service 客户端
	return rolepb.NewRoleServiceClient(conn), clean, nil
}

// GetRoleByID 对外暴露的调用方法
func GetRoleByID(roleID uint64) (*rolepb.Role, error) {
	client, cleanup, err := GetRoleClient()
	if err != nil {
		return nil, err
	}
	defer cleanup()

	resp, err := client.GetRoleByID(context.Background(), &rolepb.GetRoleByIDReq{Id: roleID})
	if err != nil {
		return nil, err
	}
	if resp.Code != 200 {
		return nil, fmt.Errorf(resp.Msg)
	}
	return resp.Data, nil
}
