package controller

import (
	"github.com/kataras/iris/v12"
	"noctua/api/core/output"
	"noctua/kernel"
	"strings"
)

type BaseController struct {
	Kernel *kernel.Kernel
	Output output.Output
}

// Init 初始化相关controller需要的数据
func (b *BaseController) Init(ctx iris.Context) {
	b.Output = output.Output{Ctx: ctx}
}

func (b *BaseController) SetKernel(kernel *kernel.Kernel) {
	b.Kernel = kernel
}

// ClientIP 获取ip
func (b *BaseController) ClientIP(ctx iris.Context) string {
	var ip string
	ip = strings.TrimSpace(ctx.GetHeader("X-Forwarded-For"))
	if ip != "" {
		return ip
	}
	ip = strings.TrimSpace(ctx.GetHeader("X-Real-Ip"))
	if ip != "" {
		return ip
	}
	ip = strings.TrimSpace(ctx.RemoteAddr())
	if ip != "" {
		return ip
	}

	return "0.0.0.0"
}
