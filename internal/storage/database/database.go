package database

import (
	"fmt"
	"time"
	"uy_micro/config"

	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

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
		Logger: logger.Default.LogMode(logger.Info),
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
