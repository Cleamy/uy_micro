package config

// AppConfig 根配置
type AppConfig struct {
	App           AppBasicConfig      `mapstructure:"app"`
	Logger        LoggerConfig        `mapstructure:"logger"`
	Database      DatabaseConfig      `mapstructure:"database"`
	Redis         RedisConfig         `mapstructure:"redis"`
	RabbitMQ      RabbitMQConfig      `mapstructure:"rabbitmq"`
	Web           WebConfig           `mapstructure:"web"`
	Grpc          GrpcConfig          `mapstructure:"grpc"`
	Consul        ConsulConfig        `mapstructure:"consul"`
	Sentinel      SentinelConfig      `mapstructure:"sentinel"`
	Gateway       GatewayConfig       `mapstructure:"gateway"`
	Observability ObservabilityConfig `mapstructure:"observability"`
}

// 基础配置
type AppBasicConfig struct {
	Name    string `mapstructure:"name"`    // 服务名
	Env     string `mapstructure:"env"`     // 环境：dev/test/prod
	Version string `mapstructure:"version"` // 服务版本
}

// 日志配置
type LoggerConfig struct {
	Level      string `mapstructure:"level"`       // debug/info/warn/error
	FilePath   string `mapstructure:"file_path"`   // 日志文件路径，为空则仅控制台输出
	MaxSize    int    `mapstructure:"max_size"`    // 单文件大小(MB)
	MaxBackups int    `mapstructure:"max_backups"` // 保留文件数
	MaxAge     int    `mapstructure:"max_age"`     // 保留天数
	Compress   bool   `mapstructure:"compress"`    // 是否压缩
	Format     string `mapstructure:"format"`      // json/console
}

// 数据库配置（多数据源）
type DatabaseConfig struct {
	Enable  bool               `mapstructure:"enable"`
	Primary DatasourceConfig   `mapstructure:"primary"` // 主库
	Slaves  []DatasourceConfig `mapstructure:"slaves"`  // 从库列表（读写分离）
}

type DatasourceConfig struct {
	Driver      string `mapstructure:"driver"` // mysql/postgres/sqlite
	Host        string `mapstructure:"host"`
	Port        int    `mapstructure:"port"`
	User        string `mapstructure:"user"`
	Password    string `mapstructure:"password"`
	DBName      string `mapstructure:"db_name"`
	Charset     string `mapstructure:"charset"`
	MaxIdle     int    `mapstructure:"max_idle"`
	MaxOpen     int    `mapstructure:"max_open"`
	MaxLifeTime int    `mapstructure:"max_lifetime"` // 连接生命周期(秒)
}

// Redis配置
type RedisConfig struct {
	Enable    bool     `mapstructure:"enable"`
	Mode      string   `mapstructure:"mode"` // single/cluster/sentinel
	Addresses []string `mapstructure:"addresses"`
	Password  string   `mapstructure:"password"`
	DB        int      `mapstructure:"db"` // 单节点模式有效
	PoolSize  int      `mapstructure:"pool_size"`
	MinIdle   int      `mapstructure:"min_idle"`
}

// RabbitMQ配置
type RabbitMQConfig struct {
	Enable   bool   `mapstructure:"enable"`
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	Vhost    string `mapstructure:"vhost"`
}

// Web服务配置
type WebConfig struct {
	Enable       bool   `mapstructure:"enable"`
	Port         int    `mapstructure:"port"`
	Mode         string `mapstructure:"mode"` // debug/release/test
	ContextPath  string `mapstructure:"context_path"`
	ReadTimeout  int    `mapstructure:"read_timeout"`  // 秒
	WriteTimeout int    `mapstructure:"write_timeout"` // 秒
}

// gRPC配置
type GrpcConfig struct {
	Enable     bool `mapstructure:"enable"`
	Port       int  `mapstructure:"port"`
	Reflection bool `mapstructure:"reflection"` // 是否开启反射
}

// Consul配置
type ConsulConfig struct {
	Enable  bool     `mapstructure:"enable"`
	Address string   `mapstructure:"address"`
	Tags    []string `mapstructure:"tags"`
	Check   struct {
		Interval   string `mapstructure:"interval"`   // 健康检查间隔
		Timeout    string `mapstructure:"timeout"`    // 超时时间
		Deregister string `mapstructure:"deregister"` // 故障注销时间
	} `mapstructure:"check"`
}

// Sentinel流量治理配置
// ==================== Sentinel流量治理配置 ====================
type SentinelConfig struct {
	Enable              bool                  `mapstructure:"enable"`
	FlowRules           []SentinelFlowRule    `mapstructure:"flow_rules"`
	CircuitBreakerRules []SentinelCircuitRule `mapstructure:"circuit_breaker_rules"`
	HotSpotRules        []SentinelHotSpotRule `mapstructure:"hot_spot_rules"`
	SystemRules         []SentinelSystemRule  `mapstructure:"system_rules"`
}

