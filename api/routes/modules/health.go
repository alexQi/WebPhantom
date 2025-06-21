package modules

import (
	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/core/router"
	"noctua/api/http/controller"
	"noctua/kernel"
)

func HealthRoutes(app router.Party, kernel *kernel.Kernel) {
	c := controller.HealthController{}
	c.SetKernel(kernel)
	app.Get("/live", func(ctx iris.Context) {
		c.Live(ctx)
	})
}
