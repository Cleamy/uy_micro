package observability

import (
	"os"
	"uy_micro/config"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

func InitLogger(cfg *config.LoggerConfig) (*zap.Logger, error) {
	// 日志级别
	level := zap.InfoLevel
	_ = level.UnmarshalText([]byte(cfg.Level))

	// 编码器配置
	encoderCfg := zap.NewProductionEncoderConfig()
	encoderCfg.TimeKey = "time"
	encoderCfg.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderCfg.EncodeLevel = zapcore.CapitalLevelEncoder

	var encoder zapcore.Encoder
	if cfg.Format == "json" {
		encoder = zapcore.NewJSONEncoder(encoderCfg)
	} else {
		encoder = zapcore.NewConsoleEncoder(encoderCfg)
	}

	var cores []zapcore.Core
	// 控制台输出
	cores = append(cores, zapcore.NewCore(encoder, zapcore.AddSync(os.Stdout), level))

	// 文件输出（配置了路径才启用）
	if cfg.FilePath != "" {
		writer := &lumberjack.Logger{
			Filename:   cfg.FilePath,
			MaxSize:    cfg.MaxSize,
			MaxBackups: cfg.MaxBackups,
			MaxAge:     cfg.MaxAge,
			Compress:   cfg.Compress,
		}
		cores = append(cores, zapcore.NewCore(
			zapcore.NewJSONEncoder(encoderCfg),
			zapcore.AddSync(writer),
			level,
		))
	}

	core := zapcore.NewTee(cores...)
	return zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1)), nil
}
