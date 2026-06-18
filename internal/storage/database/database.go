package database

import (
	"context"
	"fmt"
	"time"
	"uy_micro/config"
	"uy_micro/global"

	"go.uber.org/zap"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// gormZapLogger 自定义 GORM 日志，对接 Zap 结构化日志
type gormZapLogger struct {
	SlowThreshold time.Duration
}

func (l *gormZapLogger) LogMode(level logger.LogLevel) logger.Interface {
	return l
}

func (l *gormZapLogger) Info(_ context.Context, msg string, data ...interface{}) {
	global.Logger.Info(fmt.Sprintf(msg, data...))
}

func (l *gormZapLogger) Warn(_ context.Context, msg string, data ...interface{}) {
	global.Logger.Warn(fmt.Sprintf(msg, data...))
}

func (l *gormZapLogger) Error(_ context.Context, msg string, data ...interface{}) {
	global.Logger.Error(fmt.Sprintf(msg, data...))
}

func (l *gormZapLogger) Trace(_ context.Context, begin time.Time, fc func() (sql string, rowsAffected int64), err error) {
	cost := time.Since(begin)
	sql, rows := fc()

	if err != nil {
		global.Logger.Error("gorm sql error",
			zap.String("sql", sql),
			zap.Int64("rows", rows),
			zap.Duration("cost", cost),
			zap.Error(err))
		return
	}

	if cost > l.SlowThreshold {
		global.Logger.Warn("gorm slow sql",
			zap.String("sql", sql),
			zap.Int64("rows", rows),
			zap.Duration("cost", cost))
		return
	}

	global.Logger.Debug("gorm sql",
		zap.String("sql", sql),
		zap.Int64("rows", rows),
		zap.Duration("cost", cost))
}

func Init(cfg *config.DatabaseConfig) (*gorm.DB, error) {
	if !cfg.Enable {
		return nil, nil
	}

	var dialector gorm.Dialector
	dsn := buildDSN(cfg.Primary)

	switch cfg.Primary.Driver {
	case "mysql":
		dialector = mysql.Open(dsn)
	case "postgres", "postgresql":
		dialector = postgres.Open(dsn)
	default:
		return nil, fmt.Errorf("unsupported database driver: %s", cfg.Primary.Driver)
	}

	db, err := gorm.Open(dialector, &gorm.Config{
		Logger: &gormZapLogger{SlowThreshold: 200 * time.Millisecond},
	})
	if err != nil {
		return nil, fmt.Errorf("open database failed: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("get sql db failed: %w", err)
	}

	sqlDB.SetMaxIdleConns(cfg.Primary.MaxIdle)
	sqlDB.SetMaxOpenConns(cfg.Primary.MaxOpen)
	sqlDB.SetConnMaxLifetime(time.Duration(cfg.Primary.MaxLifeTime) * time.Second)

	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("ping database failed: %w", err)
	}
	return db, nil
}

// 根据驱动拼接不同的 DSN
func buildDSN(cfg config.DatasourceConfig) string {
	switch cfg.Driver {
	case "mysql":
		return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=True&loc=Local",
			cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.DBName, cfg.Charset)
	case "postgres", "postgresql":
		return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable TimeZone=Asia/Shanghai",
			cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName)
	default:
		return ""
	}
}
