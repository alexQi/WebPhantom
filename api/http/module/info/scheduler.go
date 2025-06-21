package info

import (
	"github.com/kataras/iris/v12"
	"noctua/api/http/controller"
)

type SchedulerController struct {
	controller.BaseController
}

func (c *SchedulerController) Pause(ctx iris.Context) error {
	data := map[string]interface{}{
		"code": 0,
		"msg":  "success",
	}
	c.Kernel.Scheduler.Pause()
	return ctx.JSON(data)
}

func (c *SchedulerController) Resume(ctx iris.Context) error {
	data := map[string]interface{}{
		"code": 0,
		"msg":  "success",
	}
	c.Kernel.Scheduler.Resume()
	return ctx.JSON(data)
}

func (c *SchedulerController) TaskTree(ctx iris.Context) error {
	data := map[string]interface{}{
		"code": 0,
		"msg":  "success",
		"data": c.Kernel.Scheduler.GetTaskTree(),
	}

	return ctx.JSON(data)
}
