package model

import (
	"gorm.io/gorm"
	"noctua/pkg/database"
	"time"
)

// CrawlTask 采集任务表
type CrawlTask struct {
	ID           uint           `json:"id" gorm:"primaryKey"`
	TaskId       string         `json:"task_id" gorm:"uniqueIndex;size:64"`
	ParentTaskId string         `json:"parent_task_id" gorm:"size:64;index"`
	SourceTaskId string         `json:"source_task_id" gorm:"size:64;index"`
	MediaCode    string         `json:"media_code" gorm:"size:32;index"`
	Type         string         `json:"type" gorm:"size:32;index"`
	Payload      string         `json:"payload" gorm:"type:text"`
	Status       string         `json:"status" gorm:"size:32;index"`
	ErrorMessage string         `json:"error_message" gorm:"type:text"`
	RetryCount   int            `json:"retry_count" gorm:"default:0"`
	CreatedAt    time.Time      `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt    time.Time      `json:"updated_at" gorm:"autoUpdateTime"`
	CompletedAt  time.Time      `json:"completed_at" gorm:"index"`
	DeletedAt    gorm.DeletedAt `json:"deleted_at" gorm:"index"`
}

// TableName 指定表名
func (m *CrawlTask) TableName() string {
	return "crawl_task"
}

// CrawlTaskQueryParams 查询参数
type CrawlTaskQueryParams struct {
	MediaCode    string // 平台代码
	Type         string // 任务类型
	Status       string // 任务状态
	ParentTaskId string // 父任务 ID
	TaskId       string // 任务 ID
	RangeTime    []time.Time
}

func (c *CrawlTask) UpsertModel() error {
	// 查找是否已经存在记录
	var existingCrawlTask CrawlTask
	err := database.DB.Where("task_id = ?", c.TaskId).First(&existingCrawlTask).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return err
	}

	// 如果记录已存在，执行更新操作
	if err == nil {
		return database.DB.Model(&existingCrawlTask).
			Where("task_id = ?", c.TaskId).
			Updates(c).Error
	} else {
		// 如果记录不存在，执行插入操作
		return database.DB.Create(c).Error
	}
}

// List 分页查询任务列表
func (c *CrawlTask) List(params *CrawlTaskQueryParams, sort, order string, page, pageSize int) (database.PageResult[CrawlTask], error) {
	query := database.DB.Model(&CrawlTask{})
	if params.MediaCode != "" {
		query = query.Where("media_code = ?", params.MediaCode)
	}
	if params.Type != "" {
		query = query.Where("type = ?", params.Type)
	}
	if params.Status != "" {
		query = query.Where("status = ?", params.Status)
	}
	if params.ParentTaskId != "" {
		query = query.Where("parent_task_id = ?", params.ParentTaskId)
	}
	if params.TaskId != "" {
		query = query.Where("task_id = ?", params.TaskId)
	}
	if len(params.RangeTime) > 0 {
		query = query.Where("created_at BETWEEN ? AND ?", params.RangeTime[0], params.RangeTime[1])
	}
	return database.Paginate[CrawlTask](query, database.ListOptions{
		Page:     page,
		PageSize: pageSize,
		Sort:     sort,
		Order:    order,
	})
}
