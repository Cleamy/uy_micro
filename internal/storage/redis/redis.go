package rediscli

import (
	"context"
	"fmt"
	"uy_micro/config"

	"github.com/go-redis/redis/v8"
)

// Init 初始化 Redis 客户端（支持单机/集群模式）
func Init(cfg *config.RedisConfig) (redis.UniversalClient, error) {
	if !cfg.Enable {
		return nil, nil
	}

	var client redis.UniversalClient

	switch cfg.Mode {
	case "single":
		if len(cfg.Addresses) == 0 {
			return nil, fmt.Errorf("redis single mode requires at least one address")
		}
		client = redis.NewClient(&redis.Options{
			Addr:         cfg.Addresses[0],
			Password:     cfg.Password,
			DB:           cfg.DB,
			PoolSize:     cfg.PoolSize,
			MinIdleConns: cfg.MinIdle,
		})

	case "cluster":
		client = redis.NewClusterClient(&redis.ClusterOptions{
			Addrs:        cfg.Addresses,
			Password:     cfg.Password,
			PoolSize:     cfg.PoolSize,
			MinIdleConns: cfg.MinIdle,
		})

	default:
		return nil, fmt.Errorf("unsupported redis mode: %s", cfg.Mode)
	}

	// 连通性校验
	ctx := context.Background()
	if _, err := client.Ping(ctx).Result(); err != nil {
		return nil, fmt.Errorf("redis ping failed: %w", err)
	}

	return client, nil
}
