package controller

import (
	"role-service/model"
	"role-service/service"
	"strconv"

	"github.com/gin-gonic/gin"
)

type RoleController struct {
	roleSvc service.RoleService
}

func NewRoleController(roleSvc service.RoleService) *RoleController {
	return &RoleController{roleSvc: roleSvc}
}

// GetByID 根据ID查询角色
func (c *RoleController) GetByID(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		ctx.JSON(400, gin.H{"code": 400, "msg": "invalid role id"})
		return
	}

	role, err := c.roleSvc.GetByID(id)
	if err != nil {
		ctx.JSON(500, gin.H{"code": 500, "msg": "query failed: " + err.Error()})
		return
	}
	ctx.JSON(200, gin.H{"code": 200, "data": role})
}

// Create 创建角色
func (c *RoleController) Create(ctx *gin.Context) {
	var role model.Role
	if err := ctx.ShouldBindJSON(&role); err != nil {
		ctx.JSON(400, gin.H{"code": 400, "msg": "invalid params: " + err.Error()})
		return
	}

	if err := c.roleSvc.Create(&role); err != nil {
		ctx.JSON(500, gin.H{"code": 500, "msg": "create failed: " + err.Error()})
		return
	}
	ctx.JSON(200, gin.H{"code": 200, "data": role})
}

// List 查询角色列表
func (c *RoleController) List(ctx *gin.Context) {
	list, err := c.roleSvc.List()
	if err != nil {
		ctx.JSON(500, gin.H{"code": 500, "msg": "query failed: " + err.Error()})
		return
	}
	ctx.JSON(200, gin.H{"code": 200, "data": list})
}
