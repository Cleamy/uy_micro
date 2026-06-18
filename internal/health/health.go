package health

import (
	"context"
	"net/http"
	"time"

	"github.com/Cleamy/uy_micro/global"

	"github.com/gin-gonic/gin"
)

type CheckResult struct {
	Status  string            `json:"status"`
	Details map[string]string `json:"details,omitempty"`
}

const (
	StatusUp   = "UP"
	StatusDown = "DOWN"
)

// LivenessCheck 存活探针
// 仅判断进程是否存活，不检查依赖；K8s 用于判断是否需要重启容器
func LivenessCheck(c *gin.Context) {
	c.JSON(http.StatusOK, CheckResult{
		Status: StatusUp,
	})
}

// ReadinessCheck 就绪探针
// 全量检查核心依赖可用性；K8s 用于判断是否将流量切进实例，滚动发布时使用
func ReadinessCheck(c *gin.Context) {
	result := CheckResult{
		Status:  StatusUp,
		Details: make(map[string]string),
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 3*time.Second)
	defer cancel()

	// 检查数据库连通性
	if global.DB != nil {
		sqlDB, err := global.DB.DB()
		if err != nil {
			result.Details["database"] = err.Error()
			result.Status = StatusDown
		} else if err := sqlDB.PingContext(ctx); err != nil {
			result.Details["database"] = err.Error()
			result.Status = StatusDown
		} else {
			result.Details["database"] = StatusUp
		}
	}

	// 检查 Redis 连通性
	if global.Redis != nil {
		if err := global.Redis.Ping(ctx).Err(); err != nil {
			result.Details["redis"] = err.Error()
			result.Status = StatusDown
		} else {
			result.Details["redis"] = StatusUp
		}
	}

	// 检查 Consul 连通性
	if global.Consul != nil {
		if _, err := global.Consul.Agent().Self(); err != nil {
			result.Details["consul"] = err.Error()
			result.Status = StatusDown
		} else {
			result.Details["consul"] = StatusUp
		}
	}

	// 状态对应 HTTP 码：正常200，不可用503
	if result.Status == StatusUp {
		c.JSON(http.StatusOK, result)
	} else {
		c.JSON(http.StatusServiceUnavailable, result)
	}
}
