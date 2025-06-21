package model

import (
	"gorm.io/gorm"
	"noctua/pkg/database"
	"time"
)

// CrawlComment 视频评论表
type CrawlComment struct {
	ID              uint           `json:"id" gorm:"primaryKey"`
	TaskId          string         `json:"task_id" gorm:"index;size:64"`
	SourceTaskId    string         `json:"source_task_id" gorm:"index;size:64"`
	MediaCode       string         `json:"media_code" gorm:"size:32;index"`
	CommentID       string         `json:"comment_id" gorm:"uniqueIndex;size:64"`
	MediaID         string         `json:"media_id" gorm:"index;size:64"`
	SecUID          string         `json:"sec_uid" gorm:"size:64;index"`
	Nickname        string         `json:"nickname" gorm:"size:255"`
	Location        string         `json:"location" gorm:"size:64"`
	Content         string         `json:"content" gorm:"type:text"`
	CreateTime      time.Time      `json:"create_time" gorm:"index"`
	SubCommentCount int            `json:"sub_comment_count" gorm:"index"`
	ParentCommentID string         `json:"parent_comment_id" gorm:"size:64;index"`
	LikeCount       int            `json:"like_count" gorm:"index"`
	Pictures        string         `json:"pictures" gorm:"type:text"`
	RawData         string         `json:"raw_data" gorm:"type:text"`
	Source          string         `json:"source" gorm:"size:255"`
	IsPurge         bool           `json:"is_purge" gorm:"default:false"`
	CreatedAt       time.Time      `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt       time.Time      `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt       gorm.DeletedAt `json:"deleted_at" gorm:"index"`
}

// TableName 指定表名
func (m *CrawlComment) TableName() string {
	return "crawl_comment"
}

// CrawlCommentQueryParams 查询参数
type CrawlCommentQueryParams struct {
	Keyword   string
	TaskId    string // 任务 ID
	CommentID string // 评论 ID
	MediaID   string // 视频 ID
	SecUID    string // 用户 SecUID
	IsPurge   int
	RangeTime []time.Time
}

// UpsertModel 更新或插入评论记录
func (c *CrawlComment) UpsertModel() error {
	// 查找是否已经存在记录
	var existingCrawlComment CrawlComment
	err := database.DB.Where("comment_id = ?", c.CommentID).First(&existingCrawlComment).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return err
	}

	// 如果记录已存在，执行更新操作
	if err == nil {
		return database.DB.Model(&existingCrawlComment).
			Where("comment_id = ?", c.CommentID).
			Updates(c).Error
	} else {
		// 如果记录不存在，执行插入操作
		return database.DB.Create(c).Error
	}
}

// List 分页查询评论列表
func (c *CrawlComment) List(params *CrawlCommentQueryParams, sort, order string, page, pageSize int) (database.PageResult[CrawlComment], error) {
	query := database.DB.Model(&CrawlComment{})
	if params.TaskId != "" {
		query = query.Where("task_id = ?", params.TaskId)
	}
	if params.CommentID != "" {
		query = query.Where("comment_id = ?", params.CommentID)
	}
	if params.MediaID != "" {
		query = query.Where("media_id = ?", params.MediaID)
	}
	if params.SecUID != "" {
		query = query.Where("sec_uid = ?", params.SecUID)
	}
	if params.Keyword != "" {
		query = query.Where("content like ?", "%"+params.Keyword+"%")
	}
	if params.IsPurge != 0 {
		query = query.Where("status = ?", params.IsPurge)
	}
	if len(params.RangeTime) > 0 {
		query = query.Where("created_at BETWEEN ? AND ?", params.RangeTime[0], params.RangeTime[1])
	}
	return database.Paginate[CrawlComment](query, database.ListOptions{
		Page:     page,
		PageSize: pageSize,
		Sort:     sort,
		Order:    order,
	})
}
