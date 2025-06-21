package douyin

import (
	"context"
	"fmt"
	"noctua/internal/media/douyin"
	"noctua/internal/signer"
	"noctua/pkg/logger"
	"noctua/types"
)

// DouyinFetcherCrawler 具体的抖音爬虫
type DouyinFetcher struct {
	ctx        context.Context
	dataClient *douyin.DouYinApiClient
}

func NewDouyinFetcher(ctx context.Context, signClient *signer.SignServerClient) *DouyinFetcher {
	return &DouyinFetcher{
		ctx:        ctx,
		dataClient: douyin.NewDouYinApiClient(signClient),
	}
}

func (d *DouyinFetcher) Initialize() {
	//d.dataClient.BuildVerifyParams()
}

// HandleSearch 使用泛型处理不同类型的通道
func (d *DouyinFetcher) HandleSearch(params *types.SearchParams, mediaChan chan types.FetchItemChan) (bool, bool, int, error) {
	searchParams := &douyin.SearchParams{
		Keyword:         params.Keyword,
		SearchChannel:   douyin.SearchChannelVideo,
		PublishTimeType: douyin.PublishTimeUnlimited,
		SortType:        douyin.SearchSortLatest,
		SearchId:        params.RequestId,
		Count:           params.PageSize,
		Offset:          params.Page * params.PageSize,
	}
	logger.Log.Infof("Douyin.search: Keyword=%s, Page=%d", params.Keyword, params.Page+1)
	// 判断总数设置的查询总计记录数
	searchResult, err := d.dataClient.SearchInfoByKeyword(searchParams)
	if err != nil {
		logger.Log.Errorf("Douyin.search: Keyword=%s, Page=%d, err：%s", params.Keyword, params.Page, err.Error())
		return false, false, 0, err
	}
	if searchResult.SearchNilInfo.SearchNilItem != "" {
		logger.Log.Infof(
			"Douyin.search nil, nil item: %s, type: %s, loadMore: %s , textType: %d",
			searchResult.SearchNilInfo.SearchNilItem,
			searchResult.SearchNilInfo.SearchNilType,
			searchResult.SearchNilInfo.IsLoadMore,
			searchResult.SearchNilInfo.TextType,
		)
	}
	// 判断是否需要启动验证实例
	if searchResult.SearchNilInfo.SearchNilType == "verify_check" {
		return true, false, 0, nil
	}
	if len(searchResult.Data) == 0 {
		logger.Log.Infof("Douyin.search: Keyword=%s, Page=%d, Page End", params.Keyword, params.Page)
		return false, false, 0, nil
	}
	params.RequestId = searchResult.Extra.Logid
	for _, searchItem := range searchResult.Data {
		// 这里进行类型断言，将特定类型的数据传递给 channel
		if v, ok := any(searchItem.AwemeInfo).(douyin.Aweme); ok {
			item := types.FetchItemChan{
				TaskId:       params.TaskId,
				SourceTaskId: params.TaskId,
				Source:       params.Keyword,
				Data:         v,
			}
			if d.ctx.Err() != nil {
				return false, false, len(searchResult.Data), d.ctx.Err()
			}
			select {
			case <-d.ctx.Done():
				logger.Log.Infof("Douyin.fetcher-search: stopped due to context cancellation")
				return false, false, len(searchResult.Data), d.ctx.Err()
			case mediaChan <- item:
			default:
				logger.Log.Warnf("Douyin.fetcher-search: mediaChan closed or full for Keyword=%s, Page=%d", params.Keyword, params.Page)
				return false, false, len(searchResult.Data), fmt.Errorf("media channel closed or full")
			}
		}
	}
	return false, true, len(searchResult.Data), nil
}

// HandleComments 使用泛型处理评论通道
func (d *DouyinFetcher) HandleComments(params *types.CommentParams, commentChan chan types.FetchItemChan) (bool, int, error) {
	commentResult, err := d.dataClient.GetAwemeComments(params.Id, params.Cursor, params.SourceKeyword)
	if err != nil {
		logger.Log.Errorf("Douyin.fetcher-comment，get media comment %s failed, err：%s", params.Id, err.Error())
		return false, 0, err
	}
	if len(commentResult.Comments) == 0 {
		return false, 0, nil
	}
	params.Cursor = commentResult.Cursor
	for _, comment := range commentResult.Comments {
		if v, ok := any(comment).(douyin.Comment); ok {
			item := types.FetchItemChan{
				TaskId:       params.TaskId,
				Source:       params.Title,
				SourceTaskId: params.SourceTaskId,
				Data:         v,
			}
			if d.ctx.Err() != nil {
				return false, 0, d.ctx.Err()
			}
			select {
			case <-d.ctx.Done():
				logger.Log.Infof("Douyin.fetcher-comment: stopped due to context cancellation")
				return false, 0, d.ctx.Err()
			case commentChan <- item:
			default:
				logger.Log.Warnf("Douyin.fetcher-comment: commentChan closed or full for ID=%s", params.Id)
				return false, 0, fmt.Errorf("comment channel closed or full")
			}
		}
	}
	return commentResult.HasMore > 0, len(commentResult.Comments), nil
}

// HandleMedia 使用泛型处理视频通道
func (d *DouyinFetcher) HandleMedia(params *types.MediaParams, mediaChan chan types.FetchItemChan) error {
	logger.Log.Infof("Douyin.fetcher-media, search media: %s", params.Id)
	mediaResult, err := d.dataClient.GetVideoByID(params.Id)
	if err != nil {
		return fmt.Errorf("Douyin.fetcher-media, err：%s", err.Error())
	}
	if v, ok := any(mediaResult).(douyin.Aweme); ok {
		item := types.FetchItemChan{
			TaskId:       params.TaskId,
			SourceTaskId: params.SourceTaskId,
			Source:       "media:" + params.Id,
			Data:         v,
		}
		select {
		case <-d.ctx.Done():
			logger.Log.Infof("Douyin.fetcher-media: stopped due to context cancellation")
			return d.ctx.Err()
		case mediaChan <- item:
		default:
			logger.Log.Warnf("Douyin.fetcher-media: mediaChan closed or full for ID=%s", params.Id)
			return fmt.Errorf("media channel closed or full")
		}
	}
	return nil
}

// HandleUser 使用泛型处理用户通道
func (d *DouyinFetcher) HandleUser(params *types.UserParams, userChan chan types.FetchItemChan) error {
	// todo 临时测试
	logger.Log.Infof("Douyin.fetcher-user, search user: %s", params.UserId)
	userResult, err := d.dataClient.GetUserInfo(params.UserId)
	if err != nil {
		return fmt.Errorf("Douyin.fetcher-user, err：%s", err.Error())
	}
	if v, ok := any(userResult).(douyin.User); ok {
		item := types.FetchItemChan{
			TaskId:       params.TaskId,
			SourceTaskId: params.SourceTaskId,
			Data:         v,
		}
		if d.ctx.Err() != nil {
			return d.ctx.Err()
		}
		select {
		case <-d.ctx.Done():
			logger.Log.Infof("Douyin.fetcher-user: stopped due to context cancellation")
			return d.ctx.Err()
		case userChan <- item:
		default:
			logger.Log.Warnf("Douyin.fetcher-user: userChan closed or full for UserID=%s", params.UserId)
			return fmt.Errorf("user channel closed or full")
		}
	}
	return nil
}
