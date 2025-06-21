package modules

import (
	"github.com/kataras/iris/v12"
	"noctua/api/core/output"
)

func BaseRoutes(app *iris.Application) {
	app.Get("/", func(ctx iris.Context) {
		o := output.Output{Ctx: ctx}
		o.Success("welcome to noctua.co")
	})
}
