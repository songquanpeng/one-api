package model

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/songquanpeng/one-api/common/helper"
	"github.com/songquanpeng/one-api/common/random"
	"gorm.io/gorm"
)

const (
	VideoTaskStatusQueued    = "queued"
	VideoTaskStatusRunning   = "running"
	VideoTaskStatusSucceeded = "succeeded"
	VideoTaskStatusFailed    = "failed"
	VideoTaskStatusCanceled  = "canceled"

	VideoTaskProviderFake   = "fake"
	VideoTaskBillingPending = "pending"
)

var (
	ErrVideoTaskNotFound       = errors.New("video task not found")
	ErrVideoTaskNotCancelable  = errors.New("video task not cancelable")
	ErrVideoTaskStateConflict  = errors.New("video task state conflict")
	ErrInvalidVideoTaskRequest = errors.New("invalid video task request")
)

type VideoTask struct {
	Id                   int    `json:"-" gorm:"primaryKey"`
	TaskID               string `json:"id" gorm:"type:varchar(64);uniqueIndex"`
	UserID               int    `json:"user_id" gorm:"index"`
	TokenID              int    `json:"token_id" gorm:"index"`
	ChannelID            int    `json:"channel_id" gorm:"index"`
	RequestID            string `json:"request_id" gorm:"type:varchar(191);index"`
	Model                string `json:"model" gorm:"type:varchar(191);index"`
	Prompt               string `json:"prompt" gorm:"type:text"`
	InputImageURL        string `json:"input_image_url" gorm:"type:text"`
	ParamsJSON           string `json:"params_json" gorm:"type:text"`
	Status               string `json:"status" gorm:"type:varchar(32);index"`
	Provider             string `json:"provider" gorm:"type:varchar(64);index"`
	ProviderJobID        string `json:"provider_job_id" gorm:"type:varchar(191);index"`
	ProviderStatus       string `json:"provider_status" gorm:"type:varchar(64)"`
	ProviderRequestJSON  string `json:"provider_request_json" gorm:"type:text"`
	ProviderResponseJSON string `json:"provider_response_json" gorm:"type:text"`
	EstimatedQuota       int64  `json:"estimated_quota" gorm:"bigint;default:0"`
	BillingStatus        string `json:"billing_status" gorm:"type:varchar(32);default:'pending'"`
	ResultJSON           string `json:"result_json" gorm:"type:text"`
	ErrorCode            string `json:"error_code" gorm:"type:varchar(191);index"`
	ErrorMessage         string `json:"error_message" gorm:"type:text"`
	CreatedAt            int64  `json:"created_at" gorm:"bigint;index"`
	UpdatedAt            int64  `json:"updated_at" gorm:"bigint;index"`
	StartedAt            int64  `json:"started_at" gorm:"bigint"`
	CompletedAt          int64  `json:"completed_at" gorm:"bigint"`
	CanceledAt           int64  `json:"canceled_at" gorm:"bigint"`
}

type VideoTaskEvent struct {
	Id           int    `json:"id" gorm:"primaryKey"`
	TaskID       string `json:"task_id" gorm:"type:varchar(64);index"`
	EventType    string `json:"event_type" gorm:"type:varchar(64);index"`
	FromStatus   string `json:"from_status" gorm:"type:varchar(32)"`
	ToStatus     string `json:"to_status" gorm:"type:varchar(32)"`
	Message      string `json:"message" gorm:"type:text"`
	MetadataJSON string `json:"metadata_json" gorm:"type:text"`
	CreatedAt    int64  `json:"created_at" gorm:"bigint;index"`
}

type VideoTaskFilter struct {
	Status         string
	Model          string
	UserID         int
	TaskID         string
	StartTimestamp int64
	EndTimestamp   int64
	Offset         int
	Limit          int
}

func NewVideoTaskID() string {
	return "task_" + random.GetRandomString(24)
}

func IsVideoTaskTerminal(status string) bool {
	return status == VideoTaskStatusSucceeded || status == VideoTaskStatusFailed || status == VideoTaskStatusCanceled
}

func CreateVideoTask(task *VideoTask) error {
	now := helper.GetTimestamp()
	if task.TaskID == "" {
		task.TaskID = NewVideoTaskID()
	}
	task.Status = VideoTaskStatusQueued
	if task.Provider == "" {
		task.Provider = VideoTaskProviderFake
	}
	task.ProviderStatus = VideoTaskStatusQueued
	if task.BillingStatus == "" {
		task.BillingStatus = VideoTaskBillingPending
	}
	task.CreatedAt = now
	task.UpdatedAt = now
	return DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(task).Error; err != nil {
			return err
		}
		return createVideoTaskEvent(tx, task.TaskID, "task.created", "", task.Status, "video task created", map[string]any{"provider": task.Provider})
	})
}

func GetVideoTaskByTaskID(taskID string) (*VideoTask, error) {
	var task VideoTask
	err := DB.Where("task_id = ?", taskID).First(&task).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrVideoTaskNotFound
	}
	return &task, err
}

