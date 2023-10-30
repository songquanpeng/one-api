package model

import (
	"one-api/common"
)

type LogText struct {
	Id         int    `json:"id"`
	UserId     int    `json:"user_id" gorm:"index"`
	CreatedAt  int64  `json:"created_at" gorm:"index"`
	Username   string `json:"username" gorm:"index;default:''"`
	TokenName  string `json:"token_name" gorm:"index;default:''"`
	Prompt     string `json:"prompt" gorm:"type:text"`
	Completion string `json:"completion" gorm:"type:text"`
}

func RecordConsumeText(userId int, token string, prompt string, completion string) {

	text := &LogText{
		UserId:     userId,
		Username:   GetUsernameById(userId),
		CreatedAt:  common.GetTimestamp(),
		TokenName:  token,
		Prompt:     prompt,
		Completion: completion,
	}
	err := DB.Create(text).Error
	if err != nil {
		common.SysError("failed to record text: " + err.Error())
	}
}
