package douyin

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"noctua/internal/constants"
	"noctua/internal/media/douyin"
	"noctua/internal/model"
	"noctua/internal/scheduler"
	"noctua/internal/signer"
	"noctua/kernel/bus"
	"noctua/kernel/reference"
	"noctua/kernel/session"
	"noctua/pkg/logger"
	"noctua/pkg/utils/str"
	"noctua/types"
	"time"
)

// DouyinCrawlerCrawler 具体的抖音爬虫
type DouyinCrawler struct {
	ctx            context.Context
	mediaCode      constants.MediaCode
	scheduler      *scheduler.Scheduler
	eventBus       *bus.EventBus
	dataFetcher    *DouyinFetcher
	dataSaver      *DouyinDataSaver
	channels       map[string]chan types.FetchItemChan
	runtimeChannel chan types.RuntimeData
}

// NewDouyinCrawler 创建 DouyinCrawler 实例
func NewDouyinCrawler(
	ctx context.Context,
	sessionRegion string,
	sessionManager *session.Manager,
	signClient *signer.SignServerClient,
	eventBus *bus.EventBus,
) reference.Crawler {
	dc := &DouyinCrawler{
		ctx:       ctx,
		eventBus:  eventBus,
		mediaCode: constants.MediaCodeDouyin,
		dataSaver: &DouyinDataSaver{},
	}
	// 创建data fetcher
	dataFetcher := NewDouyinFetcher(dc.ctx, signClient)
	// 设置获取session callback func
	dataFetcher.dataClient.OnAcquireSession(func(nowSession *types.Session) (*types.Session, error) {
		var currSession *types.Session
		params := &session.SessionParams{MediaCode: "douyin", SessionRegion: sessionRegion, AccountType: 1}
		if nowSession != nil && nowSession.Enabled {
			params.UserID = nowSession.Account.UserID
		}
		for {
			// 增加判断当前可用账号统计的判断
			newSession, err := sessionManager.GetSession(params)
			if err != nil {
				return nil, err
			}
			currSession = newSession
			break
		}
		return currSession, nil
	})
	// 设置刷新session callback func
	dataFetcher.dataClient.OnRefreshSession(func(nowSession *types.Session) (*types.Session, error) {
		// 参数校验
		if nowSession == nil || nowSession.Account == nil || nowSession.Account.UserID == "" {
			return nil, fmt.Errorf("invalid session: session and account details are required")
		}
		// 增加判断当前可用账号统计的判断
		newSession, err := sessionManager.ReplaceSession(
			nowSession.Account.MediaCode,
			nowSession.Account.UserID,
			sessionRegion,
		)
		if err != nil {
			return nil, err
		}
		return newSession, nil
	})
	// 设置丢弃session callback func
	dataFetcher.dataClient.OnDiscardSession(func(currSession *types.Session) {
		err := sessionManager.InvalidateSession(currSession)
		if err != nil {
			logger.Log.Errorf("sessionManager.InvalidateSession err: %v", err)
			return
		}
	})
	// 处理session无法找到有效账号
	dataFetcher.dataClient.OnMissingSession(func() {
		dc.eventBus.Publish(types.CrawlEndEvent{
			Code:      10,
			ReceiveAt: time.Now(),
		})
	})
	dc.dataFetcher = dataFetcher

	return dc
}

func (d *DouyinCrawler) Initialize(scheduler *scheduler.Scheduler, runtimeChannel chan types.RuntimeData, channels map[string]chan types.FetchItemChan) {
	d.scheduler = scheduler
	d.channels = channels
	d.runtimeChannel = runtimeChannel
	// 初始化dataFetcher
	d.dataFetcher.Initialize()
	// 初始化handler
	d.scheduler.RegisterHandler(str.GenerateStringKey(d.mediaCode.String(), "search"), d.handleSearch)
	d.scheduler.RegisterHandler(str.GenerateStringKey(d.mediaCode.String(), "media"), d.handleMedia)
	d.scheduler.RegisterHandler(str.GenerateStringKey(d.mediaCode.String(), "user"), d.handleUser)
	d.scheduler.RegisterHandler(str.GenerateStringKey(d.mediaCode.String(), "comment"), d.handleComment)
}

