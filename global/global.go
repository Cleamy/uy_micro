package global

import (
	"errors"
	"github.com/Cleamy/uy_micro/config"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/hashicorp/consul/api"
	"github.com/streadway/amqp"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"gorm.io/gorm"
)

var (
	// 配置
	Config *config.AppConfig

	// 可观测性
	Logger *zap.Logger
	Tracer trace.Tracer

	// 数据存储
	DB    *gorm.DB // 主库实例
	Redis redis.UniversalClient

	// 消息队列
	RabbitMQ *amqp.Connection

	// 服务容器
	Web  *gin.Engine  // Gin HTTP 引擎
	Grpc *grpc.Server // gRPC 服务端

	// 服务注册发现
	Consul *api.Client

	// 网关
	Gateway *gin.Engine

	// GRPC客户端工厂（接口声明，不引入具体实现包，消除循环依赖）
	RpcFactory RpcClientFactory
)

// RpcClientFactory 客户端工厂接口定义
type RpcClientFactory interface {
	GetConn(serviceName string) (*grpc.ClientConn, func(), error)
	CloseAll() // 关闭所有缓存的长连接，服务优雅关停时调用
}

// ==================== 空实现兜底（组件禁用时使用，防空指针） ====================

// noopRpcClientFactory gRPC客户端工厂空实现
type noopRpcClientFactory struct{}

func (n *noopRpcClientFactory) GetConn(serviceName string) (*grpc.ClientConn, func(), error) {
	return nil, nil, errors.New("rpc client factory is disabled, please enable consul first")
}

func (n *noopRpcClientFactory) CloseAll() {
	// 空实现，无操作
}

// 初始化默认兜底实例
func init() {
	// 默认先赋值空实现，避免启动过程中误调用panic
	RpcFactory = &noopRpcClientFactory{}
}
