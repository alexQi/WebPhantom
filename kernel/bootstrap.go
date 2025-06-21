package kernel

import (
	"github.com/spf13/viper"
	"noctua/internal/model"
	"noctua/internal/proxy"
	"noctua/internal/scheduler"
	"noctua/pkg/database"
	"noctua/pkg/logger"
)

func LoadConfig() KernelConfig {
	signEndpoint := viper.GetString("SIGN_SERVER_ENDPOINT")

	// 爬虫管理配置
	crawlerConfig := CrawlerManagerConfig{
		SignServEndpoint: signEndpoint,
	}

	schedulerConfig := scheduler.Config{
		MaxWorkersPerQueue: viper.GetInt("scheduler.max_workers_per_queue"),
		WorkerIdleTimeout:  viper.GetDuration("scheduler.worker_idle_timeout"),
		AutoScaleInterval:  viper.GetDuration("scheduler.auto_scale_interval"),
		MaxQueueDepth:      viper.GetInt("scheduler.max_queue_depth"),
		BaseRetryDelay:     viper.GetDuration("scheduler.base_retry_delay"),
		DefaultQPS:         viper.GetInt("scheduler.default_qps"),
	}

	// 代理配置
	proxyPoolConfig := proxy.ProxyPoolConfig{
		MinDynamic: 1,
		MinStatic:  1,
	}
	return KernelConfig{
		SchedulerConfig: schedulerConfig,
		ProxyConfig:     proxyPoolConfig,
		CrawlerConfig:   crawlerConfig,
	}
}

func MigrateModels() {
	// 在 GetDB 中调用 Migrate，确保初始化的同时完成迁移
	if err := database.Migrate(database.DB, []interface{}{
		&model.MediaAccount{},
		&model.CrawlTask{},
		&model.CrawlMedia{},
		&model.CrawlComment{},
		&model.CrawlUser{},
	}); err != nil {
		logger.Log.Errorf("Initial migration failed: %v", err)
	}
}
