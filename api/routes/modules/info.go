package modules

import (
	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/core/router"
	"noctua/api/http/module/info"
	"noctua/kernel"
)

func InfoRoutes(app router.Party, kernel *kernel.Kernel) {
	c := info.InfoController{}
	c.SetKernel(kernel)
	app.Get("/pushRuntime", func(ctx iris.Context) {
		_ = c.PushRuntime(ctx)
	})
	app.Get("/sessionStatus", func(ctx iris.Context) {
		_ = c.SessionStatus(ctx)
	})
}
