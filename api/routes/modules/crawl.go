package modules

import (
	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/core/router"
	"noctua/api/http/module/crawl"
	"noctua/kernel"
)

func CrawlRoutes(app router.Party, kernel *kernel.Kernel) {
	c := crawl.CrawlController{}
	c.SetKernel(kernel)
	app.Get("/status", func(ctx iris.Context) {
		_ = c.Status(ctx)
	})
	app.Post("/start", func(ctx iris.Context) {
		_ = c.Start(ctx)
	})
	app.Get("/stop", func(ctx iris.Context) {
		_ = c.Stop(ctx)
	})
}
