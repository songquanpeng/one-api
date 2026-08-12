package model

import (
	"time"
)

type SiftQTaskBinding struct {
	TaskID    string `json:"task_id" gorm:"type:varchar(128);primaryKey;autoIncrement:false"`
	ChannelID int    `json:"channel_id" gorm:"primaryKey;autoIncrement:false;index"`
	UserID    int    `json:"user_id" gorm:"index:idx_siftq_user_created,priority:1"`
	TokenID   int    `json:"token_id" gorm:"index"`
	Model     string `json:"model" gorm:"type:varchar(64)"`
	Prompt    string `json:"prompt" gorm:"type:text"`
	Seconds   int    `json:"seconds"`
	Size      string `json:"size" gorm:"type:varchar(32)"`
	CreatedAt int64  `json:"created_at" gorm:"bigint;index:idx_siftq_user_created,priority:2"`
}

func SaveSiftQTaskBinding(binding *SiftQTaskBinding) error {
	if binding.CreatedAt == 0 {
		binding.CreatedAt = time.Now().Unix()
	}
	return DB.Create(binding).Error
}

func GetSiftQTaskBinding(userID int, taskID string) (*SiftQTaskBinding, error) {
	var binding SiftQTaskBinding
	err := DB.Where("user_id = ? AND task_id = ?", userID, taskID).First(&binding).Error
	return &binding, err
}

func ListSiftQTaskBindings(userID int) ([]SiftQTaskBinding, error) {
	var bindings []SiftQTaskBinding
	cutoff := time.Now().Add(-7 * 24 * time.Hour).Unix()
	err := DB.Where("user_id = ? AND created_at >= ?", userID, cutoff).Order("created_at DESC").Find(&bindings).Error
	return bindings, err
}

func DeleteSiftQTaskBinding(userID int, taskID string) error {
	return DB.Where("user_id = ? AND task_id = ?", userID, taskID).Delete(&SiftQTaskBinding{}).Error
}
