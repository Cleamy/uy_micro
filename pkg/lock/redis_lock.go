package lock

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"uy_micro/global"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

const (
	defaultLockExpire   = 30 * time.Second // 默认锁过期时间
	defaultWatchDogTick = 10 * time.Second // 看门狗续期间隔
)

// RedisLock 基于Redis的可重入分布式锁
type RedisLock struct {
	key            string        // 锁的Redis键
	holderID       string        // 持有者唯一标识（锁内唯一）
	expire         time.Duration // 锁过期时间
	watchDogCtx    context.Context
	watchDogCancel context.CancelFunc
	running        bool
	mu             sync.Mutex // 本地互斥，保护重入计数
	reentrant      int        // 重入计数
}

// NewRedisLock 创建分布式锁实例
func NewRedisLock(key string) *RedisLock {
	return &RedisLock{
		key:      fmt.Sprintf("lock:%s", key),
		holderID: uuid.New().String(),
		expire:   defaultLockExpire,
	}
}

// TryLock 尝试加锁，非阻塞，立即返回结果
func (l *RedisLock) TryLock(ctx context.Context) (bool, error) {
	if global.Redis == nil {
		return false, errors.New("redis is not enabled")
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	// 可重入：当前持有者已持有锁，计数+1
	if l.running {
		l.reentrant++
		return true, nil
	}

	// 原子加锁：SET NX PX
	ok, err := global.Redis.SetNX(ctx, l.key, l.holderID, l.expire).Result()
	if err != nil {
		return false, err
	}
	if !ok {
		return false, nil
	}

	// 加锁成功，启动看门狗续期
	l.running = true
	l.reentrant = 1
	l.startWatchDog()

	return true, nil
}

// Lock 阻塞加锁，支持超时；超时返回错误
func (l *RedisLock) Lock(ctx context.Context, timeout time.Duration) error {
	deadline := time.After(timeout)
	ticker := time.NewTicker(50 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-deadline:
			return errors.New("lock acquire timeout")
		case <-ticker.C:
			ok, err := l.TryLock(ctx)
			if err != nil {
				return err
			}
			if ok {
				return nil
			}
		}
	}
}

// Unlock 释放锁
func (l *RedisLock) Unlock() error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if !l.running {
		return errors.New("lock is not held")
	}

	// 重入计数-1，未到0不真正释放
	l.reentrant--
	if l.reentrant > 0 {
		return nil
	}

	// 停止看门狗
	l.stopWatchDog()
	l.running = false

	// 原子释放：校验持有者再删除，避免误删别人的锁
	luaScript := `
	if redis.call("GET", KEYS[1]) == ARGV[1] then
		return redis.call("DEL", KEYS[1])
	else
		return 0
	end
	`
	_, err := global.Redis.Eval(context.Background(), luaScript, []string{l.key}, l.holderID).Result()
	if err != nil && !errors.Is(err, redis.Nil) {
		return err
	}
	return nil
}

// startWatchDog 启动看门狗，自动续期
func (l *RedisLock) startWatchDog() {
	l.watchDogCtx, l.watchDogCancel = context.WithCancel(context.Background())

	go func() {
		ticker := time.NewTicker(defaultWatchDogTick)
		defer ticker.Stop()

		for {
			select {
			case <-l.watchDogCtx.Done():
				return
			case <-ticker.C:
				// 续期：校验持有者并刷新过期时间
				luaScript := `
				if redis.call("GET", KEYS[1]) == ARGV[1] then
					return redis.call("EXPIRE", KEYS[1], ARGV[2])
				else
					return 0
				end
				`
				expireSec := int64(l.expire.Seconds())
				_, _ = global.Redis.Eval(l.watchDogCtx, luaScript, []string{l.key}, l.holderID, expireSec).Result()
			}
		}
	}()
}

// stopWatchDog 停止看门狗
func (l *RedisLock) stopWatchDog() {
	if l.watchDogCancel != nil {
		l.watchDogCancel()
	}
}
