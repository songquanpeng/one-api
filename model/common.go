package model

import (
	"fmt"
	"one-api/common"
	"strings"

	"gorm.io/gorm"
)

type GenericParams struct {
	PaginationParams
	Keyword string `form:"keyword"`
}

type PaginationParams struct {
	Page  int    `form:"page"`
	Size  int    `form:"size"`
	Order string `form:"order"`
}

type DataResult struct {
	Data       interface{} `json:"data"`
	Page       int         `json:"page"`
	Size       int         `json:"size"`
	TotalCount int64       `json:"total_count"`
}

func PaginateAndOrder(db *gorm.DB, params *PaginationParams, result interface{}, allowedOrderFields map[string]bool) (*DataResult, error) {
	// 获取总数
	var totalCount int64
	err := db.Model(result).Count(&totalCount).Error
	if err != nil {
		return nil, err
	}

	fmt.Println("totalCount", totalCount)

	// 分页
	if params.Page < 1 {
		params.Page = 1
	}
	if params.Size < 1 {
		params.Size = common.ItemsPerPage
	}

	if params.Size > common.MaxRecentItems {
		return nil, fmt.Errorf("size 参数不能超过 %d", common.MaxRecentItems)
	}

	offset := (params.Page - 1) * params.Size
	db = db.Offset(offset).Limit(params.Size)

	// 排序
	if params.Order != "" {
		orderFields := strings.Split(params.Order, ",")
		for _, field := range orderFields {
			field = strings.TrimSpace(field)
			desc := strings.HasPrefix(field, "-")
			if desc {
				field = field[1:]
			}
			if !allowedOrderFields[field] {
				return nil, fmt.Errorf("不允许对字段 '%s' 进行排序", field)
			}
			if desc {
				field = field + " DESC"
			}
			db = db.Order(field)
		}
	} else {
		// 默认排序
		db = db.Order("id DESC")
	}

	// 查询
	err = db.Find(result).Error
	if err != nil {
		return nil, err
	}

	// 返回结果
	return &DataResult{
		Data:       result,
		Page:       params.Page,
		Size:       params.Size,
		TotalCount: totalCount,
	}, nil
}

func getDateFormat(groupType string) string {
	var dateFormat string
	if groupType == "day" {
		dateFormat = "%Y-%m-%d"
		if common.UsingPostgreSQL {
			dateFormat = "YYYY-MM-DD"
		}
	} else {
		dateFormat = "%Y-%m"
		if common.UsingPostgreSQL {
			dateFormat = "YYYY-MM"
		}
	}
	return dateFormat
}

func getTimestampGroupsSelect(fieldName, groupType, alias string) string {
	dateFormat := getDateFormat(groupType)
	var groupSelect string

	if common.UsingPostgreSQL {
		groupSelect = fmt.Sprintf(`TO_CHAR(date_trunc('%s', to_timestamp(%s)), '%s') as %s`, groupType, fieldName, dateFormat, alias)
	} else if common.UsingSQLite {
		groupSelect = fmt.Sprintf(`strftime('%s', datetime(%s, 'unixepoch')) as %s`, dateFormat, fieldName, alias)
	} else {
		groupSelect = fmt.Sprintf(`DATE_FORMAT(FROM_UNIXTIME(%s), '%s') as %s`, fieldName, dateFormat, alias)
	}

	return groupSelect
}
