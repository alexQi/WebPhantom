package model

import (
	"gorm.io/gorm"
	"noctua/pkg/database"
	"time"
)

// CrawlUser 媒体用户表
type CrawlUser struct {
	ID           uint           `json:"id" gorm:"primaryKey"`
	TaskId       string         `json:"task_id" gorm:"index;size:64"`
	SourceTaskId string         `json:"source_task_id" gorm:"index;size:64"`
	MediaCode    string         `json:"media_code" gorm:"size:32;index"`
	SecUID       string         `json:"sec_uid" gorm:"uniqueIndex;size:64"`
	ShortUserID  string         `json:"short_user_id" gorm:"size:64"`
	UserUniqueID string         `json:"user_unique_id" gorm:"size:64"`
	Nickname     string         `json:"nickname" gorm:"size:255"`
	Avatar       string         `json:"avatar" gorm:"size:255"`
	Gender       string         `json:"gender" gorm:"type:varchar(2)"` // 性别
	Signature    string         `json:"signature" gorm:"type:text"`
	Location     string         `json:"location" gorm:"size:64"`
	Follows      string         `json:"follows" gorm:"type:varchar(16)"`      // 关注数
	Fans         string         `json:"fans" gorm:"type:varchar(16)"`         // 粉丝数
	Interaction  string         `json:"interaction" gorm:"type:varchar(16)"`  // 获赞数
	VideosCount  string         `json:"videos_count" gorm:"type:varchar(16)"` // 作品数
	IsGreet      int            `json:"is_greet" gorm:"type:tinyint(1);default 0;not null"`
	GreetTime    time.Time      `json:"greet_time" gorm:"type:datetime"` // 视频发布时间戳
	Source       string         `json:"source" gorm:"size:255"`
	RawData      string         `json:"raw_data" gorm:"type:text"`
	CreatedAt    time.Time      `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt    time.Time      `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt    gorm.DeletedAt `json:"deleted_at" gorm:"index"`
}

// TableName 指定表名
func (m *CrawlUser) TableName() string {
	return "crawl_user"
}

// CrawlUserQueryParams 查询参数
type CrawlUserQueryParams struct {
	TaskId    string // 任务 ID
	SecUID    string // 用户 SecUID
	Nickname  string // 用户昵称
	IsGreet   int
	RangeTime []time.Time
}

func (c *CrawlUser) UpsertModel() error {
	// 查找是否已经存在记录
	var existingCrawlUser CrawlUser
	err := database.DB.Where("sec_uid = ?", c.SecUID).First(&existingCrawlUser).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return err
	}

	// 如果记录已存在，执行更新操作
	if err == nil {
		return database.DB.Model(&existingCrawlUser).
			Where("sec_uid = ?", c.SecUID).
			Updates(c).Error
	} else {
		// 如果记录不存在，执行插入操作
		return database.DB.Create(c).Error
	}
}

// List 分页查询用户列表
func (m *CrawlUser) List(params *CrawlUserQueryParams, sort, order string, page, pageSize int) (database.PageResult[CrawlUser], error) {

	query := database.DB.Model(&CrawlUser{})
	if params.TaskId != "" {
		query = query.Where("task_id = ?", params.TaskId)
	}
	if params.SecUID != "" {
		query = query.Where("sec_uid = ?", params.SecUID)
	}
	if params.Nickname != "" {
		query = query.Where("nickname LIKE ?", "%"+params.Nickname+"%")
	}
	if params.IsGreet >= 0 {
		query = query.Where("is_greet = ?", params.IsGreet)
	}
	if len(params.RangeTime) > 0 {
		query = query.Where("created_at BETWEEN ? AND ?", params.RangeTime[0], params.RangeTime[1])
	}
	return database.Paginate[CrawlUser](query, database.ListOptions{
		Page:     page,
		PageSize: pageSize,
		Sort:     sort,
		Order:    order,
	})
}
