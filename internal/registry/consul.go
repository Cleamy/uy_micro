package registry

import (
	"fmt"
	"math/rand"
	"strconv"
	"sync"
	"time"

	"uy_micro/config"
	"uy_micro/global"

	"github.com/hashicorp/consul/api"
	"go.uber.org/zap"
)

// 负载均衡策略枚举
type BalanceStrategy string

const (
	BalanceRoundRobin BalanceStrategy = "round_robin" // 轮询
	BalanceRandom     BalanceStrategy = "random"      // 随机
)

// 全局轮询游标：按服务隔离，并发安全
var (
	balanceIdxMap = make(map[string]int)
	idxMu         sync.Mutex
	rander        = rand.New(rand.NewSource(time.Now().UnixNano()))
)

// Init 初始化 Consul 客户端并自动注册当前服务
func Init(cfg *config.ConsulConfig) (*api.Client, error) {
	if !cfg.Enable {
		return nil, nil
	}

	// 1. 创建 Consul 客户端
	client, err := api.NewClient(&api.Config{
		Address: cfg.Address,
	})
	if err != nil {
		return nil, fmt.Errorf("create consul client failed: %w", err)
	}

	// 固定使用 Web 端口做主体注册，HTTP 健康检测
	svcPort := global.Config.Web.Port
	if !global.Config.Web.Enable {
		return nil, fmt.Errorf("web server must enable for consul register health check")
	}

	grpcPort := 0
	if global.Config.Grpc.Enable {
		grpcPort = global.Config.Grpc.Port
	}

	// 2. 构造服务注册信息
	svcID := fmt.Sprintf("%s-web-%d", global.Config.App.Name, svcPort)
	reg := &api.AgentServiceRegistration{
		ID:      svcID,
		Name:    global.Config.App.Name,
		Address: "127.0.0.1",
		Port:    svcPort,
		Tags:    cfg.Tags,
		Meta: map[string]string{
			"grpc_enable": strconv.FormatBool(global.Config.Grpc.Enable),
			"grpc_port":   strconv.Itoa(grpcPort),
		},
		Check: &api.AgentServiceCheck{
			Interval:                       cfg.Check.Interval,
			Timeout:                        cfg.Check.Timeout,
			DeregisterCriticalServiceAfter: cfg.Check.Deregister,
			HTTP:                           fmt.Sprintf("http://127.0.0.1:%d/health", svcPort),
		},
	}

	// 3. 执行服务注册
	if err := client.Agent().ServiceRegister(reg); err != nil {
		return nil, fmt.Errorf("consul service register failed: %w", err)
	}

	global.Logger.Info("consul service registered",
		zap.String("service_id", svcID),
		zap.String("service_name", global.Config.App.Name),
		zap.Int("web_port", svcPort),
		zap.Int("grpc_port", grpcPort),
	)

	return client, nil
}

// GetServiceGrpcTarget 获取服务 GRPC 地址，内置负载均衡与日志
func GetServiceGrpcTarget(serviceName string, passingOnly bool, strategy BalanceStrategy) (string, error) {
	// 1. 拉取健康实例列表
	services, _, err := global.Consul.Health().Service(serviceName, "", passingOnly, nil)
	if err != nil {
		global.Logger.Error("registry discover service failed",
			zap.String("service", serviceName),
			zap.Error(err))
		return "", fmt.Errorf("discover %s err: %w", serviceName, err)
	}
	if len(services) == 0 {
		errMsg := fmt.Sprintf("no healthy instance for service [%s]", serviceName)
		global.Logger.Error(errMsg, zap.String("service", serviceName))
		return "", fmt.Errorf(errMsg)
	}

	// 2. 按策略选择节点
	var selectEntry *api.ServiceEntry
	switch strategy {
	case BalanceRandom:
		idxMu.Lock()
		r := rander.Intn(len(services))
		selectEntry = services[r]
		idxMu.Unlock()
	case BalanceRoundRobin:
		fallthrough
	default:
		idxMu.Lock()
		idx := balanceIdxMap[serviceName] % len(services)
		selectEntry = services[idx]
		balanceIdxMap[serviceName]++
		idxMu.Unlock()
	}

	svc := selectEntry.Service
	// 3. 读取元数据中的 GRPC 端口
	rpcPort, ok := svc.Meta["grpc_port"]
	if !ok || rpcPort == "" {
		errMsg := fmt.Sprintf("service [%s] missing grpc_port meta", serviceName)
		global.Logger.Error(errMsg,
			zap.String("service_id", svc.ID),
			zap.String("service", serviceName))
		return "", fmt.Errorf(errMsg)
	}

	target := fmt.Sprintf("%s:%s", svc.Address, rpcPort)
	// 负载均衡选择日志
	global.Logger.Info("[RegistryBalance] select instance success",
		zap.String("service", serviceName),
		zap.String("strategy", string(strategy)),
		zap.String("instance_id", svc.ID),
		zap.String("grpc_target", target))

	return target, nil
}

// GetServiceGrpcTargetDefault 默认轮询策略，简化调用
func GetServiceGrpcTargetDefault(serviceName string, passingOnly bool) (string, error) {
	return GetServiceGrpcTarget(serviceName, passingOnly, BalanceRoundRobin)
}

// GetServiceHttpTarget 获取服务 HTTP 地址，内置负载均衡（网关专用）
func GetServiceHttpTarget(serviceName string, passingOnly bool) (string, error) {
	return GetServiceHttpTargetWithStrategy(serviceName, passingOnly, BalanceRoundRobin)
}

func GetServiceHttpTargetWithStrategy(serviceName string, passingOnly bool, strategy BalanceStrategy) (string, error) {
	services, _, err := global.Consul.Health().Service(serviceName, "", passingOnly, nil)
	if err != nil {
		global.Logger.Error("registry discover http service failed",
			zap.String("service", serviceName), zap.Error(err))
		return "", fmt.Errorf("discover %s err: %w", serviceName, err)
	}
	if len(services) == 0 {
		return "", fmt.Errorf("no healthy http instance for service [%s]", serviceName)
	}

	var selectEntry *api.ServiceEntry
	switch strategy {
	case BalanceRandom:
		idxMu.Lock()
		r := rander.Intn(len(services))
		selectEntry = services[r]
		idxMu.Unlock()
	case BalanceRoundRobin:
		fallthrough
	default:
		idxMu.Lock()
		idx := balanceIdxMap[serviceName+"_http"] % len(services)
		selectEntry = services[idx]
		balanceIdxMap[serviceName+"_http"]++
		idxMu.Unlock()
	}

	svc := selectEntry.Service
	target := svc.Address + ":" + strconv.Itoa(svc.Port)

	global.Logger.Info("[RegistryBalance] select http instance success",
		zap.String("service", serviceName),
		zap.String("strategy", string(strategy)),
		zap.String("instance_id", svc.ID),
		zap.String("http_target", target))

	return target, nil
}
