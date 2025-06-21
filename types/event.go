package types

import (
	"time"
)

type CrawlEndCode int

const (
	CrawlEndCodeNilSession CrawlEndCode = 10
	CrawlEndCodeReachClean CrawlEndCode = 20
	CrawlEndCodeOverdLimit CrawlEndCode = 30
	CrawlEndCodeVerifyFail CrawlEndCode = 40
	CrawlEndCodeForcedStop CrawlEndCode = 50
	CrawlEndCodeRoundMaxed CrawlEndCode = 60
)

// 单个清洗完成事件
type CrawlStartEvent struct {
	*CrawlParams
}

// 单个清洗完成事件
type CrawlStopEvent struct {
	ReachedAt time.Time
}

// 采集结束事件
type CrawlEndEvent struct {
	Code      CrawlEndCode
	ReceiveAt time.Time
}
