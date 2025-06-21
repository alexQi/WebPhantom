package errors

import (
	"github.com/kataras/golog"
	"github.com/kataras/iris/v12"
)

func ErrorHandle(app *iris.Application) {
	httpError := HttpError{}

	httpError.logger = app.Logger()
	app.OnErrorCode(iris.StatusNotFound, httpError.notFound)
	app.OnErrorCode(iris.StatusInternalServerError, httpError.internalServerError)
}

type HttpError struct {
	logger *golog.Logger
}

// notFound 404
func (h *HttpError) notFound(ctx iris.Context) {
	h.logger.Warningf("route not found : %s", ctx.Request().RequestURI)
	op := output.Output{Ctx: ctx}
	op.Code = 404
	op.Error("404 not found :" + ctx.Request().RequestURI)
}

// internalServerError 服务器错误
func (h *HttpError) internalServerError(ctx iris.Context) {
	h.logger.Warningf("server error : %s", ctx.Request().RequestURI)
	op := output.Output{Ctx: ctx}
	op.Code = 500
	op.Error("Oops something went wrong, try again")
}

// AuthError 认证失败
func (h *HttpError) AuthError(ctx iris.Context, msg string) {
	op := output.Output{Ctx: ctx}
	op.Code = 401

	if msg == "" {
		op.Error("Oops auth fail")
	} else {
		op.Error(msg)
	}
}
