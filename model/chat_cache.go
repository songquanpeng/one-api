package model

import (
	"time"

	"gorm.io/gorm/clause"
)

type ChatCache struct {
	Hash       string `json:"hash" gorm:"type:varchar(32);primaryKey"`
	UserId     int    `json:"user_id" gorm:"type:int;not null;index"`
	Data       string `json:"data" gorm:"type:json;not null"`
	Expiration int64  `json:"expiration" gorm:"type:bigint;not null;index"`
}

func (cache *ChatCache) Insert() error {
	return DB.Clauses(clause.OnConflict{
		UpdateAll: true,
	}).Create(cache).Error
}

func GetChatCache(hash string, userId int) (*ChatCache, error) {
	var chatCache ChatCache
	// 获取当前时间戳
	now := time.Now().Unix()
	err := DB.Where("hash = ? and user_id = ? and expiration > ?", hash, userId, now).Find(&chatCache).Error
	return &chatCache, err
}

func GetChatCacheListByUserId(userId int) ([]*ChatCache, error) {
	var chatCaches []*ChatCache
	// 获取当前时间戳
	now := time.Now().Unix()
	err := DB.Where("user_id = ? and expiration >", userId, now).Find(&chatCaches).Error
	return chatCaches, err
}

func RemoveChatCache() error {
	now := time.Now().Unix()
	return DB.Where("expiration < ?", now).Delete(ChatCache{}).Error
}
