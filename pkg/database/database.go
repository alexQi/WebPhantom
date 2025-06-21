package database

import (
	"fmt"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	glogger "gorm.io/gorm/logger"
	"noctua/pkg/logger"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// DB 全局数据库实例
var DB *gorm.DB
var once sync.Once

// InitDB 初始化数据库连接
func InitDB(config *Config) {
	once.Do(func() {
		var dialector gorm.Dialector
		switch config.Driver {
		case "mysql":
			dialector = mysql.Open(config.DSN)
		case "postgres":
			dialector = postgres.Open(config.DSN)
		case "sqlite":
			logger.Log.Infof("Init sqlite database, dns %s", config.DSN)
			if err := ensureSQLitePath(config.DSN); err != nil {
				logger.Log.Errorf("Initialize SQLite database failed, path: %s, error: %v", config.DSN, err)
				return
			}
			dialector = sqlite.Open(config.DSN)
		default:
			logger.Log.Errorf("Unsupported database type: %s", config.Driver)
			return
		}

		// 自定义 GORM 日志配置，仅记录 Error 级别
		gormLogger := glogger.New(
			logger.Log, // 使用 noctua/pkg/logger 作为输出
			glogger.Config{
				SlowThreshold:             200 * time.Millisecond,
				IgnoreRecordNotFoundError: true,
				Colorful:                  false,
				LogLevel:                  glogger.LogLevel(config.LogLevel),
				//LogLevel: glogger.Info,
			},
		)

		// 连接数据库
		db, err := gorm.Open(dialector, &gorm.Config{
			Logger: gormLogger,
		})
		if err != nil {
			logger.Log.Errorf("Connect database failed: %v", err)
			return
		}

		DB = db
		logger.Log.Infof("Connect database success: %s", config.Driver)
	})
}

// Migrate 执行数据库迁移
func Migrate(db *gorm.DB, models []interface{}) error {
	if db == nil {
		return fmt.Errorf("database instance is nil")
	}

	if err := db.AutoMigrate(models...); err != nil {
		logger.Log.Errorf("Migration failed: %v", err)
		return err
	}
	logger.Log.Info("Migration success")
	return nil
}

// GetDB 获取数据库实例
func GetDB(config *Config) *gorm.DB {
	if DB == nil {
		InitDB(config)
	}
	return DB
}

// ensureSQLitePath 确保 SQLite 数据库文件所在的路径存在
func ensureSQLitePath(dsn string) error {
	filePath := dsn
	if !filepath.IsAbs(filePath) { // 使用 filepath.IsAbs 替代自定义函数
		wd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("Error getting current working directory: %v", err)
		}
		filePath = filepath.Join(wd, filePath) // 使用 filepath.Join 拼接路径
	}

	// 提取目录部分
	dir := filepath.Dir(filePath)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		err := os.MkdirAll(dir, os.ModePerm)
		if err != nil {
			return fmt.Errorf("Failed to create database directory: %v", err)
		}
	}

	// 检查并创建数据库文件
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		_, err := os.Create(filePath)
		if err != nil {
			return fmt.Errorf("Failed to create db file: %v", err)
		}
	}
	return nil
}
