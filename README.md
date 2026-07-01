# uy_micro 封装框架

## 封装思想

基于gin、gorm、sentinel、consul、grpc实现
简化配置文件，为微服务子集提供通用基础的配置    
将通用配置文件config.yaml 自动配置相关
并提供生成后的global全局变量控制


alpha 版本实现内容

包含
微服务的注册、发现、负载均衡
外部统一gateway入口
gateway实现：跨域检测、黑名单（公网ip地址）
rpc：使用google的 grpc（已经内置了grpc工厂实现grpc的负载均衡功能，具体负载模式通过yml文件配置），【rpc-http、dubbo、feign、webservice：暂不支持】
服务发现与注册：consul、【nacos、ectd：由服务问题暂不提供】
限流、熔断、降级：sentinel（其他方案暂未提供）
auth：内置jwt、（outh2暂未适配）
消息通信：内置rabbitmq（kafka、rocket后续）
持久层Storage：gorm框架（适配mysql、postgresql两种数据库，【sqlit、MongoDB暂不提供】）
路由：gin框架（后续更具性能要求适配fasthttp、echo、或hertx）
埋点：后续有需求
可视化数据矩阵：ing~~~~ 适配dev ing prometheus
缓存数据：redis
日志：zap框架


容器化：docker 暂时为提供
容器化管理：k8s 后续


```txt
./uy_micro
├── application.example.yml             给用户参考的config 文件
├── bootstrap                           启动文件类夹
│   └── bootstrap.go                        启动器文件                 
├── config                              配置文件夹
│   ├── config.go                           配置类对象
│   └── loader.go                           配置加载器
├── global                              全局变量存储
│   └── global.go                           全局变量
├── go.mod              
├── go.sum
├── internal                            内部文件夹
│   ├── gateway                             网关文件夹
│   │   ├── engine.go                           引擎文件
│   │   ├── filter                              过滤器文件夹
│   │   │   ├── auth.go                             认证
│   │   │   ├── cors.go                             跨域
│   │   │   ├── ip_bloacklist.go                    黑名单
│   │   │   ├── logger.go                           日志配置
│   │   │   ├── rate_limit.go                       限速
│   │   │   └── sentinel_limit.go                   sentinel 配置
│   │   └── router_loader.go                    路由加载器
│   ├── grpc                                grpc文件夹
│   │   ├── client.go                           客户端
│   │   └── server.go                           服务端
│   ├── mq                                  中间件MQ消息通信
│   │   └── rabbitmq.go                         rabbitmq配置
│   ├── observability                       可观察文件夹
│   │   ├── logger.go                           日志配置
│   │   ├── metrics.go                          
│   │   └── tracer.go                           跟踪
│   ├── registry                            注册器
│   │   └── consul.go                           consul服务注册
│   ├── storage                             存储
│   │   ├── database                            数据库
│   │   │   └── database.go                         数据库配置
│   │   └── redis                               redis缓存
│   │       └── redis.go                            redis配置
│   ├── traffic                             traffic 拥塞控制
│   │   └── seninel.go                          sentinel 配置
│   └── web                                 web服务端
│       ├── engine.go                           gin配置文件
│       └── middleware                          中间件
│           └── health.go                           健康监测请求
├── pkg                                 公共包
│   ├── jwt                                 jwt
│   │   └── jwt.go                          
│   ├── response                            响应文件价
│   └── utils                               工具文件夹
├── READMD.md
└── server.go                           服务文件
```


框架配置：
```bash
curl htpp://github.com/Cleamy/uy_micro/application.example.yml
```



配置文件根据需求所填写（如果不需要的服务不填写默认都为false
注册服务consul如不需，可不填默认为false

服务其实使用
uy_micro.Server.OnBootstrap() -- 为生命钩子函数在 bootstrap启动后执行
uy_mirco.Server.Run() -- 启动服务，run方法需要放在代码的最后，以防custom 配置为生效



代码案例
main.go
```golang 
func main() {
	uy_micro.Server.OnBootstrap(func() {
		// 1. 自动迁移表结构（PostgreSQL 兼容）
		_ = global.DB.AutoMigrate(&model.Role{})

		// 2. 依赖注入：业务层 → 接口层
		roleSvc := service.NewRoleService()
		roleController := controller.NewRoleController(roleSvc)
		roleGrpcSvc := service.NewRoleGrpcService(roleSvc)

		// 3. 注册 HTTP 路由
		registerHttpRoutes(global.Web, roleController)

		// 4. 注册 gRPC 服务到框架实例
		rolepb.RegisterRoleServiceServer(global.Grpc, roleGrpcSvc)
	})
	uy_micro.Server.Run()
}

func registerHttpRoutes(r *gin.Engine, c *controller.RoleController) {
	group := r.Group("/api/v1/role")
	{
		group.GET("/:id", c.GetByID)
		group.POST("", c.Create)
		group.GET("/list", c.List)
	}
}
```

框架默认提供global的一个全局变量，db、gin、consul、grpc、mq、redis、等实力均提供公开实例
如涉及pkg 文件 实例方法按需求使用

global package
```txt
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
```