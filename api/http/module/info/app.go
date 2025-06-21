package info

import (
	"github.com/kataras/iris/v12"
	"noctua/api/http/controller"
	"noctua/pkg/utils/encrypt"
	"noctua/types"
	"time"
)

type InfoController struct {
	controller.BaseController
}

func (c *InfoController) PushRuntime(ctx iris.Context) error {
	data := map[string]interface{}{
		"code": 0,
		"msg":  "success",
	}
	c.Kernel.RuntimeChannel <- types.NewRuntimeData(
		types.RuntimeEventCodeNotification,
		types.EventData{
			Title:     "洞察中心",
			CheckHash: encrypt.Md5(time.Now().Format("2006-01-02 15:04:05")),
			Message:   "数据账号超频，请尝试更换账号后再次启动...",
			Optional: types.MessageOptional{
				IsNotify: true,
				IsStore:  true,
				ShowType: "modal",
			},
		},
	)
	return ctx.JSON(data)
}

// Status  监测投放数据
func (c *InfoController) SessionStatus(ctx iris.Context) error {
	data := c.Kernel.SessionManager.Status()
	return ctx.JSON(data)
}
