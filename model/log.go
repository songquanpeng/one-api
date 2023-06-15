package model

import (
	"gorm.io/gorm"
	"one-api/common"
)

type Log struct {
	Id        int    `json:"id"`
	UserId    int    `json:"user_id" gorm:"index"`
	CreatedAt int64  `json:"created_at" gorm:"bigint"`
	Type      int    `json:"type" gorm:"index"`
	Content   string `json:"content"`
}

const (
	LogTypeUnknown = iota
	LogTypeTopup
	LogTypeConsume
	LogTypeManage
	LogTypeSystem
)

func RecordLog(userId int, logType int, content string) {
	if logType == LogTypeConsume && !common.LogConsumeEnabled {
		return
	}
	log := &Log{
		UserId:    userId,
		CreatedAt: common.GetTimestamp(),
		Type:      logType,
		Content:   content,
	}
	err := DB.Create(log).Error
	if err != nil {
		common.SysError("failed to record log: " + err.Error())
	}
}

func GetAllLogs(logType int, startIdx int, num int) (logs []*Log, err error) {
	var tx *gorm.DB
	if logType == LogTypeUnknown {
		tx = DB
	} else {
		tx = DB.Where("type = ?", logType)
	}
	err = tx.Order("id desc").Limit(num).Offset(startIdx).Find(&logs).Error
	return logs, err
}

func GetUserLogs(userId int, logType int, startIdx int, num int) (logs []*Log, err error) {
	var tx *gorm.DB
	if logType == LogTypeUnknown {
		tx = DB.Where("user_id = ?", userId)
	} else {
		tx = DB.Where("user_id = ? and type = ?", userId, logType)
	}
	err = tx.Order("id desc").Limit(num).Offset(startIdx).Omit("id").Find(&logs).Error
	return logs, err
}

func SearchAllLogs(keyword string) (logs []*Log, err error) {
	err = DB.Where("type = ? or content LIKE ?", keyword, keyword+"%").Order("id desc").Limit(common.MaxRecentItems).Find(&logs).Error
	return logs, err
}

func SearchUserLogs(userId int, keyword string) (logs []*Log, err error) {
	err = DB.Where("user_id = ? and type = ?", userId, keyword).Order("id desc").Limit(common.MaxRecentItems).Omit("id").Find(&logs).Error
	return logs, err
}
