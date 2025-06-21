package kernel

import (
	"context"
	"errors"
	"fmt"
	"noctua/internal/constants"
	"noctua/internal/scheduler"
	"noctua/internal/signer"
	"noctua/kernel/bus"
	"noctua/kernel/crawls/douyin"
	"noctua/kernel/reference"
	"noctua/kernel/session"
	"noctua/pkg/logger"
	"noctua/pkg/utils/encrypt"
	"noctua/pkg/utils/str"
	"noctua/types"
	"sync"
	"sync/atomic"
	"time"
)

const ROUND_MAX = 10
const ROUND_SLEEP = 20
const ROUND_INCR = 48

// CrawlerStatus 定义爬虫管理器的状态
type CrawlerStatus struct {
	Running        bool                    `json:"running"`        // 是否正在运行
	ContextActive  bool                    `json:"contextActive"`  // 上下文是否活跃
	CrawlerActive  bool                    `json:"crawlerActive"`  // 是否有爬虫实例
	Channels       map[string]*ChannelInfo `json:"channels"`       // 各通道状态
	SupportedMedia []string                `json:"supportedMedia"` // 支持的媒体平台
}

// ChannelInfo 定义通道状态
type ChannelInfo struct {
	Length   int `json:"length"`
	Capacity int `json:"capacity"`
}

// CrawlerCreator 用于创建爬虫实例
type CrawlerCreator func(
	ctx context.Context,
	sessionRegion string,
	sessionManager *session.Manager,
	signClient *signer.SignServerClient,
	eventer *bus.EventBus,
) reference.Crawler

type CrawlerManagerConfig struct {
	SignServEndpoint string
	SchedulerConfig  scheduler.Config
}

// Manager 负责管理爬虫任务
type CrawlerManager struct {
	ctx                context.Context
	mu                 sync.RWMutex
	wg                 *sync.WaitGroup
	running            atomic.Bool
	crawlerInstance    reference.Crawler
	signServer         *signer.SignServerClient
	eventBus           *bus.EventBus
	sessionManager     *session.Manager
	scheduler          *scheduler.Scheduler
	runtimeChannel     chan types.RuntimeData
	crawlers           map[constants.MediaCode]CrawlerCreator
	mapDataChannel     map[string]chan types.FetchItemChan
	currentCrawlParams *types.CrawlParams // 当前轮次参数
}

// NewManager 创建爬虫管理器
func NewCrawlerManager(
	ctx context.Context,
	config CrawlerManagerConfig,
	sessionManager *session.Manager,
	eventBus *bus.EventBus,
	scheduler *scheduler.Scheduler,
	runtimeChannel chan types.RuntimeData,
) *CrawlerManager {
	cm := &CrawlerManager{
		ctx:                ctx,
		running:            atomic.Bool{},
		wg:                 &sync.WaitGroup{},
		sessionManager:     sessionManager,
		eventBus:           eventBus,
		runtimeChannel:     runtimeChannel,
		mapDataChannel:     map[string]chan types.FetchItemChan{},
		crawlers:           make(map[constants.MediaCode]CrawlerCreator),
		signServer:         signer.NewSignServerClient(config.SignServEndpoint),
		scheduler:          scheduler,
		currentCrawlParams: &types.CrawlParams{},
	}

	// 注册抖音爬虫
	cm.Register(constants.MediaCodeDouyin, douyin.NewDouyinCrawler)

	return cm
}

// Register 注册爬虫
func (cm *CrawlerManager) Register(media constants.MediaCode, creator CrawlerCreator) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	cm.crawlers[media] = creator
}

// Create 通过平台名称创建爬虫实例
func (cm *CrawlerManager) Create(media constants.MediaCode, sessionRegion string) reference.Crawler {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	creator, exists := cm.crawlers[media]
	if !exists {
		panic(fmt.Sprintf("Invalid Platform: %s", media))
	}
	return creator(cm.ctx, sessionRegion, cm.sessionManager, cm.signServer, cm.eventBus)
}

