package uy_micro

import "uy_micro/bootstrap"

// Server 框架全局入口实例
var Server = &microServer{}

type microServer struct{}

// Run 一键启动微服务框架
// 使用者仅需调用 uy_micro.Server.Run()
func (s *microServer) Run() {
	if err := bootstrap.Bootstrap(); err != nil {
		panic("uy_micro start failed: " + err.Error())
	}
	bootstrap.Run()
}

// OnBootstrap 注册启动后钩子
// 用于业务层注册路由、gRPC 服务、自定义初始化逻辑
func (s *microServer) OnBootstrap(fn func()) {
	bootstrap.PostBootHooks = append(bootstrap.PostBootHooks, fn)
}
