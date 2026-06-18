package configwatch

import (
	"sync"
	"time"

	"uy_micro/config"
	"uy_micro/global"

	"github.com/hashicorp/consul/api"
	"go.uber.org/zap"
)

// ConfigChangeCallback 配置变更回调
type ConfigChangeCallback func(oldConf, newConf *config.AppConfig)

var (
	consulCli *api.Client
	configKey string
	mu        sync.Mutex
	callbacks []ConfigChangeCallback
	oldConfig *config.AppConfig
	stopWatch chan struct{}
	watchRun  bool
)

// Init 启动consul KV配置监听
// cli: consul客户端实例
// conf: 当前全局配置快照
// kvKey: 配置在consul中的kv路径
func Init(cli *api.Client, conf *config.AppConfig, kvKey string) error {
	mu.Lock()
	defer mu.Unlock()
	if watchRun {
		return nil
	}

	consulCli = cli
	configKey = kvKey
	oldConfig = conf
	stopWatch = make(chan struct{})
	watchRun = true

	// 异步启动监听协程
	go watchLoop()
	global.Logger.Info("consul config watch started, key: " + kvKey)
	return nil
}

// RegisterCallback 注册配置变更回调
func RegisterCallback(cb ConfigChangeCallback) {
	mu.Lock()
	defer mu.Unlock()
	callbacks = append(callbacks, cb)
}

// Close 停止监听，优雅关停调用
func Close() {
	mu.Lock()
	defer mu.Unlock()
	if !watchRun {
		return
	}
	close(stopWatch)
	watchRun = false
	global.Logger.Info("consul config watch stopped")
}

// 循环watch consul kv变更
func watchLoop() {
	kv := consulCli.KV()
	opts := &api.QueryOptions{
		WaitTime: 10 * time.Second, // consul阻塞长轮询
	}

	for {
		select {
		case <-stopWatch:
			return
		default:
		}

		pair, meta, err := kv.Get(configKey, opts)
		if err != nil {
			global.Logger.Error("consul kv get config failed", zap.Error(err))
			time.Sleep(2 * time.Second)
			continue
		}
		// WaitIndex不变代表无更新
		// WaitIndex 不变且无新 pair → 无实质变更（结合 pair==nil 更安全）
		if opts.WaitIndex == meta.LastIndex && pair != nil {
			continue
		}
		opts.WaitIndex = meta.LastIndex // 下次阻塞用这个值
		// 配置为空跳过
		if pair == nil || len(pair.Value) == 0 {
			continue
		}

		// 解析新配置
		newConf, err := parseConfigContent(pair.Value)
		if err != nil {
			global.Logger.Error("unmarshal consul config yaml fail", zap.Error(err))
			continue
		}

		// 执行更新逻辑
		mu.Lock()
		global.Config = newConf
		for _, cb := range callbacks {
			cb(oldConfig, newConf)
		}
		oldConfig = newConf
		mu.Unlock()
		global.Logger.Info("consul config hot reload success")
	}
}

// 解析yaml配置字节
func parseConfigContent(content []byte) (*config.AppConfig, error) {
	return config.LoadFromBytes(content)
}
