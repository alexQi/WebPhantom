package middlewares

import (
	"github.com/kataras/iris/v12/context"
)

func AppNew() context.Handler {
	return func(ctx *context.Context) {
		ctx.Next()
	}
}