// 1. 流控规则 100% 对齐 flow.Rule 官方定义
type SentinelFlowRule struct {
	ID                     string  `mapstructure:"id"`
	Resource               string  `mapstructure:"resource"`
	Threshold              float64 `mapstructure:"threshold"`
	StatIntervalInMs       uint32  `mapstructure:"stat_interval_ms"` // 官方类型 uint32
	ControlBehavior        int     `mapstructure:"control_behavior"`
	MaxQueueingTimeMs      uint32  `mapstructure:"max_queueing_time_ms"` // 官方类型 uint32
	TokenCalculateStrategy int     `mapstructure:"token_calculate_strategy"`
	RelationStrategy       int     `mapstructure:"relation_strategy"`
	RefResource            string  `mapstructure:"ref_resource"`
	WarmUpPeriodSec        uint32  `mapstructure:"warm_up_period_sec"`
	WarmUpColdFactor       uint32  `mapstructure:"warm_up_cold_factor"`
	LowMemUsageThreshold   int64   `mapstructure:"low_mem_usage_threshold"`
	HighMemUsageThreshold  int64   `mapstructure:"high_mem_usage_threshold"`
	MemLowWaterMarkBytes   int64   `mapstructure:"mem_low_water_mark_bytes"`
	MemHighWaterMarkBytes  int64   `mapstructure:"mem_high_water_mark_bytes"`
}

// 2. 热点参数限流 100% 对齐 hotspot.Rule 官方定义
// 已删除所有错误字段：StatDurationMs / CountMode / MaxBurstMultiplier
type SentinelHotSpotRule struct {
	ID                string                `mapstructure:"id"`
	Resource          string                `mapstructure:"resource"`
	MetricType        int                   `mapstructure:"metric_type"` // 0:QPS 1:并发
	ControlBehavior   int                   `mapstructure:"control_behavior"`
	ParamIndex        int                   `mapstructure:"param_index"`
	ParamKey          string                `mapstructure:"param_key"`
	Threshold         int64                 `mapstructure:"threshold"` // 官方类型 int64
	MaxQueueingTimeMs int64                 `mapstructure:"max_queueing_time_ms"`
	BurstCount        int64                 `mapstructure:"burst_count"`
	DurationInSec     int64                 `mapstructure:"duration_in_sec"` // 官方单位：秒，int64
	ParamsMaxCapacity int64                 `mapstructure:"params_max_capacity"`
	SpecificItems     map[interface{}]int64 `mapstructure:"specific_items"`
}

// 3. 熔断降级规则 对齐 circuitbreaker.Rule 官方定义
type SentinelCircuitRule struct {
	ID               string  `mapstructure:"id"`
	Resource         string  `mapstructure:"resource"`
	Strategy         int     `mapstructure:"strategy"`
	Threshold        float64 `mapstructure:"threshold"`
	RetryTimeoutMs   uint32  `mapstructure:"retry_timeout_ms"` // 官方类型 uint32，单位毫秒
	MinRequestAmount uint64  `mapstructure:"min_request_amount"`
	StatIntervalMs   uint32  `mapstructure:"stat_interval_ms"`  // 官方类型 uint64
	MaxAllowedRtMs   uint64  `mapstructure:"max_allowed_rt_ms"` // 慢调用比例策略必填
}

type SentinelSystemRule struct {
	TriggerCount float64 `mapstructure:"trigger_count"`
}

// API网关配置
type GatewayConfig struct {
	Enable      bool              `mapstructure:"enable"`
	Port        int               `mapstructure:"port"`
	ContextPath string            `mapstructure:"context_path"`
	Cors        CorsConfig        `mapstructure:"cors"`
	Auth        GatewayAuthConfig `mapstructure:"auth"`
	IPBlacklist []string          `mapstructure:"ip_blacklist"`
	Routes      []GatewayRoute    `mapstructure:"routes"` // 路由列表
}

type CorsConfig struct {
	Enable           bool     `mapstructure:"enable"`
	AllowOrigins     []string `mapstructure:"allow_origins"`
	AllowMethods     []string `mapstructure:"allow_methods"`
	AllowHeaders     []string `mapstructure:"allow_headers"`
	AllowCredentials bool     `mapstructure:"allow_credentials"`
	MaxAge           int      `mapstructure:"max_age"` // 预检缓存时间(秒)
}

type GatewayAuthConfig struct {
	Enable       bool     `mapstructure:"enable"`
	Type         string   `mapstructure:"type"` // jwt/api_key
	JwtSecret    string   `mapstructure:"jwt_secret"`
	JwtExpire    int      `mapstructure:"jwt_expire"`    // 小时
	ExcludePaths []string `mapstructure:"exclude_paths"` // 免鉴权路径
}

type GatewayRoute struct {
	ID          string `mapstructure:"id"`
	Path        string `mapstructure:"path"`         // 网关路径
	TargetURL   string `mapstructure:"target_url"`   // 转发目标（直连）
	ServiceName string `mapstructure:"service_name"` // 转发目标（服务发现）
	StripPrefix bool   `mapstructure:"strip_prefix"` // 是否移除路径前缀
	AuthEnable  bool   `mapstructure:"auth_enable"`  // 该路由是否鉴权
	RateLimit   bool   `mapstructure:"rate_limit"`   // 该路由是否限流
}

// 可观测性配置
type ObservabilityConfig struct {
	Tracer  TracerConfig  `mapstructure:"tracer"`
	Metrics MetricsConfig `mapstructure:"metrics"`
}

type TracerConfig struct {
	Enable      bool   `mapstructure:"enable"`
	Endpoint    string `mapstructure:"endpoint"` // Jaeger 采集地址
	ServiceName string `mapstructure:"service_name"`
}

type MetricsConfig struct {
	Enable bool   `mapstructure:"enable"`
	Path   string `mapstructure:"path"` // 指标暴露路径
	Port   int    `mapstructure:"port"` // 指标端口
}
