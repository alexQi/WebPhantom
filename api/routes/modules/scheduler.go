package modules

import (
	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/core/router"
	"noctua/api/http/module/info"
	"noctua/kernel"
)

func SchedulerRoutes(app router.Party, kernel *kernel.Kernel) {
	c := info.SchedulerController{}
	c.SetKernel(kernel)
	app.Get("/pause", func(ctx iris.Context) {
		_ = c.Pause(ctx)
	})
	app.Get("/resume", func(ctx iris.Context) {
		_ = c.Resume(ctx)
	})
	app.Get("/taskTree", func(ctx iris.Context) {
		_ = c.TaskTree(ctx)
	})
}