// SubmitSubTasks 提交子任务（集中在 crawler 中）
func (d *DouyinCrawler) HandleChannel(item types.FetchItemChan, params *types.CrawlParams) error {
	switch data := item.Data.(type) {
	case douyin.Aweme:
		if params.WithUser {
			err := d.SubmitJob("user", types.UserParams{
				UserId:           data.Author.SecUID,
				WithAllCreations: params.WithAllCreations,
				WithComment:      params.WithComment,
				WithCommentUser:  params.WithCommentUser,
			}, scheduler.TaskOptions{
				ParentTaskID: item.TaskId,
			})
			if err != nil {
				return err
			}
		}
		if params.WithComment {
			// 视频数据过少不处理
			if data.Statistics.CommentCount < 5 {
				return nil
			}

			err := d.SubmitJob("comment", types.CommentParams{
				Id:              data.AwemeID,
				Title:           data.Desc,
				WithCommentUser: params.WithCommentUser,
				SourceKeyword:   item.Source,
			}, scheduler.TaskOptions{
				ParentTaskID: item.TaskId,
				SourceTaskID: item.SourceTaskId,
			})
			if err != nil {
				return err
			}
		}
		go func() {
			err := d.dataSaver.HandleMedia(data, item.TaskId, item.SourceTaskId, item.Source)
			if err != nil {
				logger.Log.Errorf("TaskID=%s: SaveMedia error: %v", item.TaskId, err)
			}
		}()
		return nil
	case douyin.Comment:
		// 提交采集用户信息任务
		if params.WithCommentUser {
			err := d.SubmitJob("user", types.UserParams{
				UserId: data.User.SecUID,
			}, scheduler.TaskOptions{
				ParentTaskID: item.TaskId,
				SourceTaskID: item.SourceTaskId,
			})
			if err != nil {
				return err
			}
		}
		// 评论为空 跳过
		if len(data.Text) == 0 {
			return nil
		}
		d.runtimeChannel <- types.NewRuntimeData(
			types.RuntimeEventCodeCrawl,
			types.EventData{
				MetaData: map[string]string{
					"user":      data.User.Nickname,
					"content":   data.Text,
					"createdAt": time.Unix(data.CreateTime, 0).Format("2006-01-02 15:04:05"),
				},
			},
		)
		// 评论内容
		go func() {
			err := d.dataSaver.HandleComment(data, item.TaskId, item.SourceTaskId, item.Source)
			if err != nil {
				logger.Log.Errorf("TaskID=%s: SaveComment error: %v", item.TaskId, err)
			}
		}()
		return nil
	case douyin.User:
		if params.WithAllCreations {
			// TODO: 提交用户作品采集任务
		}
		go func() {
			err := d.dataSaver.HandleUser(data, item.TaskId, item.SourceTaskId, item.Source)
			if err != nil {
				logger.Log.Errorf("TaskID=%s: SaveUser error: %v", item.TaskId, err)
			}
		}()
		return nil
	default:
		return fmt.Errorf("unsupported data type: %T", item.Data)
	}
}

// SubmitTask 提交爬虫任务
func (d *DouyinCrawler) SubmitJob(taskType string, payload interface{}, options scheduler.TaskOptions) error {
	// 创建任务
	task, err := scheduler.NewTask(str.GenerateStringKey(d.mediaCode.String(), taskType), payload, options)
	if err != nil {
		return fmt.Errorf("Build task failed: %v", err)
	}
	// 提交任务到调度器
	taskId, err := d.scheduler.SubmitTask(task)
	if err != nil {
		return fmt.Errorf("Submit task failed: %v", err)
	}
	// 日志记录任务提交成功
	logger.Log.Infof("Task submitted: Queue=%s, Parent=%t, ID=%s", task.QueueKey, len(task.ParentTaskID) > 0, taskId)
	// 保存任务数据
	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("Marshal payload failed: %v", err)
	}
	// 创建任务记录 (Pending)
	taskRecord := &model.CrawlTask{
		TaskId:       taskId, // 临时 ID
		MediaCode:    string(d.mediaCode),
		Type:         taskType,
		Payload:      string(payloadJSON),
		Status:       "Running",
		ParentTaskId: options.ParentTaskID,
		SourceTaskId: task.SourceTaskID,
	}
	if err := taskRecord.UpsertModel(); err != nil {
		return fmt.Errorf("Save crawl task failed: %v", err)
	}
	return nil
}

