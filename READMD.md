# uy_micro 封装框架

## 封装思想

基于gin、gorm、sentinel、consul、grpc实现
简化配置文件，为微服务子集提供通用基础的配置    
将通用配置文件config.yaml 自动配置相关
并提供生成后的global全局变量控制


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
