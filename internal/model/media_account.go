package model

import (
	"gorm.io/gorm"
	"noctua/pkg/database"
	"time"
)

// MediaAccount 结构体表示 media_account 表
type MediaAccount struct {
	ID         uint           `json:"id" gorm:"primaryKey"`                              // 主键
	MediaCode  string         `json:"media_code" gorm:"not null"`                        // 平台名称
	Type       int            `json:"type" gorm:"default:0"`                             // 0: 通用，1: 爬虫用 2 私信
	UserID     string         `json:"user_id" gorm:"not null"`                           // 用户id，我方生成
	UID        string         `json:"uid" gorm:"not null"`                               // uid 媒体用户id
	Username   string         `json:"username" gorm:"default:''"`                        // 账号名称
	Nickname   string         `json:"nickname" gorm:"not null"`                          // 昵称
	Cookie     string         `json:"cookie" gorm:"not null"`                            // 账号cookie
	UserAgent  string         `json:"user_agent" gorm:"default:''"`                      // userAgent
	DeviceInfo string         `json:"device_info" gorm:"default:''"`                     // 设置信息
	Status     int            `json:"status" gorm:"default:10"`                          // 状态：10 正常,100 禁用
	IsReal     int            `json:"is_real" gorm:"type:tinyint(1);default:1;not null"` // 是否为真实的，数据库中都为真实
	LastUsed   time.Time      `json:"last_used" gorm:"default:CURRENT_TIMESTAMP"`        // 最后使用时间
	CreateTime time.Time      `json:"create_time" gorm:"autoCreateTime"`                 // 自动创建时间
	UpdateTime time.Time      `json:"update_time" gorm:"autoUpdateTime"`                 // 自动更新时间
	DeletedAt  gorm.DeletedAt `json:"deleted_at" gorm:"index"`
}

type QueryMediaAccountParams struct {
	MediaCode      string
	Type           int
	ID             uint
	UserID         string
	Username       string
	Nickname       string
	Status         int
	ExcludeUserIDs []string
	RangeTime      []time.Time
}

// TableName 指定表名
func (m *MediaAccount) TableName() string {
	return "media_account"
}

func (m *MediaAccount) FindMediaAccount(params *QueryMediaAccountParams) (*MediaAccount, error) {
	mediaAccount := &MediaAccount{}
	query := database.DB.Where("status = 10").Where("deleted_at is null").Order("id desc")
	if params.MediaCode != "" {
		query.Where("media_code = ?", params.MediaCode)
	}
	if params.Type > 0 {
		query.Where("type = ?", params.Type)
	}
	if len(params.UserID) > 0 {
		query.Where("user_id = ?", params.UserID)
	}
	if len(params.ExcludeUserIDs) > 0 {
		query.Where("user_id not in (?)", params.ExcludeUserIDs)
	}
	result := query.First(&mediaAccount)
	if result.Error != nil && result.Error != gorm.ErrRecordNotFound {
		return mediaAccount, result.Error
	}
	return mediaAccount, nil
}

func (m *MediaAccount) QueryMediaAccounts(params *QueryMediaAccountParams) ([]*MediaAccount, error) {
	MediaAccounts := make([]*MediaAccount, 0)
	query := database.DB.Order("id desc").Where("deleted_at is null")
	if params.MediaCode != "" {
		query.Where("media_code = ?", params.MediaCode)
	}
	if params.Type > 0 {
		query.Where("type = ?", params.Type)
	}
	if params.Status > 0 {
		query.Where("status = ?", params.Status)
	}
	if params.ID > 0 {
		query.Where("id = ?", params.ID)
	}
	if len(params.UserID) > 0 {
		query.Where("user_id = ?", params.UserID)
	}
	if len(params.Username) > 0 {
		query.Where("username = ?", params.Username)
	}
	result := query.Find(&MediaAccounts)
	if result.Error != nil && result.Error != gorm.ErrRecordNotFound {
		return MediaAccounts, result.Error
	}
	return MediaAccounts, nil
}

// UpsertModel 如果账号已存在，则更新；如果账号不存在，则插入新的账号
func (m *MediaAccount) UpsertModel() (*MediaAccount, error) {
	// 检查账号是否已存在
	var existingAccount MediaAccount
	err := database.DB.
		Where("media_code = ? AND user_id = ?", m.MediaCode, m.UserID).
		Where("deleted_at is null").
		First(&existingAccount).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}
	if err == gorm.ErrRecordNotFound {
		// 判断当前用户secUID在当前typ下是否存在
		err = database.DB.
			Where("type = ? AND uid = ?", m.Type, m.UID).
			Where("deleted_at is null").
			First(&existingAccount).Error
		if err != nil && err != gorm.ErrRecordNotFound {
			return nil, err
		}
	}
	// 如果账号存在，执行更新操作
	if err == nil {
		// 更新账号信息
		if updateErr := database.DB.Model(&existingAccount).Updates(m).Error; updateErr != nil {
			return nil, updateErr
		}
		// 返回更新后的数据
		return &existingAccount, nil
	} else {
		// 如果账号不存在，执行插入操作
		if insertErr := database.DB.Create(m).Error; insertErr != nil {
			return nil, insertErr
		}
		// 返回插入后的数据
		return m, nil
	}
}

func (m *MediaAccount) DeleteItems(ids []int) (bool, error) {
	model := &MediaAccount{}
	if updateErr := database.DB.Model(model).Where("id in (?)", ids).Updates(map[string]interface{}{
		"deleted_at": time.Now(), // 假设用 deleted 字段标记删除
	}).Error; updateErr != nil {
		return false, updateErr
	}
	return true, nil
}

// List 分页查询任务列表
func (m *MediaAccount) List(params *QueryMediaAccountParams, sort, order string, page, pageSize int) (database.PageResult[MediaAccount], error) {
	query := database.DB.Model(&MediaAccount{}).Where("deleted_at is null")
	if params.MediaCode != "" {
		query.Where("media_code = ?", params.MediaCode)
	}
	if params.Type > 0 {
		query.Where("type = ?", params.Type)
	}
	if params.Status > 0 {
		query.Where("status = ?", params.Status)
	}
	if params.ID > 0 {
		query.Where("id = ?", params.ID)
	}
	if len(params.UserID) > 0 {
		query.Where("user_id = ?", params.UserID)
	}
	if len(params.Username) > 0 {
		query.Where("username = ?", params.Username)
	}
	if len(params.RangeTime) > 0 {
		query = query.Where("update_time BETWEEN ? AND ?", params.RangeTime[0], params.RangeTime[1])
	}
	return database.Paginate[MediaAccount](query, database.ListOptions{
		Page:     page,
		PageSize: pageSize,
		Sort:     sort,
		Order:    order,
	})
}
