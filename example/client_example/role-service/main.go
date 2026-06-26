package main

import (
	rolepb "common_pkg/proto/role/v1"
	"role-service/controller"
	"role-service/model"
	"role-service/service"

	"github.com/Cleamy/uy_micro"
	"github.com/Cleamy/uy_micro/global"

	"github.com/gin-gonic/gin"
)

func main() {
	uy_micro.Server.OnBootstrap(func() {
		// 1. 自动迁移表结构（PostgreSQL 兼容）
		_ = global.DB.AutoMigrate(&model.Role{})

		// 2. 依赖注入：业务层 → 接口层
		roleSvc := service.NewRoleService()
		roleController := controller.NewRoleController(roleSvc)
		roleGrpcSvc := service.NewRoleGrpcService(roleSvc)

		// 3. 注册 HTTP 路由
		registerHttpRoutes(global.Web, roleController)

		// 4. 注册 gRPC 服务到框架实例
		rolepb.RegisterRoleServiceServer(global.Grpc, roleGrpcSvc)
	})

	uy_micro.Server.Run()
}

func registerHttpRoutes(r *gin.Engine, c *controller.RoleController) {
	group := r.Group("/api/v1/role")
	{
		group.GET("/:id", c.GetByID)
		group.POST("", c.Create)
		group.GET("/list", c.List)
	}
}
