package registry

import (
	"fmt"
	"math/rand"
	"strconv"
	"sync"
	"time"

	"github.com/Cleamy/uy_micro/config"
	"github.com/Cleamy/uy_micro/global"

	"github.com/hashicorp/consul/api"
	"go.uber.org/zap"
)

// 负载均衡策略枚举
type BalanceStrategy string

const (
	BalanceRoundRobin    BalanceStrategy = "round_robin"    // 轮询
	BalanceRandom        BalanceStrategy = "random"         // 随机
	cacheRefreshInterval                 = 30 * time.Second // 缓存刷新间隔
)

// 全局状态变量
var (
	// 轮询游标，按服务+协议隔离
	balanceIdxMap = make(map[string]int)
	idxMu         sync.Mutex
	rander        = rand.New(rand.NewSource(time.Now().UnixNano()))

	// 服务实例本地缓存：key = 服务名_协议，value = 实例列表
	serviceCache sync.Map
	cacheTicker  *time.Ticker
)

// Init 初始化 Consul 客户端并自动注册当前服务
func Init(cfg *config.ConsulConfig) (*api.Client, error) {
	if !cfg.Enable {
		return nil, nil
	}

	client, err := api.NewClient(&api.Config{
		Address: cfg.Address,
	})
	if err != nil {
		return nil, fmt.Errorf("create consul client failed: %w", err)
	}

	svcPort := global.Config.Web.Port
	if !global.Config.Web.Enable {
		return nil, fmt.Errorf("web server must enable for consul register health check")
	}

	grpcPort := 0
	if global.Config.Grpc.Enable {
		grpcPort = global.Config.Grpc.Port
	}

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

	if err := client.Agent().ServiceRegister(reg); err != nil {
		return nil, fmt.Errorf("consul service register failed: %w", err)
	}

	global.Logger.Info("consul service registered",
		zap.String("service_id", svcID),
		zap.String("service_name", global.Config.App.Name),
		zap.Int("web_port", svcPort),
		zap.Int("grpc_port", grpcPort),
	)

	// 启动后台缓存刷新协程
	cacheTicker = time.NewTicker(cacheRefreshInterval)
	go refreshCacheLoop()

	return client, nil
}

// 后台循环刷新所有已缓存服务的实例列表
func refreshCacheLoop() {
	for range cacheTicker.C {
		serviceCache.Range(func(key, value interface{}) bool {
			serviceKey := key.(string)
			// 从key中还原服务名（去掉协议后缀）
			serviceName := ""
			if len(serviceKey) > 5 && serviceKey[len(serviceKey)-5:] == "_grpc" {
				serviceName = serviceKey[:len(serviceKey)-5]
			} else if len(serviceKey) > 5 && serviceKey[len(serviceKey)-5:] == "_http" {
				serviceName = serviceKey[:len(serviceKey)-5]
			}
			if serviceName == "" {
				return true
			}

			// 异步刷新，失败不覆盖旧缓存
			services, _, err := global.Consul.Health().Service(serviceName, "", true, nil)
			if err == nil && len(services) > 0 {
				serviceCache.Store(serviceKey, services)
				global.Logger.Debug("[RegistryCache] refresh service cache success",
					zap.String("service_key", serviceKey),
					zap.Int("instance_count", len(services)))
			} else {
				global.Logger.Warn("[RegistryCache] refresh service cache failed, use old cache",
					zap.String("service_key", serviceKey),
					zap.Error(err))
			}
			return true
		})
	}
}

// 通用：从缓存+Consul获取服务实例列表（带兜底）
func getServiceInstances(serviceKey string, serviceName string, passingOnly bool) ([]*api.ServiceEntry, error) {
	// 1. 优先读本地缓存
	if cached, ok := serviceCache.Load(serviceKey); ok {
		services := cached.([]*api.ServiceEntry)
		if len(services) > 0 {
			return services, nil
		}
	}

	// 2. 缓存未命中，实时拉取Consul
	services, _, err := global.Consul.Health().Service(serviceName, "", passingOnly, nil)
	if err != nil {
		return nil, fmt.Errorf("discover service %s failed: %w", serviceName, err)
	}
	if len(services) == 0 {
		return nil, fmt.Errorf("no healthy instance for service [%s]", serviceName)
	}

	// 3. 写入缓存，后续请求直接复用
	serviceCache.Store(serviceKey, services)
	global.Logger.Info("[RegistryCache] new service cached",
		zap.String("service_key", serviceKey),
		zap.Int("instance_count", len(services)))

	return services, nil
}

// ==================== gRPC 服务发现 ====================

// GetServiceGrpcTarget 获取服务 gRPC 地址，内置负载均衡与缓存兜底
func GetServiceGrpcTarget(serviceName string, passingOnly bool, strategy BalanceStrategy) (string, error) {
	serviceKey := serviceName + "_grpc"

	services, err := getServiceInstances(serviceKey, serviceName, passingOnly)
	if err != nil {
		// Consul故障强兜底：再查一次缓存，哪怕是过期的
		if cached, ok := serviceCache.Load(serviceKey); ok {
			services = cached.([]*api.ServiceEntry)
			if len(services) > 0 {
				global.Logger.Warn("[RegistryCache] consul unavailable, fallback to cache",
					zap.String("service", serviceName))
			}
		}
		if len(services) == 0 {
			return "", err
		}
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
		idx := balanceIdxMap[serviceKey] % len(services)
		selectEntry = services[idx]
		balanceIdxMap[serviceKey]++
		idxMu.Unlock()
	}

	svc := selectEntry.Service
	rpcPort, ok := svc.Meta["grpc_port"]
	if !ok || rpcPort == "" {
		errMsg := fmt.Sprintf("service [%s] missing grpc_port meta", serviceName)
		global.Logger.Error(errMsg,
			zap.String("service_id", svc.ID),
			zap.String("service", serviceName))
		return "", fmt.Errorf(errMsg)
	}

	target := fmt.Sprintf("%s:%s", svc.Address, rpcPort)
	global.Logger.Debug("[RegistryBalance] select grpc instance",
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

// ==================== HTTP 服务发现（网关专用） ====================

// GetServiceHttpTarget 获取服务 HTTP 地址，内置负载均衡与缓存兜底
func GetServiceHttpTarget(serviceName string, passingOnly bool) (string, error) {
	serviceKey := serviceName + "_http"

	services, err := getServiceInstances(serviceKey, serviceName, passingOnly)
	if err != nil {
		// Consul故障强兜底：再查一次缓存，哪怕是过期的
		if cached, ok := serviceCache.Load(serviceKey); ok {
			services = cached.([]*api.ServiceEntry)
			if len(services) > 0 {
				global.Logger.Warn("[RegistryCache] consul unavailable, fallback to cache",
					zap.String("service", serviceName))
			}
		}
		if len(services) == 0 {
			return "", err
		}
	}

	var selectEntry *api.ServiceEntry
	strategy := BalanceRoundRobin

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
		idx := balanceIdxMap[serviceKey] % len(services)
		selectEntry = services[idx]
		balanceIdxMap[serviceKey]++
		idxMu.Unlock()
	}

	svc := selectEntry.Service
	target := svc.Address + ":" + strconv.Itoa(svc.Port)

	global.Logger.Debug("[RegistryBalance] select http instance",
		zap.String("service", serviceName),
		zap.String("strategy", string(strategy)),
		zap.String("instance_id", svc.ID),
		zap.String("http_target", target))

	return target, nil
}
