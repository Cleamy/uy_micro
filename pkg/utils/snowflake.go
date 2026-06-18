package utils

import (
	"errors"
	"sync"
	"time"
)

// 雪花 ID 配置（64位：1位符号 + 41位时间戳 + 10位机器ID + 12位序列号）
const (
	workerIDBits = 10
	sequenceBits = 12
	maxWorkerID  = -1 ^ (-1 << workerIDBits)
	maxSequence  = -1 ^ (-1 << sequenceBits)
	timeShift    = workerIDBits + sequenceBits
	workerShift  = sequenceBits
	epoch        = 1704067200000 // 2024-01-01 00:00:00 时间戳起点
)

// Snowflake 雪花 ID 生成器
type Snowflake struct {
	mu        sync.Mutex
	workerID  int64
	sequence  int64
	lastStamp int64
}

// NewSnowflake 创建雪花生成器，workerID 范围 0~1023
func NewSnowflake(workerID int64) (*Snowflake, error) {
	if workerID < 0 || workerID > maxWorkerID {
		return nil, errors.New("worker ID out of range (0~1023)")
	}
	return &Snowflake{workerID: workerID}, nil
}

// Generate 生成唯一 ID
func (s *Snowflake) Generate() int64 {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now().UnixMilli()
	if now == s.lastStamp {
		s.sequence = (s.sequence + 1) & maxSequence
		// 序列号用完，等待下一毫秒
		if s.sequence == 0 {
			for now <= s.lastStamp {
				now = time.Now().UnixMilli()
			}
		}
	} else {
		s.sequence = 0
	}
	s.lastStamp = now

	return (now-epoch)<<timeShift | s.workerID<<workerShift | s.sequence
}
