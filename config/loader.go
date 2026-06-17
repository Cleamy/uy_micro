package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

var AppCfg *AppConfig

func LoadConfig(customPath ...string) error {
	v := viper.New()
	v.SetConfigType("yaml")

	// 自定义 目录  >  当前目录 > config 目录 > 上级目录
	if len(customPath) > 0 && customPath[0] != "" {
		v.SetConfigFile(customPath[0])
	} else {
		v.SetConfigName("config")
		v.AddConfigPath(".")
		v.AddConfigPath("./config")
		v.AddConfigPath("../config")
		v.AddConfigPath("/etc/uy_micro")
	}

	v.SetEnvPrefix("UY_MICRO")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	// 注入全量默认值
	setAllDefaults(v)

	// 读取配置，文件不存在不报错，使用默认值
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return fmt.Errorf("read config failed: %w", err)
		}
	}

	// 反序列化到结构体
	AppCfg = &AppConfig{}
	if err := v.Unmarshal(AppCfg); err != nil {
		return fmt.Errorf("unmarshal config failed: %w", err)
	}

	// 配置热更新监听（可选）
	// v.WatchConfig()
	// v.OnConfigChange(func(e fsnotify.Event) {
	//     _ = v.Unmarshal(AppCfg)
	// })

	return nil

}

// 全量默认值配置
func setAllDefaults(v *viper.Viper) {
	// 基础配置
	v.SetDefault("app.name", "uy-micro-service")
	v.SetDefault("app.env", "dev")
	v.SetDefault("app.version", "v1.0.0")

	// 日志默认值（必选组件）
	v.SetDefault("logger.level", "info")
	v.SetDefault("logger.format", "console")
	v.SetDefault("logger.max_size", 100)
	v.SetDefault("logger.max_backups", 10)
	v.SetDefault("logger.max_age", 30)
	v.SetDefault("logger.compress", false)

	// Web默认值（默认启用）
	v.SetDefault("web.enable", true)
	v.SetDefault("web.port", 8080)
	v.SetDefault("web.mode", "debug")
	v.SetDefault("web.context_path", "/")
	v.SetDefault("web.read_timeout", 30)
	v.SetDefault("web.write_timeout", 30)

	// 数据库默认禁用
	v.SetDefault("database.enable", false)
	v.SetDefault("database.primary.driver", "mysql")
	v.SetDefault("database.primary.charset", "utf8mb4")
	v.SetDefault("database.primary.max_idle", 10)
	v.SetDefault("database.primary.max_open", 100)
	v.SetDefault("database.primary.max_lifetime", 3600)

	// Redis默认禁用
	v.SetDefault("redis.enable", false)
	v.SetDefault("redis.mode", "single")
	v.SetDefault("redis.db", 0)
	v.SetDefault("redis.pool_size", 20)
	v.SetDefault("redis.min_idle", 5)

	// MQ默认禁用
	v.SetDefault("rabbitmq.enable", false)
	v.SetDefault("rabbitmq.port", 5672)
	v.SetDefault("rabbitmq.vhost", "/")

	// gRPC默认禁用
	v.SetDefault("grpc.enable", false)
	v.SetDefault("grpc.port", 9000)
	v.SetDefault("grpc.reflection", false)

	// Consul默认禁用
	v.SetDefault("consul.enable", false)
	v.SetDefault("consul.address", "127.0.0.1:8500")
	v.SetDefault("consul.check.interval", "10s")
	v.SetDefault("consul.check.timeout", "5s")
	v.SetDefault("consul.check.deregister", "30s")

	// Sentinel默认禁用
	v.SetDefault("sentinel.enable", false)

	// 网关默认禁用
	v.SetDefault("gateway.enable", false)
	v.SetDefault("gateway.port", 8000)
	v.SetDefault("gateway.context_path", "/")
	v.SetDefault("gateway.cors.enable", true)
	v.SetDefault("gateway.cors.max_age", 3600)
	v.SetDefault("gateway.auth.enable", false)
	v.SetDefault("gateway.auth.type", "jwt")
	v.SetDefault("gateway.auth.jwt_expire", 24)

	// 可观测性默认禁用
	v.SetDefault("observability.tracer.enable", false)
	v.SetDefault("observability.metrics.enable", false)
	v.SetDefault("observability.metrics.path", "/metrics")
	v.SetDefault("observability.metrics.port", 9090)
}
