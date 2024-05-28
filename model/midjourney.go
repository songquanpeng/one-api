// Copyright (c) 2024 Calcium-Ion
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.
//
// Author: Calcium-Ion
// GitHub: https://github.com/Calcium-Ion/new-api

package model

type Midjourney struct {
	Id          int    `json:"id"`
	Code        int    `json:"code"`
	UserId      int    `json:"user_id" gorm:"index"`
	Action      string `json:"action" gorm:"type:varchar(40);index"`
	MjId        string `json:"mj_id" gorm:"index"`
	Prompt      string `json:"prompt"`
	PromptEn    string `json:"prompt_en"`
	Description string `json:"description"`
	State       string `json:"state"`
	SubmitTime  int64  `json:"submit_time" gorm:"index"`
	StartTime   int64  `json:"start_time" gorm:"index"`
	FinishTime  int64  `json:"finish_time" gorm:"index"`
	ImageUrl    string `json:"image_url"`
	Status      string `json:"status" gorm:"type:varchar(20);index"`
	Progress    string `json:"progress" gorm:"type:varchar(30);index"`
	FailReason  string `json:"fail_reason"`
	ChannelId   int    `json:"channel_id"`
	Quota       int    `json:"quota"`
	Buttons     string `json:"buttons"`
	Properties  string `json:"properties"`
}

// TaskQueryParams 用于包含所有搜索条件的结构体，可以根据需求添加更多字段
type TaskQueryParams struct {
	ChannelID      int    `form:"channel_id"`
	MjID           string `form:"mj_id"`
	StartTimestamp int    `form:"start_timestamp"`
	EndTimestamp   int    `form:"end_timestamp"`
	PaginationParams
}

var allowedMidjourneyOrderFields = map[string]bool{
	"id":          true,
	"user_id":     true,
	"code":        true,
	"action":      true,
	"mj_id":       true,
	"submit_time": true,
	"start_time":  true,
	"finish_time": true,
	"status":      true,
	"channel_id":  true,
}

func GetAllUserTask(userId int, params *TaskQueryParams) (*DataResult[Midjourney], error) {
	var tasks []*Midjourney

	// 初始化查询构建器
	query := DB.Where("user_id = ?", userId)

	if params.MjID != "" {
		query = query.Where("mj_id = ?", params.MjID)
	}
	if params.StartTimestamp != 0 {
		// 假设您已将前端传来的时间戳转换为数据库所需的时间格式，并处理了时间戳的验证和解析
		query = query.Where("submit_time >= ?", params.StartTimestamp)
	}
	if params.EndTimestamp != 0 {
		query = query.Where("submit_time <= ?", params.EndTimestamp)
	}

	return PaginateAndOrder(query, &params.PaginationParams, &tasks, allowedMidjourneyOrderFields)
}

func GetAllTasks(params *TaskQueryParams) (*DataResult[Midjourney], error) {
	var tasks []*Midjourney

	// 初始化查询构建器
	query := DB

	// 添加过滤条件
	if params.ChannelID != 0 {
		query = query.Where("channel_id = ?", params.ChannelID)
	}
	if params.MjID != "" {
		query = query.Where("mj_id = ?", params.MjID)
	}
	if params.StartTimestamp != 0 {
		query = query.Where("submit_time >= ?", params.StartTimestamp)
	}
	if params.EndTimestamp != 0 {
		query = query.Where("submit_time <= ?", params.EndTimestamp)
	}

	return PaginateAndOrder(query, &params.PaginationParams, &tasks, allowedMidjourneyOrderFields)
}

func GetAllUnFinishTasks() []*Midjourney {
	var tasks []*Midjourney
	// get all tasks progress is not 100%
	err := DB.Where("progress != ?", "100%").Find(&tasks).Error
	if err != nil {
		return nil
	}
	return tasks
}

func GetByOnlyMJId(mjId string) *Midjourney {
	var mj *Midjourney
	err := DB.Where("mj_id = ?", mjId).First(&mj).Error
	if err != nil {
		return nil
	}
	return mj
}

func GetByMJId(userId int, mjId string) *Midjourney {
	var mj *Midjourney
	err := DB.Where("user_id = ? and mj_id = ?", userId, mjId).First(&mj).Error
	if err != nil {
		return nil
	}
	return mj
}

func GetByMJIds(userId int, mjIds []string) []*Midjourney {
	var mj []*Midjourney
	err := DB.Where("user_id = ? and mj_id in (?)", userId, mjIds).Find(&mj).Error
	if err != nil {
		return nil
	}
	return mj
}

func (midjourney *Midjourney) Insert() error {
	return DB.Create(midjourney).Error
}

func (midjourney *Midjourney) Update() error {
	return DB.Save(midjourney).Error
}

func MjBulkUpdate(mjIds []string, params map[string]any) error {
	return DB.Model(&Midjourney{}).
		Where("mj_id in (?)", mjIds).
		Updates(params).Error
}

func MjBulkUpdateByTaskIds(taskIDs []int, params map[string]any) error {
	return DB.Model(&Midjourney{}).
		Where("id in (?)", taskIDs).
		Updates(params).Error
}
