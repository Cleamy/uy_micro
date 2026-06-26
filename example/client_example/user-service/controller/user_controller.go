package controller

import (
	"strconv"
	"user-service/model"
	"user-service/service"

	"github.com/gin-gonic/gin"
)

type UserController struct {
	userSvc service.UserService
}

func NewUserController(userSvc service.UserService) *UserController {
	return &UserController{userSvc: userSvc}
}

// GetByID 根据ID查询用户
func (c *UserController) GetByID(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		ctx.JSON(400, gin.H{"code": 400, "msg": "invalid user id"})
		return
	}

	user, err := c.userSvc.GetByID(id)
	if err != nil {
		ctx.JSON(500, gin.H{"code": 500, "msg": "query failed: " + err.Error()})
		return
	}
	ctx.JSON(200, gin.H{"code": 200, "data": user})
}

// GetWithRole 查询用户及关联角色信息
func (c *UserController) GetWithRole(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		ctx.JSON(400, gin.H{"code": 400, "msg": "invalid user id"})
		return
	}

	user, role, err := c.userSvc.GetWithRole(id)
	if err != nil {
		ctx.JSON(500, gin.H{"code": 500, "msg": "query failed: " + err.Error()})
		return
	}
	ctx.JSON(200, gin.H{
		"code": 200,
		"data": gin.H{
			"user": user,
			"role": role,
		},
	})
}

// Create 创建用户
func (c *UserController) Create(ctx *gin.Context) {
	var user model.User
	if err := ctx.ShouldBindJSON(&user); err != nil {
		ctx.JSON(400, gin.H{"code": 400, "msg": "invalid params: " + err.Error()})
		return
	}

	if err := c.userSvc.Create(&user); err != nil {
		ctx.JSON(500, gin.H{"code": 500, "msg": "create failed: " + err.Error()})
		return
	}
	ctx.JSON(200, gin.H{"code": 200, "data": user})
}

// List 查询用户列表
func (c *UserController) List(ctx *gin.Context) {
	list, err := c.userSvc.List()
	if err != nil {
		ctx.JSON(500, gin.H{"code": 500, "msg": "query failed: " + err.Error()})
		return
	}
	ctx.JSON(200, gin.H{"code": 200, "data": list})
}
