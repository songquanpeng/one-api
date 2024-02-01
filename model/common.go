package model

import (
	"fmt"
	"one-api/common"
)

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
