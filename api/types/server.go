package types

import (
	"github.com/kataras/iris/v12"
	"noctua/kernel"
)

type Server struct {
	app    *iris.Application
	kernel *kernel.Kernel
}
