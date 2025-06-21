package model

import (
	"gorm.io/gorm"
	"noctua/pkg/database"
	"time"
)

// CrawlMedia 采集到的视频表
type CrawlMedia struct {
	ID             uint           `json:"id" gorm:"primaryKey"`
	TaskId         string         `json:"task_id" gorm:"index;size:64"`
	SourceTaskId   string         `json:"source_task_id" gorm:"index;size:64"`
	MediaCode      string         `json:"media_code" gorm:"size:32;index"`
	MediaID        string         `json:"media_id" gorm:"uniqueIndex;size:64"`
	Type           int            `json:"type" gorm:"index"`
	Title          string         `json:"title" gorm:"size:1024"`
	Description    string         `json:"description" gorm:"type:text"`
	SecUID         string         `json:"sec_uid" gorm:"size:64;index"`
	Nickname       string         `json:"nickname" gorm:"size:255"`
	CreateTime     time.Time      `json:"create_time" gorm:"index"`
	LikedCount     int64          `json:"liked_count" gorm:"index"`
	CommentCount   int64          `json:"comment_count" gorm:"index"`
	ShareCount     int64          `json:"share_count" gorm:"index"`
	CollectedCount int64          `json:"collected_count" gorm:"index"`
	URL            string         `json:"url" gorm:"size:255"`
	RawData        string         `json:"raw_data" gorm:"type:text"`
	Source         string         `json:"source" gorm:"size:255;index"`
	CreatedAt      time.Time      `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt      time.Time      `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt      gorm.DeletedAt `json:"deleted_at" gorm:"index"`
}

// TableName 指定表名
func (m *CrawlMedia) TableName() string {
	return "crawl_media"
}

// CrawlMediaQueryParams 查询参数
type CrawlMediaQueryParams struct {
	Keyword   string
	TaskId    string      `json:"task_id"`  // 任务 ID
	MediaID   string      `json:"media_id"` // 视频 ID
	SecUID    string      `json:"sec_uid"`  // 用户 SecUID
	Source    string      `json:"source"`   // 来源关键字
	RangeTime []time.Time `json:"range_time"`
}

// UpsertModel 更新或插入视频记录
func (m *CrawlMedia) UpsertModel() error {
	// 查找是否已经存在记录
	var existingCrawlMedia CrawlMedia
	err := database.DB.Where("media_id = ?", m.MediaID).First(&existingCrawlMedia).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return err
	}

	// 如果记录已存在，执行更新操作
	if err == nil {
		// 使用 Omit 排除不需要更新的字段（如 ID 和更新时间字段）
		return database.DB.Model(&existingCrawlMedia).
			Where("media_id = ?", m.MediaID).
			Updates(m).Error
	} else {
		// 如果记录不存在，执行插入操作
		return database.DB.Create(m).Error
	}
}

// List 分页查询视频列表
func (m *CrawlMedia) List(params *CrawlMediaQueryParams, sort, order string, page, pageSize int) (database.PageResult[CrawlMedia], error) {
	query := database.DB.Model(&CrawlMedia{})
	if params.TaskId != "" {
		query = query.Where("task_id = ?", params.TaskId)
	}
	if params.MediaID != "" {
		query = query.Where("media_id = ?", params.MediaID)
	}
	if params.SecUID != "" {
		query = query.Where("sec_uid = ?", params.SecUID)
	}
	if params.Source != "" {
		query = query.Where("source LIKE ?", "%"+params.Source+"%")
	}
	if params.Keyword != "" {
		query = query.Where("title like ?", "%"+params.Keyword+"%")
	}
	if len(params.RangeTime) > 0 {
		query = query.Where("created_at BETWEEN ? AND ?", params.RangeTime[0], params.RangeTime[1])
	}
	return database.Paginate[CrawlMedia](query, database.ListOptions{
		Page:     page,
		PageSize: pageSize,
		Sort:     sort,
		Order:    order,
	})
}
