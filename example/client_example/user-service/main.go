package main

import (
	"user-service/controller"
	"user-service/model"
	"user-service/service"

	"github.com/Cleamy/uy_micro"
	"github.com/Cleamy/uy_micro/global"

	"github.com/gin-gonic/gin"
)

func main() {
	uy_micro.Server.OnBootstrap(func() {
		// 1. 自动迁移表结构
		_ = global.DB.AutoMigrate(&model.User{})

		// 2. 依赖注入：service → controller
		userSvc := service.NewUserService()
		userController := controller.NewUserController(userSvc)

		// 3. 注册 HTTP 路由
		registerRoutes(global.Web, userController)
	})

	uy_micro.Server.Run()
}

func registerRoutes(r *gin.Engine, c *controller.UserController) {
	group := r.Group("/api/v1/user")
	{
		group.GET("/:id", c.GetByID)
		group.GET("/:id/role", c.GetWithRole)
		group.POST("", c.Create)
		group.GET("/list", c.List)
	}
}
