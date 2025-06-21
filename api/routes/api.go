package routes

import (
	"github.com/kataras/iris/v12"
	"noctua/api/routes/modules"
	"noctua/kernel"
)

// ApiRoutes api路由加载
func ApiRoutes(app *iris.Application, kernel *kernel.Kernel) {
	// 默认路由
	modules.BaseRoutes(app)
	// Debug路由
	modules.PprofRoutes(app)
	// 存活探针
	healthGroup := app.Party("/health")
	{
		modules.HealthRoutes(healthGroup, kernel)
	}
	// 设置api分组
	infoGroup := app.Party("/v1/info")
	{
		modules.InfoRoutes(infoGroup, kernel)
	}
	// 调度器
	schedulerGroup := app.Party("v1/scheduler")
	{
		modules.SchedulerRoutes(schedulerGroup, kernel)
	}
	// 设置api分组
	crawlGroup := app.Party("/v1/crawl")
	{
		modules.CrawlRoutes(crawlGroup, kernel)
	}
}
