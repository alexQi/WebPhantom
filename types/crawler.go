package types

import (
	"strings"
	"time"
)

type FetchItemChan struct {
	TaskId       string
	SourceTaskId string
	Source       string
	Data         interface{}
}

type Keywords []string

func (s *Keywords) String() string {
	return strings.Join(*s, ",")
}

func (s *Keywords) Set(value string) error {
	*s = strings.Split(value, ",") // 按逗号分隔字符串
	return nil
}

type CrawlParams struct {
	MediaCode        string   `json:"mediaCode"`
	CrawlType        string   `json:"crawlType"`
	Region           string   `json:"region"`
	MaxCount         int      `json:"maxCount"`
	Keywords         []string `json:"keywords"`
	WithUser         bool     `json:"withUser"`
	WithComment      bool     `json:"withComment"`
	WithCommentUser  bool     `json:"withCommentUser"`
	WithAllCreations bool     `json:"withAllCreations"` // 是否获取全部作品
	AutoPagination   bool     `json:"autoPagination"`   // 是否自动翻页
	TargetPurgeCount int64    `json:"targetPurgeCount"` // 目标清洗数量
}

type SearchParams struct {
	Keyword          string
	MaxCount         int
	WithUser         bool
	WithComment      bool
	WithCommentUser  bool
	WithAllCreations bool
	RequestId        string
	Page             int
	PageSize         int
	TaskId           string
}

type MediaParams struct {
	Id              string
	WithUser        bool
	WithComment     bool
	WithCommentUser bool
	TaskId          string
	SourceTaskId    string
}

type UserParams struct {
	UserId           string
	WithAllCreations bool
	WithComment      bool
	WithCommentUser  bool
	TaskId           string
	SourceTaskId     string
}

type CommentParams struct {
	Id              string
	Title           string
	WithCommentUser bool
	Cursor          int
	Page            int
	TaskId          string
	SourceTaskId    string
	SourceKeyword   string
}

type MediaData struct {
	TaskId         string
	MediaCode      string
	MediaID        string
	Type           int
	Title          string
	Description    string
	SecUID         string
	LikedCount     int64
	CommentCount   int64
	ShareCount     int64
	CollectedCount int64
	URL            string
	Source         string
	ShortUserID    string
	UserUniqueID   string
	Nickname       string
	Avatar         string
	Gender         string
	Signature      string
	Location       string
	CreateTime     time.Time
}

type CommentData struct {
	TaskId          string
	MediaCode       string
	CommentID       string
	MediaID         string
	SecUID          string
	Content         string
	CreateTime      time.Time
	SubCommentCount int
	ParentCommentID string
	LikeCount       int
	Pictures        string
	RawData         string
	Source          string
	ShortUserID     string
	UserUniqueID    string
	Nickname        string
	Avatar          string
	Gender          string
	Signature       string
	Location        string
}
