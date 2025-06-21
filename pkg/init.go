package pkg

import (
	"github.com/spf13/viper"
	"math/rand"
	"noctua/pkg/cache"
	"noctua/pkg/config"
	"noctua/pkg/database"
	"noctua/pkg/logger"
	"noctua/pkg/utils/file"
	"path/filepath"
	"time"
)

func init() {
	//
	rand.Seed(time.Now().UnixNano())
	// 设置时区为CST 东八区
	time.Local = time.FixedZone("CST", 8*3600) // 东八
	// 初始化config
	config.Init(&config.ConfigSetting{
		YamlPath: file.GetResourcePath("config/"),
		EnvPath:  file.GetResourcePath(".env"),
	})
	// 获取runtime目录
	runtimePath := file.GetRuntimeDir()
	// 初始化日志
	logger.Init(&logger.LoggerConfig{
		Stdout: viper.GetBool("log.stdout"),
		Level:  viper.GetString("log.level"),
		Path:   filepath.Join(runtimePath, "logs"),
	})
	// 初始化缓存
	cache.NewCache(&cache.CacheConfig{
		CacheType:   viper.GetString("cache.type"),
		ExpireTime:  viper.GetDuration("cache.expire"),
		Cleanup:     viper.GetDuration("cache.cleanup"),
		PersistFile: filepath.Join(runtimePath, "cache/data.gob"),
	})
	dbDriver := viper.GetString("db.driver")
	if len(dbDriver) == 0 {
		dbDriver = "sqlite"
	}
	dsn := viper.GetString("db.dsn")
	if dbDriver == "sqlite" {
		dsn = filepath.Join(runtimePath, "data/noctua.db")
	}
	// 初始化数据库
	database.GetDB(
		&database.Config{
			Driver:   dbDriver,                 // "mysql" / "postgres" / "sqlite"
			DSN:      dsn,                      // 连接字符串
			LogLevel: viper.GetInt("db.level"), // "info" / "warn" / "error"
		},
	)
}
