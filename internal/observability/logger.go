package observability

import (
	"os"
	"time"
	"uy_micro/config"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

func InitLogger(cfg *config.LoggerConfig) (*zap.Logger, error) {
	// 日志级别
	level := zap.InfoLevel
	_ = level.UnmarshalText([]byte(cfg.Level))

	// 加载上海时区
	shanghaiLoc, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		return nil, err
	}

	// 自定义本地时间格式化函数
	localTimeEncoder := func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
		enc.AppendString(t.In(shanghaiLoc).Format("2006-01-02T15:04:05.000"))
	}

	// 编码器配置
	encoderCfg := zap.NewProductionEncoderConfig()
	encoderCfg.TimeKey = "time"
	// 替换UTC编码器为本地北京时间
	encoderCfg.EncodeTime = localTimeEncoder
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
