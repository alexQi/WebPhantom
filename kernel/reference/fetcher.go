package reference

import "noctua/types"

// Fetcher 爬虫接口，所有爬虫都必须实现
type Fetcher interface {
	Initialize()

	HandleSearch(params *types.SearchParams, mediaChan chan types.FetchItemChan) error

	HandleMedia(params *types.MediaParams, mediaChan chan types.FetchItemChan) error

	HandleComments(params *types.CommentParams, commentChan types.FetchItemChan) error

	HandleUser(params *types.UserParams, userChan chan types.FetchItemChan) error
}
