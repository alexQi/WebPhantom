package database

import (
	"fmt"
	"gorm.io/gorm"
)

type ListOptions struct {
	Page     int
	PageSize int
	Sort     string
	Order    string
}

// PageResult 泛型分页结果
type PageResult[T any] struct {
	Page  int   `json:"page"`
	Total int64 `json:"total"`
	Items []T   `json:"items"`
}

// Paginate 泛型分页方法，适用于任何 GORM 结构体
func Paginate[T any](query *gorm.DB, opt ListOptions) (PageResult[T], error) {
	var total int64
	items := make([]T, 0, opt.PageSize)

	// 获取总记录数
	if err := query.Count(&total).Error; err != nil {
		return PageResult[T]{}, err
	}
	if len(opt.Sort) == 0 {
		opt.Sort = "id"
	}
	if len(opt.Order) == 0 {
		opt.Sort = "desc"
	}

	// 计算偏移量
	offset := (opt.Page - 1) * opt.PageSize

	// 执行分页查询
	if err := query.
		Order(fmt.Sprintf("%s %s", opt.Sort, opt.Order)).
		Offset(offset).
		Limit(opt.PageSize).
		Find(&items).Error; err != nil {
		return PageResult[T]{}, err
	}
	// 返回分页数据
	return PageResult[T]{
		Page:  opt.Page,
		Total: total,
		Items: items,
	}, nil
}
