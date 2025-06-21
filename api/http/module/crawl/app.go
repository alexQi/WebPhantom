package crawl

import (
	"fmt"
	"github.com/kataras/iris/v12"
	"noctua/api/http/controller"
	"noctua/pkg/logger"
	"noctua/types"
)

type CrawlController struct {
	controller.BaseController
}

// Status  监测投放数据
func (c *CrawlController) Status(ctx iris.Context) error {
	data := c.Kernel.CrawlerManager.Status()

	return ctx.JSON(data)
}

// Status  监测投放数据
func (c *CrawlController) Start(ctx iris.Context) error {
	data := map[string]interface{}{
		"code": 0,
		"msg":  "success",
	}
	crawlParams := &types.CrawlParams{}
	err := ctx.ReadJSON(crawlParams)
	if err != nil {
		return ctx.JSON(map[string]interface{}{
			"code": 200,
			"msg":  fmt.Sprintf("Unmarshal request params failed: %s", err.Error()),
		})
	}
	if crawlParams.MediaCode == "" {
		return ctx.JSON(map[string]interface{}{
			"code": 200,
			"msg":  "Media code can not be none",
		})
	}
	if crawlParams.CrawlType == "" {
		return ctx.JSON(map[string]interface{}{
			"code": 200,
			"msg":  "Crawl type can not be none",
		})
	}
	if crawlParams.Region == "" {
		return ctx.JSON(map[string]interface{}{
			"code": 200,
			"msg":  "Region can not be none",
		})
	}
	if len(crawlParams.Keywords) == 0 {
		return ctx.JSON(map[string]interface{}{
			"code": 200,
			"msg":  "Keywords can not be none",
		})
	}

	go func() {
		err := c.Kernel.CrawlerManager.Run(crawlParams)
		if err != nil {
			logger.Log.Errorf("crawler manager run failed: %s", err.Error())
			return
		}
	}()

	return ctx.JSON(data)
}

func (c *CrawlController) Stop(ctx iris.Context) error {
	data := map[string]interface{}{
		"code": 0,
		"msg":  "success",
	}
	c.Kernel.Scheduler.Shutdown()
	return ctx.JSON(data)
}
