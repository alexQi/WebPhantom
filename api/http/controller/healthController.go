package controller

import (
	"github.com/kataras/iris/v12"
)

type HealthController struct {
	BaseController
}

// Live 存活探针
func (c *HealthController) Live(ctx iris.Context) {
	c.Init(ctx)
	c.Output.Success("service alive")
}