// Run 启动爬虫任务
func (cm *CrawlerManager) Run(crawlParams *types.CrawlParams) error {
	if cm.running.Load() {
		logger.Log.Infof("Crawler %s is already running", crawlParams.MediaCode)
		return nil
	}
	// 发送通知
	cm.runtimeChannel <- types.NewRuntimeData(types.RuntimeEventCodeNotification, types.EventData{
		Title:     "数据洞察",
		CheckHash: encrypt.Md5(time.Now().Format("2006-01-02 15:04:05")),
		Message:   "🧠智识引擎上线，开始处理任务...",
		Optional: types.MessageOptional{
			IsNotify: true,
			IsStore:  true,
			ShowType: "notification",
		},
	})
	// 重置调度器状态
	cm.scheduler.Reset()
	// 检查是否已在运行
	if cm.ctx.Err() != nil {
		cm.ctx = cm.scheduler.Context()
	}
	// 修改running为运行中
	cm.running.Store(true)
	// 设置采集QPS
	cm.scheduler.SetQueueQPS(str.GenerateStringKey(crawlParams.MediaCode, "search"), 2)
	cm.scheduler.SetQueueQPS(str.GenerateStringKey(crawlParams.MediaCode, "media"), 6)
	cm.scheduler.SetQueueQPS(str.GenerateStringKey(crawlParams.MediaCode, "user"), 10)
	cm.scheduler.SetQueueQPS(str.GenerateStringKey(crawlParams.MediaCode, "comment"), 6)
	// 初始化channel
	cm.mapDataChannel = map[string]chan types.FetchItemChan{
		"media":   make(chan types.FetchItemChan, 1000),
		"comment": make(chan types.FetchItemChan, 1000),
		"user":    make(chan types.FetchItemChan, 1000),
	}
	// 获取爬虫实例
	cm.crawlerInstance = cm.Create(constants.MediaCode(crawlParams.MediaCode), crawlParams.Region)
	// 初始化爬虫
	cm.crawlerInstance.Initialize(cm.scheduler, cm.runtimeChannel, cm.mapDataChannel)
	// 设置初始采集参数
	cm.currentCrawlParams = crawlParams
	// 启动处理数据线程
	for taskType := range cm.mapDataChannel {
		cm.wg.Add(1)
		go cm.processChannel(cm.mapDataChannel[taskType])
	}
	// 定义是否有子任务参数
	var jobPayloads []interface{}
	// 投递入口任务
	switch crawlParams.CrawlType {
	case string(constants.CrawlerTypeSearch):
		for _, keyword := range crawlParams.Keywords {
			jobPayloads = append(jobPayloads, types.SearchParams{
				Keyword:         keyword,
				WithUser:        crawlParams.WithUser,
				WithComment:     crawlParams.WithComment,
				WithCommentUser: crawlParams.WithCommentUser,
				MaxCount:        crawlParams.MaxCount,
				PageSize:        16,
			})
		}
	case string(constants.CrawlerTypeMedia):
		for _, keyword := range crawlParams.Keywords {
			jobPayloads = append(jobPayloads, types.MediaParams{
				Id:              keyword,
				WithUser:        crawlParams.WithUser,
				WithComment:     crawlParams.WithComment,
				WithCommentUser: crawlParams.WithCommentUser,
			})
		}
	case string(constants.CrawlerTypeUser):
		for _, keyword := range crawlParams.Keywords {
			jobPayloads = append(jobPayloads, types.UserParams{
				UserId:           keyword,
				WithAllCreations: crawlParams.WithAllCreations,
				WithComment:      crawlParams.WithComment,
				WithCommentUser:  crawlParams.WithCommentUser,
			})
		}
	default:
		return errors.New("Unsupport CrawlerType")
	}

	roundSignal := make(chan struct{}, ROUND_MAX*2)
	roundSignal <- struct{}{}

	roundWg := &sync.WaitGroup{}
	roundWg.Add(1)
	go func() {
		defer roundWg.Done()
		// 处理数据轮次
		currentRound := 0
		for {
			select {
			case <-roundSignal:
				// 超出次数，跳出循环
				if currentRound >= ROUND_MAX {
					return
				}
				logger.Log.Infof("Start crawl task round check %d，MediaCode %s", currentRound, crawlParams.MediaCode)
				// 提交任务
				for _, payload := range jobPayloads {
					err := cm.crawlerInstance.SubmitJob(
						crawlParams.CrawlType, payload, scheduler.TaskOptions{},
					)
					if err != nil {
						logger.Log.Error(err.Error())
					}
				}
				// 等待本轮任务完成
				cm.scheduler.WaitUntilEmpty()
				// 增加轮次计数
				currentRound++
				// 如果还有下一轮，暂停一段时间（可配置）
				if currentRound < ROUND_MAX {
					select {
					case <-time.After(time.Duration(ROUND_SLEEP) * time.Second):
					case <-cm.ctx.Done():
						return
					}
				}
				// 添加信号
				roundSignal <- struct{}{}
			case <-cm.ctx.Done():
				return
			}
		}
	}()

	roundWg.Wait()

	if cm.ctx.Err() == nil {
		cm.eventBus.Publish(types.CrawlEndEvent{
			Code:      types.CrawlEndCodeRoundMaxed,
			ReceiveAt: time.Now(),
		})
	}

	cm.cleanup()

	return nil
}