func GetUserVideoTask(taskID string, userID int) (*VideoTask, error) {
	var task VideoTask
	err := DB.Where("task_id = ? AND user_id = ?", taskID, userID).First(&task).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrVideoTaskNotFound
	}
	return &task, err
}

func ListVideoTasks(filter VideoTaskFilter) ([]*VideoTask, error) {
	var tasks []*VideoTask
	query := DB.Order("id desc")
	if filter.Status != "" {
		query = query.Where("status = ?", filter.Status)
	}
	if filter.Model != "" {
		query = query.Where("model = ?", filter.Model)
	}
	if filter.UserID > 0 {
		query = query.Where("user_id = ?", filter.UserID)
	}
	if filter.TaskID != "" {
		query = query.Where("task_id = ?", filter.TaskID)
	}
	if filter.StartTimestamp > 0 {
		query = query.Where("created_at >= ?", filter.StartTimestamp)
	}
	if filter.EndTimestamp > 0 {
		query = query.Where("created_at <= ?", filter.EndTimestamp)
	}
	if filter.Limit > 0 {
		query = query.Limit(filter.Limit)
	}
	if filter.Offset > 0 {
		query = query.Offset(filter.Offset)
	}
	err := query.Find(&tasks).Error
	return tasks, err
}

func GetVideoTaskEvents(taskID string) ([]*VideoTaskEvent, error) {
	var events []*VideoTaskEvent
	err := DB.Where("task_id = ?", taskID).Order("id asc").Find(&events).Error
	return events, err
}

func TransitionVideoTask(taskID string, nextStatus string, eventType string, message string, metadata map[string]any) error {
	return DB.Transaction(func(tx *gorm.DB) error {
		var task VideoTask
		err := tx.Where("task_id = ?", taskID).First(&task).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrVideoTaskNotFound
		}
		if err != nil {
			return err
		}
		if IsVideoTaskTerminal(task.Status) {
			return fmt.Errorf("%w: terminal status %s", ErrVideoTaskStateConflict, task.Status)
		}
		if !isAllowedVideoTaskTransition(task.Status, nextStatus) {
			return fmt.Errorf("%w: %s -> %s", ErrVideoTaskStateConflict, task.Status, nextStatus)
		}
		now := helper.GetTimestamp()
		updates := map[string]any{
			"status":          nextStatus,
			"provider_status": nextStatus,
			"updated_at":      now,
		}
		if nextStatus == VideoTaskStatusRunning {
			updates["started_at"] = now
		}
		if nextStatus == VideoTaskStatusSucceeded || nextStatus == VideoTaskStatusFailed {
			updates["completed_at"] = now
		}
		if nextStatus == VideoTaskStatusCanceled {
			updates["canceled_at"] = now
		}
		if result, ok := metadata["result_json"].(string); ok {
			updates["result_json"] = result
		}
		if errorCode, ok := metadata["error_code"].(string); ok {
			updates["error_code"] = errorCode
		}
		if errorMessage, ok := metadata["error_message"].(string); ok {
			updates["error_message"] = errorMessage
		}
		if responseJSON, ok := metadata["provider_response_json"].(string); ok {
			updates["provider_response_json"] = responseJSON
		}
		if err := tx.Model(&VideoTask{}).Where("task_id = ?", taskID).Updates(updates).Error; err != nil {
			return err
		}
		return createVideoTaskEvent(tx, taskID, eventType, task.Status, nextStatus, message, metadata)
	})
}

func CancelVideoTask(taskID string, actor string) error {
	task, err := GetVideoTaskByTaskID(taskID)
	if err != nil {
		return err
	}
	if task.Status != VideoTaskStatusQueued && task.Status != VideoTaskStatusRunning {
		return ErrVideoTaskNotCancelable
	}
	return TransitionVideoTask(taskID, VideoTaskStatusCanceled, "task.canceled", "video task canceled", map[string]any{"actor": actor})
}

func isAllowedVideoTaskTransition(current string, next string) bool {
	switch current {
	case VideoTaskStatusQueued:
		return next == VideoTaskStatusRunning || next == VideoTaskStatusFailed || next == VideoTaskStatusCanceled
	case VideoTaskStatusRunning:
		return next == VideoTaskStatusSucceeded || next == VideoTaskStatusFailed || next == VideoTaskStatusCanceled
	default:
		return false
	}
}

func createVideoTaskEvent(tx *gorm.DB, taskID string, eventType string, fromStatus string, toStatus string, message string, metadata map[string]any) error {
	metadataJSON := "{}"
	if len(metadata) > 0 {
		raw, err := json.Marshal(metadata)
		if err != nil {
			return err
		}
		metadataJSON = string(raw)
	}
	event := VideoTaskEvent{
		TaskID:       taskID,
		EventType:    eventType,
		FromStatus:   fromStatus,
		ToStatus:     toStatus,
		Message:      message,
		MetadataJSON: metadataJSON,
		CreatedAt:    helper.GetTimestamp(),
	}
	return tx.Create(&event).Error
}