// handler 函数抽取为独立方法，提升可读性
func (d *DouyinCrawler) handleSearch(t *scheduler.Task) error {
	params, ok := t.Payload.(types.SearchParams)
	if !ok {
		return fmt.Errorf("data type error, expected: types.SearchParams, got: %T", t.Payload)
	}
	params.TaskId = t.ID
	verify, hasMore, count, err := d.dataFetcher.HandleSearch(&params, d.channels["media"])
	if err != nil {
		return err
	}
	// 获取最大可用页数，判断是否要提交分页
	maxPage := int(math.Ceil(float64(params.MaxCount) / float64(params.PageSize)))
	// 提交翻页请求
	if hasMore && (params.Page+1) < maxPage {
		// 页码递增
		params.Page += 1
		// 提交Job到调度器
		err := d.SubmitJob("search", params, scheduler.TaskOptions{
			SourceTaskID: params.TaskId,
			ParentTaskID: params.TaskId,
		})
		if err != nil {
			logger.Log.Errorf("TaskID=%s: Search error: %v", params.TaskId, err)
		}
	}
	// 子任务在 HandleChannel 中提交
	if verify && count == 0 {
		return fmt.Errorf("Douyin account need handle captcha")
	}
	return nil
}

func (d *DouyinCrawler) handleComment(t *scheduler.Task) error {
	params, ok := t.Payload.(types.CommentParams)
	if !ok {
		return fmt.Errorf("data type error, expected: types.CommentParams, got: %T", t.Payload)
	}
	params.TaskId = t.ID
	params.SourceTaskId = t.SourceTaskID
	hasMore, count, err := d.dataFetcher.HandleComments(&params, d.channels["comment"])
	if err != nil {
		return err
	}
	if params.Cursor == 0 && count == 0 {
		d.eventBus.Publish(types.CrawlEndEvent{
			Code:      types.CrawlEndCodeOverdLimit,
			ReceiveAt: time.Now(),
		})
		return fmt.Errorf("Account execeed limit...")
	}

	// 提交翻页请求
	if hasMore {
		// 提交Job到调度器
		err := d.SubmitJob("comment", params, scheduler.TaskOptions{
			ParentTaskID: t.ParentTaskID,
			SourceTaskID: params.SourceTaskId,
		})
		if err != nil {
			logger.Log.Errorf("TaskID=%s: Comment error: %v", params.TaskId, err)
		}
	}
	return nil
}

func (d *DouyinCrawler) handleMedia(t *scheduler.Task) error {
	params, ok := t.Payload.(types.MediaParams)
	if !ok {
		return fmt.Errorf("data type error, expected: types.MediaParams, got: %T", t.Payload)
	}
	params.TaskId = t.ID
	params.SourceTaskId = t.SourceTaskID
	err := d.dataFetcher.HandleMedia(&params, d.channels["media"])
	if err != nil {
		return err
	}
	return nil
}

func (d *DouyinCrawler) handleUser(t *scheduler.Task) error {
	params, ok := t.Payload.(types.UserParams)
	if !ok {
		return fmt.Errorf("data type error, expected: types.UserParams, got: %T", t.Payload)
	}
	params.TaskId = t.ID
	params.SourceTaskId = t.SourceTaskID
	err := d.dataFetcher.HandleUser(&params, d.channels["user"])
	if err != nil {
		return err
	}
	return nil
}