// processChannel 通用通道处理函数
func (cm *CrawlerManager) processChannel(ch chan types.FetchItemChan) {
	defer cm.wg.Done()
	for {
		select {
		case item, ok := <-ch:
			if !ok {
				return
			}
			if item.Data == nil {
				continue
			}
			cm.mu.RLock()
			params := cm.currentCrawlParams
			cm.mu.RUnlock()
			if err := cm.crawlerInstance.HandleChannel(item, params); err != nil {
				logger.Log.Errorf("TaskID=%s: SubmitSubTasks error: %v", item.TaskId, err)
			}
		case <-cm.ctx.Done():
			return
		}
	}
}

// Status 返回爬虫管理器的当前状态
func (cm *CrawlerManager) Status() *CrawlerStatus {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	// 通道状态
	channels := make(map[string]*ChannelInfo)
	for key, ch := range cm.mapDataChannel {
		channels[key] = &ChannelInfo{
			Length:   len(ch),
			Capacity: cap(ch),
		}
	}

	// 支持的媒体平台
	supportedMedia := make([]string, 0, len(cm.crawlers))
	for media := range cm.crawlers {
		supportedMedia = append(supportedMedia, media.String())
	}

	return &CrawlerStatus{
		Running:        cm.running.Load(),
		ContextActive:  cm.ctx.Err() == nil,
		CrawlerActive:  cm.crawlerInstance != nil,
		Channels:       channels,
		SupportedMedia: supportedMedia,
	}
}

// Stop 停止爬虫任务
func (cm *CrawlerManager) Stop() {
	if !cm.running.Load() {
		return
	}
	cm.eventBus.Publish(types.CrawlEndEvent{
		Code:      types.CrawlEndCodeForcedStop,
		ReceiveAt: time.Now(),
	})
}

// cleanup 清理资源
func (cm *CrawlerManager) cleanup() {
	// 关闭所有chan
	for _, channel := range cm.mapDataChannel {
		close(channel)
	}
	cm.mapDataChannel = make(map[string]chan types.FetchItemChan)
	// 等待采集程序process结束
	cm.wg.Wait()
	// 删除当前爬虫实例
	cm.crawlerInstance = nil
	cm.running.Store(false)

	// 清除采集参数
	cm.mu.Lock()
	cm.currentCrawlParams = nil
	cm.mu.Unlock()
}
