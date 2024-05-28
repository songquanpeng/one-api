package model

import (
	"fmt"
	"one-api/common"
	"one-api/common/config"
	"strings"

	"gorm.io/gorm"
)

type modelable interface {
	any
}

type GenericParams struct {
	PaginationParams
	Keyword string `form:"keyword"`
}

type PaginationParams struct {
	Page  int    `form:"page"`
	Size  int    `form:"size"`
	Order string `form:"order"`
}

type DataResult[T modelable] struct {
	Data       *[]*T `json:"data"`
	Page       int   `json:"page"`
	Size       int   `json:"size"`
	TotalCount int64 `json:"total_count"`
}

func PaginateAndOrder[T modelable](db *gorm.DB, params *PaginationParams, result *[]*T, allowedOrderFields map[string]bool) (*DataResult[T], error) {
	// 获取总数
	var totalCount int64
	err := db.Model(result).Count(&totalCount).Error
	if err != nil {
		return nil, err
	}

	// 分页
	if params.Page < 1 {
		params.Page = 1
	}
	if params.Size < 1 {
		params.Size = config.ItemsPerPage
	}

	if params.Size > config.MaxRecentItems {
		return nil, fmt.Errorf("size 参数不能超过 %d", config.MaxRecentItems)
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
	return &DataResult[T]{
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

func quotePostgresField(field string) string {
	if common.UsingPostgreSQL {
		return fmt.Sprintf(`"%s"`, field)
	}

	return fmt.Sprintf("`%s`", field)
}

func assembleSumSelectStr(selectStr string) string {
	sumSelectStr := "%s(sum(%s),0)"
	nullfunc := "ifnull"
	if common.UsingPostgreSQL {
		nullfunc = "coalesce"
	}

	sumSelectStr = fmt.Sprintf(sumSelectStr, nullfunc, selectStr)

	return sumSelectStr
}

func RecordExists(table interface{}, fieldName string, fieldValue interface{}, excludeID interface{}) bool {
	var count int64
	query := DB.Model(table).Where(fmt.Sprintf("%s = ?", fieldName), fieldValue)
	if excludeID != nil {
		query = query.Not("id", excludeID)
	}
	query.Count(&count)
	return count > 0
}
