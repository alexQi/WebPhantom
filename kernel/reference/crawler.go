package reference

import (
	"noctua/internal/scheduler"
	"noctua/types"
)

// Crawler 爬虫接口，所有爬虫都必须实现
type Crawler interface {
	Initialize(scheduler *scheduler.Scheduler, runtimeChannel chan types.RuntimeData, channels map[string]chan types.FetchItemChan)
	HandleChannel(item types.FetchItemChan, params *types.CrawlParams) error
	SubmitJob(taskType string, payload interface{}, options scheduler.TaskOptions) error
}
