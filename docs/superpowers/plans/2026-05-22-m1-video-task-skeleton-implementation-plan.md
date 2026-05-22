# M1 Video Task Skeleton Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build the M1 video task skeleton: public video task API, admin observability pages, two-table persistence, and a fake provider that proves success, failure, and cancellation flows.

**Architecture:** Add a video task domain beside the existing relay stack rather than forcing video generation into `controller.Relay`. Persist current task state in `video_tasks`, persist the audit timeline in `video_task_events`, and use a small in-process fake worker for M1 state transitions. Expose developer APIs under `/v1/videos/...` with `TokenAuth`, and admin/debug APIs under `/api/video_task/...` with `AdminAuth`.

**Tech Stack:** Go + Gin + GORM for backend; React + Semantic UI + i18next for `web/default`; existing `go test ./...` and `npm test` / `npm run build` verification.

---

## Scope Check

This plan covers one subsystem: the M1 asynchronous video task skeleton. It intentionally excludes real Seedance integration, real quota deduction, ledger entries, object storage, webhook delivery, and the user-facing Playground.

## File Structure

Backend files:

- Create `model/video_task.go`: GORM models, status constants, validation, state transitions, query helpers.
- Create `model/video_task_test.go`: state machine, query, cancellation, and event persistence tests.
- Create `controller/video_task.go`: public API handlers and admin API handlers.
- Create `controller/video_task_test.go`: controller tests for auth-scoped API behavior and admin behavior.
- Modify `model/main.go`: add `AutoMigrate(&VideoTask{}, &VideoTaskEvent{})`.
- Modify `router/relay.go`: add public `/v1/videos/generations`, `/v1/videos/tasks/:id`, `/v1/videos/tasks/:id/cancel` routes before the general relay group.
- Modify `router/api.go`: add admin `/api/video_task` routes.

Frontend files:

- Create `web/default/src/pages/VideoTask/index.js`: admin page wrapper.
- Create `web/default/src/pages/VideoTask/VideoTasksTable.js`: list page, filters, status badges, cancel action, debug create modal.
- Create `web/default/src/pages/VideoTask/VideoTaskDetail.js`: detail page with task facts, provider fields, result/error, and event timeline.
- Modify `web/default/src/App.js`: routes for `/video-task` and `/video-task/:id`.
- Modify `web/default/src/components/Header.js`: admin-only navigation item.
- Modify `web/default/src/locales/zh/translation.json`: Chinese labels.
- Modify `web/default/src/locales/en/translation.json`: English labels.

Docs and tracking:

- Modify `docs/framebridge-implementation-progress.md`: mark M1 tasks as in progress after backend work starts and complete after verification.

## Cross-Cutting Rules

- Keep `provider_jobs` out of M1. Store M1 fake provider fields on `video_tasks`.
- Every state change must write exactly one `video_task_events` row.
- Terminal states are immutable: `succeeded`, `failed`, `canceled`.
- Public API must return OpenAI-style error payloads for request failures.
- Admin APIs can use the existing `{ success, message, data }` envelope.
- Backend timestamps use Unix seconds, matching existing project style.
- Use short polling in frontend detail page for non-terminal states; stop polling at terminal states.

---

### Task 1: Add Video Task Models And State Machine

**Files:**

- Create: `model/video_task.go`
- Create: `model/video_task_test.go`
- Modify: `model/main.go`

- [ ] **Step 1: Write failing model tests**

Create `model/video_task_test.go`:

```go
package model

import (
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupVideoTaskTestDB(t *testing.T) {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&VideoTask{}, &VideoTaskEvent{}); err != nil {
		t.Fatalf("migrate video task tables: %v", err)
	}
	DB = db
}

func TestVideoTaskStateTransitionsWriteEvents(t *testing.T) {
	setupVideoTaskTestDB(t)

	task := &VideoTask{
		TaskID:        "task_state_success",
		UserID:        1,
		TokenID:       2,
		Model:         "fake-seedance-lite",
		Prompt:        "hello video",
		ParamsJSON:    `{"test_mode":"success"}`,
		Status:        VideoTaskStatusQueued,
		Provider:      VideoTaskProviderFake,
		BillingStatus: VideoTaskBillingPending,
	}
	if err := CreateVideoTask(task); err != nil {
		t.Fatalf("create task: %v", err)
	}
	if err := TransitionVideoTask(task.TaskID, VideoTaskStatusRunning, "task.started", "fake worker started", map[string]any{"provider": "fake"}); err != nil {
		t.Fatalf("transition running: %v", err)
	}
	if err := TransitionVideoTask(task.TaskID, VideoTaskStatusSucceeded, "task.succeeded", "fake worker succeeded", map[string]any{"video_url": "https://example.com/fake-video.mp4"}); err != nil {
		t.Fatalf("transition succeeded: %v", err)
	}

	reloaded, err := GetVideoTaskByTaskID(task.TaskID)
	if err != nil {
		t.Fatalf("get task: %v", err)
	}
	if reloaded.Status != VideoTaskStatusSucceeded {
		t.Fatalf("status = %s, want %s", reloaded.Status, VideoTaskStatusSucceeded)
	}
	events, err := GetVideoTaskEvents(task.TaskID)
	if err != nil {
		t.Fatalf("get events: %v", err)
	}
	if len(events) != 3 {
		t.Fatalf("events len = %d, want 3", len(events))
	}
	if events[0].EventType != "task.created" || events[1].EventType != "task.started" || events[2].EventType != "task.succeeded" {
		t.Fatalf("unexpected events: %#v", events)
	}
}

func TestVideoTaskTerminalStateIsImmutable(t *testing.T) {
	setupVideoTaskTestDB(t)

	task := &VideoTask{
		TaskID:        "task_terminal",
		UserID:        1,
		TokenID:       2,
		Model:         "fake-seedance-lite",
		Prompt:        "hello video",
		Status:        VideoTaskStatusSucceeded,
		Provider:      VideoTaskProviderFake,
		BillingStatus: VideoTaskBillingPending,
	}
	if err := CreateVideoTask(task); err != nil {
		t.Fatalf("create task: %v", err)
	}

	err := TransitionVideoTask(task.TaskID, VideoTaskStatusCanceled, "task.canceled", "late cancel", nil)
	if err == nil {
		t.Fatalf("expected terminal state transition error")
	}
}

func TestCancelVideoTaskOnlyAllowsQueuedOrRunning(t *testing.T) {
	setupVideoTaskTestDB(t)

	task := &VideoTask{
		TaskID:        "task_cancel",
		UserID:        1,
		TokenID:       2,
		Model:         "fake-seedance-lite",
		Prompt:        "hello video",
		Status:        VideoTaskStatusQueued,
		Provider:      VideoTaskProviderFake,
		BillingStatus: VideoTaskBillingPending,
	}
	if err := CreateVideoTask(task); err != nil {
		t.Fatalf("create task: %v", err)
	}
	if err := CancelVideoTask(task.TaskID, "admin"); err != nil {
		t.Fatalf("cancel task: %v", err)
	}
	reloaded, err := GetVideoTaskByTaskID(task.TaskID)
	if err != nil {
		t.Fatalf("get task: %v", err)
	}
	if reloaded.Status != VideoTaskStatusCanceled {
		t.Fatalf("status = %s, want %s", reloaded.Status, VideoTaskStatusCanceled)
	}
}
```

- [ ] **Step 2: Run model tests and verify they fail**

Run:

```bash
go test ./model -run 'TestVideoTask' -count=1
```

Expected: compile failure because `VideoTask`, `CreateVideoTask`, `TransitionVideoTask`, `CancelVideoTask`, and related constants do not exist.

- [ ] **Step 3: Implement models and transition helpers**

Create `model/video_task.go`:

```go
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
	CreatedAt           int64  `json:"created_at" gorm:"bigint;index"`
	UpdatedAt           int64  `json:"updated_at" gorm:"bigint;index"`
	StartedAt           int64  `json:"started_at" gorm:"bigint"`
	CompletedAt         int64  `json:"completed_at" gorm:"bigint"`
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
	if task.Status == "" {
		task.Status = VideoTaskStatusQueued
	}
	if task.Provider == "" {
		task.Provider = VideoTaskProviderFake
	}
	if task.ProviderStatus == "" {
		task.ProviderStatus = task.Status
	}
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
```

Modify `model/main.go` inside `migrateDB()` after `Log` migration:

```go
if err = DB.AutoMigrate(&VideoTask{}); err != nil {
	return err
}
if err = DB.AutoMigrate(&VideoTaskEvent{}); err != nil {
	return err
}
```

- [ ] **Step 4: Run model tests and verify they pass**

Run:

```bash
go test ./model -run 'TestVideoTask' -count=1
```

Expected: PASS.

- [ ] **Step 5: Run full Go tests**

Run:

```bash
go test ./...
```

Expected: PASS.

- [ ] **Step 6: Commit model layer**

```bash
git add model/main.go model/video_task.go model/video_task_test.go
git commit -m "feat: add video task model"
```

---

### Task 2: Add Fake Provider Worker

**Files:**

- Create: `controller/video_task.go`
- Create: `controller/video_task_test.go`

- [ ] **Step 1: Write failing fake worker tests**

Create `controller/video_task_test.go` with model setup and fake worker tests:

```go
package controller

import (
	"testing"
	"time"

	"github.com/songquanpeng/one-api/model"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupControllerVideoTaskDB(t *testing.T) {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&model.VideoTask{}, &model.VideoTaskEvent{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	model.DB = db
}

func TestRunFakeVideoTaskSuccess(t *testing.T) {
	setupControllerVideoTaskDB(t)

	task := &model.VideoTask{
		TaskID:        "task_fake_success",
		UserID:        1,
		TokenID:       1,
		Model:         "fake-seedance-lite",
		Prompt:        "success prompt",
		ParamsJSON:    `{"test_mode":"success"}`,
		Status:        model.VideoTaskStatusQueued,
		Provider:      model.VideoTaskProviderFake,
		BillingStatus: model.VideoTaskBillingPending,
	}
	if err := model.CreateVideoTask(task); err != nil {
		t.Fatalf("create task: %v", err)
	}

	runFakeVideoTask(task.TaskID, fakeVideoTaskTimings{StartDelay: time.Millisecond, CompleteDelay: time.Millisecond})

	reloaded, err := model.GetVideoTaskByTaskID(task.TaskID)
	if err != nil {
		t.Fatalf("get task: %v", err)
	}
	if reloaded.Status != model.VideoTaskStatusSucceeded {
		t.Fatalf("status = %s, want %s", reloaded.Status, model.VideoTaskStatusSucceeded)
	}
	if reloaded.ResultJSON == "" {
		t.Fatalf("result_json must be set")
	}
}

func TestRunFakeVideoTaskFailure(t *testing.T) {
	setupControllerVideoTaskDB(t)

	task := &model.VideoTask{
		TaskID:        "task_fake_failure",
		UserID:        1,
		TokenID:       1,
		Model:         "fake-seedance-lite",
		Prompt:        "failure prompt",
		ParamsJSON:    `{"test_mode":"fail"}`,
		Status:        model.VideoTaskStatusQueued,
		Provider:      model.VideoTaskProviderFake,
		BillingStatus: model.VideoTaskBillingPending,
	}
	if err := model.CreateVideoTask(task); err != nil {
		t.Fatalf("create task: %v", err)
	}

	runFakeVideoTask(task.TaskID, fakeVideoTaskTimings{StartDelay: time.Millisecond, CompleteDelay: time.Millisecond})

	reloaded, err := model.GetVideoTaskByTaskID(task.TaskID)
	if err != nil {
		t.Fatalf("get task: %v", err)
	}
	if reloaded.Status != model.VideoTaskStatusFailed {
		t.Fatalf("status = %s, want %s", reloaded.Status, model.VideoTaskStatusFailed)
	}
	if reloaded.ErrorCode != "fake_upstream_error" {
		t.Fatalf("error_code = %s, want fake_upstream_error", reloaded.ErrorCode)
	}
}

func TestRunFakeVideoTaskDoesNotAdvanceCanceledTask(t *testing.T) {
	setupControllerVideoTaskDB(t)

	task := &model.VideoTask{
		TaskID:        "task_fake_canceled",
		UserID:        1,
		TokenID:       1,
		Model:         "fake-seedance-lite",
		Prompt:        "slow prompt",
		ParamsJSON:    `{"test_mode":"slow"}`,
		Status:        model.VideoTaskStatusQueued,
		Provider:      model.VideoTaskProviderFake,
		BillingStatus: model.VideoTaskBillingPending,
	}
	if err := model.CreateVideoTask(task); err != nil {
		t.Fatalf("create task: %v", err)
	}
	if err := model.CancelVideoTask(task.TaskID, "test"); err != nil {
		t.Fatalf("cancel task: %v", err)
	}

	runFakeVideoTask(task.TaskID, fakeVideoTaskTimings{StartDelay: time.Millisecond, CompleteDelay: time.Millisecond})

	reloaded, err := model.GetVideoTaskByTaskID(task.TaskID)
	if err != nil {
		t.Fatalf("get task: %v", err)
	}
	if reloaded.Status != model.VideoTaskStatusCanceled {
		t.Fatalf("status = %s, want %s", reloaded.Status, model.VideoTaskStatusCanceled)
	}
}
```

- [ ] **Step 2: Run worker tests and verify they fail**

Run:

```bash
go test ./controller -run 'TestRunFakeVideoTask' -count=1
```

Expected: compile failure because `runFakeVideoTask` and `fakeVideoTaskTimings` do not exist.

- [ ] **Step 3: Implement fake provider worker helpers**

Create `controller/video_task.go` with this worker-focused content. Task 3 extends the same file with HTTP handlers.

```go
package controller

import (
	"encoding/json"
	"time"

	"github.com/songquanpeng/one-api/model"
)

type createVideoTaskRequest struct {
	Model         string `json:"model"`
	Prompt        string `json:"prompt"`
	InputImageURL string `json:"input_image_url"`
	Duration      int    `json:"duration"`
	AspectRatio   string `json:"aspect_ratio"`
	Resolution    string `json:"resolution"`
	TestMode      string `json:"test_mode"`
}

type fakeVideoTaskTimings struct {
	StartDelay    time.Duration
	CompleteDelay time.Duration
}

func defaultFakeVideoTaskTimings(testMode string) fakeVideoTaskTimings {
	if testMode == "slow" {
		return fakeVideoTaskTimings{StartDelay: 3 * time.Second, CompleteDelay: 60 * time.Second}
	}
	return fakeVideoTaskTimings{StartDelay: 500 * time.Millisecond, CompleteDelay: 1500 * time.Millisecond}
}

func normalizeCreateVideoTaskRequest(req *createVideoTaskRequest) {
	if req.Model == "" {
		req.Model = "fake-seedance-lite"
	}
	if req.Duration == 0 {
		req.Duration = 5
	}
	if req.AspectRatio == "" {
		req.AspectRatio = "16:9"
	}
	if req.Resolution == "" {
		req.Resolution = "720p"
	}
	if req.TestMode == "" {
		req.TestMode = "success"
	}
}

func runFakeVideoTask(taskID string, timings fakeVideoTaskTimings) {
	time.Sleep(timings.StartDelay)
	task, err := model.GetVideoTaskByTaskID(taskID)
	if err != nil || model.IsVideoTaskTerminal(task.Status) {
		return
	}
	if err := model.TransitionVideoTask(taskID, model.VideoTaskStatusRunning, "task.started", "fake worker started", map[string]any{"provider": "fake"}); err != nil {
		return
	}
	time.Sleep(timings.CompleteDelay)
	task, err = model.GetVideoTaskByTaskID(taskID)
	if err != nil || model.IsVideoTaskTerminal(task.Status) {
		return
	}

	var params createVideoTaskRequest
	_ = json.Unmarshal([]byte(task.ParamsJSON), &params)
	if params.TestMode == "fail" {
		errorResponse := `{"error":{"code":"fake_upstream_error","message":"Fake provider simulated failure"}}`
		_ = model.TransitionVideoTask(taskID, model.VideoTaskStatusFailed, "task.failed", "fake worker failed", map[string]any{
			"error_code":             "fake_upstream_error",
			"error_message":          "Fake provider simulated failure",
			"provider_response_json": errorResponse,
		})
		return
	}

	resultJSON := `{"video_url":"https://example.com/fake-video.mp4","cover_url":"https://example.com/fake-cover.jpg","duration":5,"metadata":{"fake":true}}`
	_ = model.TransitionVideoTask(taskID, model.VideoTaskStatusSucceeded, "task.succeeded", "fake worker succeeded", map[string]any{
		"result_json":             resultJSON,
		"provider_response_json":  resultJSON,
		"video_url":               "https://example.com/fake-video.mp4",
	})
}
```

- [ ] **Step 4: Run fake worker tests and verify they pass**

Run:

```bash
go test ./controller -run 'TestRunFakeVideoTask' -count=1
```

Expected: PASS.

- [ ] **Step 5: Commit fake provider worker**

```bash
git add controller/video_task.go controller/video_task_test.go
git commit -m "feat: add fake video task worker"
```

---

### Task 3: Add Public Video Task API

**Files:**

- Modify: `controller/video_task.go`
- Modify: `controller/video_task_test.go`
- Modify: `router/relay.go`

- [ ] **Step 1: Write failing public API handler tests**

Append to `controller/video_task_test.go`:

```go
func TestBuildVideoTaskResponseMapsResultAndError(t *testing.T) {
	task := &model.VideoTask{
		TaskID:        "task_response",
		Status:        model.VideoTaskStatusSucceeded,
		Model:         "fake-seedance-lite",
		Prompt:        "prompt",
		Provider:      "fake",
		ProviderJobID: "fake_job_response",
		ResultJSON:    `{"video_url":"https://example.com/fake-video.mp4","cover_url":"https://example.com/fake-cover.jpg"}`,
		CreatedAt:     1,
		UpdatedAt:     2,
		StartedAt:     3,
		CompletedAt:   4,
	}
	resp := buildVideoTaskResponse(task)
	if resp["id"] != "task_response" {
		t.Fatalf("id = %v", resp["id"])
	}
	if resp["object"] != "video.task" {
		t.Fatalf("object = %v", resp["object"])
	}
	if resp["result"] == nil {
		t.Fatalf("result must be present")
	}
	if resp["error"] != nil {
		t.Fatalf("error must be nil")
	}

	task.Status = model.VideoTaskStatusFailed
	task.ResultJSON = ""
	task.ErrorCode = "fake_upstream_error"
	task.ErrorMessage = "Fake provider simulated failure"
	resp = buildVideoTaskResponse(task)
	if resp["result"] != nil {
		t.Fatalf("result must be nil for failed task")
	}
	if resp["error"] == nil {
		t.Fatalf("error must be present")
	}
}
```

- [ ] **Step 2: Run the response test and verify it fails**

Run:

```bash
go test ./controller -run 'TestBuildVideoTaskResponseMapsResultAndError' -count=1
```

Expected: compile failure because `buildVideoTaskResponse` does not exist.

- [ ] **Step 3: Implement public API handlers and response builder**

Add imports to `controller/video_task.go`:

```go
import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/songquanpeng/one-api/common/ctxkey"
	"github.com/songquanpeng/one-api/common/helper"
	"github.com/songquanpeng/one-api/model"
	relaymodel "github.com/songquanpeng/one-api/relay/model"
)
```

Keep one import block; remove duplicates created by Task 2.

Add these helpers and handlers to `controller/video_task.go`:

```go
func CreateVideoTask(c *gin.Context) {
	var req createVideoTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondVideoTaskError(c, http.StatusBadRequest, "invalid_video_task_request", err.Error())
		return
	}
	normalizeCreateVideoTaskRequest(&req)
	if req.Prompt == "" {
		respondVideoTaskError(c, http.StatusBadRequest, "invalid_video_task_request", "prompt is required")
		return
	}
	paramsJSON, _ := json.Marshal(req)
	providerJobID := "fake_job_" + model.NewVideoTaskID()[5:]
	task := &model.VideoTask{
		UserID:              c.GetInt(ctxkey.Id),
		TokenID:             c.GetInt(ctxkey.TokenId),
		RequestID:           c.GetString(helper.RequestIdKey),
		Model:               req.Model,
		Prompt:              req.Prompt,
		InputImageURL:       req.InputImageURL,
		ParamsJSON:          string(paramsJSON),
		Status:              model.VideoTaskStatusQueued,
		Provider:            model.VideoTaskProviderFake,
		ProviderJobID:       providerJobID,
		ProviderStatus:      model.VideoTaskStatusQueued,
		ProviderRequestJSON: string(paramsJSON),
		BillingStatus:       model.VideoTaskBillingPending,
	}
	if err := model.CreateVideoTask(task); err != nil {
		respondVideoTaskError(c, http.StatusInternalServerError, "video_task_create_failed", err.Error())
		return
	}
	go runFakeVideoTask(task.TaskID, defaultFakeVideoTaskTimings(req.TestMode))
	c.JSON(http.StatusOK, buildVideoTaskResponse(task))
}

func GetVideoTask(c *gin.Context) {
	taskID := c.Param("id")
	task, err := model.GetUserVideoTask(taskID, c.GetInt(ctxkey.Id))
	if err != nil {
		respondVideoTaskError(c, http.StatusNotFound, "video_task_not_found", "video task not found")
		return
	}
	c.JSON(http.StatusOK, buildVideoTaskResponse(task))
}

func CancelVideoTask(c *gin.Context) {
	taskID := c.Param("id")
	task, err := model.GetUserVideoTask(taskID, c.GetInt(ctxkey.Id))
	if err != nil {
		respondVideoTaskError(c, http.StatusNotFound, "video_task_not_found", "video task not found")
		return
	}
	if err := model.CancelVideoTask(task.TaskID, "api"); err != nil {
		status := http.StatusConflict
		code := "video_task_state_conflict"
		if errors.Is(err, model.ErrVideoTaskNotCancelable) {
			code = "video_task_not_cancelable"
		}
		respondVideoTaskError(c, status, code, err.Error())
		return
	}
	task, _ = model.GetVideoTaskByTaskID(taskID)
	c.JSON(http.StatusOK, buildVideoTaskResponse(task))
}

func buildVideoTaskResponse(task *model.VideoTask) gin.H {
	resp := gin.H{
		"id":              task.TaskID,
		"object":          "video.task",
		"status":          task.Status,
		"model":           task.Model,
		"prompt":          task.Prompt,
		"provider":        task.Provider,
		"provider_job_id": task.ProviderJobID,
		"estimated_quota": task.EstimatedQuota,
		"billing_status":  task.BillingStatus,
		"created_at":      task.CreatedAt,
		"updated_at":      task.UpdatedAt,
		"started_at":      nullableTimestamp(task.StartedAt),
		"completed_at":    nullableTimestamp(task.CompletedAt),
		"canceled_at":     nullableTimestamp(task.CanceledAt),
		"result":          nil,
		"error":           nil,
	}
	if task.ResultJSON != "" {
		var result map[string]any
		if json.Unmarshal([]byte(task.ResultJSON), &result) == nil {
			resp["result"] = result
		}
	}
	if task.ErrorCode != "" {
		resp["error"] = gin.H{"code": task.ErrorCode, "message": task.ErrorMessage}
	}
	return resp
}

func nullableTimestamp(value int64) any {
	if value == 0 {
		return nil
	}
	return value
}

func respondVideoTaskError(c *gin.Context, status int, code string, message string) {
	c.JSON(status, gin.H{
		"error": relaymodel.Error{
			Message: message,
			Type:    "framebridge_video_task_error",
			Code:    code,
		},
	})
}
```

- [ ] **Step 4: Add public routes**

Modify `router/relay.go` so video task routes use `TokenAuth` but do not use `middleware.Distribute()`:

```go
videoRouter := router.Group("/v1/videos")
videoRouter.Use(middleware.TokenAuth())
{
	videoRouter.POST("/generations", controller.CreateVideoTask)
	videoRouter.GET("/tasks/:id", controller.GetVideoTask)
	videoRouter.POST("/tasks/:id/cancel", controller.CancelVideoTask)
}
```

Place this block before `relayV1Router := router.Group("/v1")`.

- [ ] **Step 5: Run public API tests**

Run:

```bash
go test ./controller -run 'TestBuildVideoTaskResponseMapsResultAndError|TestRunFakeVideoTask' -count=1
go test ./...
```

Expected: both commands PASS.

- [ ] **Step 6: Commit public API**

```bash
git add controller/video_task.go controller/video_task_test.go router/relay.go
git commit -m "feat: add video task public API"
```

---

### Task 4: Add Admin Video Task API

**Files:**

- Modify: `controller/video_task.go`
- Modify: `router/api.go`

- [ ] **Step 1: Add admin handlers**

Ensure `controller/video_task.go` imports the existing pagination config:

```go
import "github.com/songquanpeng/one-api/common/config"
```

Append to `controller/video_task.go`:

```go
func GetAllVideoTasks(c *gin.Context) {
	p, _ := strconv.Atoi(c.Query("p"))
	if p < 0 {
		p = 0
	}
	userID, _ := strconv.Atoi(c.Query("user_id"))
	startTimestamp, _ := strconv.ParseInt(c.Query("start_timestamp"), 10, 64)
	endTimestamp, _ := strconv.ParseInt(c.Query("end_timestamp"), 10, 64)
	tasks, err := model.ListVideoTasks(model.VideoTaskFilter{
		Status:         c.Query("status"),
		Model:          c.Query("model"),
		UserID:         userID,
		TaskID:         c.Query("task_id"),
		StartTimestamp: startTimestamp,
		EndTimestamp:   endTimestamp,
		Offset:         p * config.ItemsPerPage,
		Limit:          config.ItemsPerPage,
	})
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "", "data": tasks})
}

func GetAdminVideoTask(c *gin.Context) {
	task, err := model.GetVideoTaskByTaskID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "视频任务不存在"})
		return
	}
	events, err := model.GetVideoTaskEvents(task.TaskID)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "", "data": gin.H{"task": task, "events": events}})
}

func CreateAdminDebugVideoTask(c *gin.Context) {
	var req createVideoTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": err.Error()})
		return
	}
	normalizeCreateVideoTaskRequest(&req)
	if req.Prompt == "" {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "prompt is required"})
		return
	}
	paramsJSON, _ := json.Marshal(req)
	task := &model.VideoTask{
		UserID:              c.GetInt(ctxkey.Id),
		TokenID:             c.GetInt(ctxkey.TokenId),
		RequestID:           c.GetString(helper.RequestIdKey),
		Model:               req.Model,
		Prompt:              req.Prompt,
		InputImageURL:       req.InputImageURL,
		ParamsJSON:          string(paramsJSON),
		Status:              model.VideoTaskStatusQueued,
		Provider:            model.VideoTaskProviderFake,
		ProviderJobID:       "fake_job_" + model.NewVideoTaskID()[5:],
		ProviderStatus:      model.VideoTaskStatusQueued,
		ProviderRequestJSON: string(paramsJSON),
		BillingStatus:       model.VideoTaskBillingPending,
	}
	if err := model.CreateVideoTask(task); err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": err.Error()})
		return
	}
	go runFakeVideoTask(task.TaskID, defaultFakeVideoTaskTimings(req.TestMode))
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "", "data": task})
}

func CancelAdminVideoTask(c *gin.Context) {
	if err := model.CancelVideoTask(c.Param("id"), "admin"); err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": err.Error()})
		return
	}
	task, _ := model.GetVideoTaskByTaskID(c.Param("id"))
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "", "data": task})
}
```

- [ ] **Step 2: Add admin routes**

Modify `router/api.go` after `logRoute`:

```go
videoTaskRoute := apiRouter.Group("/video_task")
videoTaskRoute.Use(middleware.AdminAuth())
{
	videoTaskRoute.GET("/", controller.GetAllVideoTasks)
	videoTaskRoute.GET("/:id", controller.GetAdminVideoTask)
	videoTaskRoute.POST("/debug", controller.CreateAdminDebugVideoTask)
	videoTaskRoute.POST("/:id/cancel", controller.CancelAdminVideoTask)
}
```

- [ ] **Step 3: Run backend tests**

Run:

```bash
go test ./...
```

Expected: PASS.

- [ ] **Step 4: Commit admin API**

```bash
git add controller/video_task.go router/api.go
git commit -m "feat: add video task admin API"
```

---

### Task 5: Add Admin Video Task Frontend

**Files:**

- Create: `web/default/src/pages/VideoTask/index.js`
- Create: `web/default/src/pages/VideoTask/VideoTasksTable.js`
- Create: `web/default/src/pages/VideoTask/VideoTaskDetail.js`
- Modify: `web/default/src/App.js`
- Modify: `web/default/src/components/Header.js`
- Modify: `web/default/src/locales/zh/translation.json`
- Modify: `web/default/src/locales/en/translation.json`

- [ ] **Step 1: Add page wrapper**

Create `web/default/src/pages/VideoTask/index.js`:

```javascript
import React from 'react';
import { Card } from 'semantic-ui-react';
import { useTranslation } from 'react-i18next';
import VideoTasksTable from './VideoTasksTable';

const VideoTask = () => {
  const { t } = useTranslation();

  return (
    <div className='dashboard-container'>
      <Card fluid className='chart-card'>
        <Card.Content>
          <Card.Header className='header'>{t('video_task.title')}</Card.Header>
          <VideoTasksTable />
        </Card.Content>
      </Card>
    </div>
  );
};

export default VideoTask;
```

- [ ] **Step 2: Add list table with debug create modal**

Create `web/default/src/pages/VideoTask/VideoTasksTable.js`:

```javascript
import React, { useEffect, useState } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import {
  Button,
  Dropdown,
  Form,
  Label,
  Modal,
  Pagination,
  Table,
} from 'semantic-ui-react';
import { useTranslation } from 'react-i18next';
import { API, showError, showSuccess, timestamp2string } from '../../helpers';
import { ITEMS_PER_PAGE } from '../../constants';

const STATUS_OPTIONS = [
  { key: 'all', text: '全部', value: '' },
  { key: 'queued', text: 'queued', value: 'queued' },
  { key: 'running', text: 'running', value: 'running' },
  { key: 'succeeded', text: 'succeeded', value: 'succeeded' },
  { key: 'failed', text: 'failed', value: 'failed' },
  { key: 'canceled', text: 'canceled', value: 'canceled' },
];

const TEST_MODE_OPTIONS = [
  { key: 'success', text: 'success', value: 'success' },
  { key: 'fail', text: 'fail', value: 'fail' },
  { key: 'slow', text: 'slow', value: 'slow' },
];

function renderStatus(status) {
  const colorMap = {
    queued: 'grey',
    running: 'blue',
    succeeded: 'green',
    failed: 'red',
    canceled: 'orange',
  };
  return <Label basic color={colorMap[status] || 'grey'}>{status}</Label>;
}

function isCancelable(status) {
  return status === 'queued' || status === 'running';
}

const VideoTasksTable = () => {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const [tasks, setTasks] = useState([]);
  const [activePage, setActivePage] = useState(1);
  const [loading, setLoading] = useState(false);
  const [createOpen, setCreateOpen] = useState(false);
  const [filters, setFilters] = useState({
    status: '',
    model: '',
    user_id: '',
    task_id: '',
  });
  const [form, setForm] = useState({
    prompt: '',
    model: 'fake-seedance-lite',
    duration: 5,
    aspect_ratio: '16:9',
    resolution: '720p',
    test_mode: 'success',
  });

  const loadTasks = async (page) => {
    setLoading(true);
    const params = new URLSearchParams();
    params.set('p', String(page - 1));
    Object.entries(filters).forEach(([key, value]) => {
      if (value !== '') params.set(key, value);
    });
    const res = await API.get(`/api/video_task/?${params.toString()}`);
    const { success, message, data } = res.data;
    if (success) {
      setTasks(data || []);
    } else {
      showError(message);
    }
    setLoading(false);
  };

  useEffect(() => {
    loadTasks(activePage).then();
  }, [activePage]);

  const refresh = async () => {
    setActivePage(1);
    await loadTasks(1);
  };

  const cancelTask = async (taskId) => {
    const res = await API.post(`/api/video_task/${taskId}/cancel`);
    const { success, message } = res.data;
    if (success) {
      showSuccess(t('video_task.messages.cancel_success'));
      await refresh();
    } else {
      showError(message);
    }
  };

  const createDebugTask = async () => {
    const res = await API.post('/api/video_task/debug', form);
    const { success, message, data } = res.data;
    if (success) {
      showSuccess(t('video_task.messages.create_success'));
      setCreateOpen(false);
      navigate(`/video-task/${data.id}`);
    } else {
      showError(message);
    }
  };

  return (
    <>
      <Form>
        <Form.Group>
          <Form.Field width={3}>
            <label>{t('video_task.filters.status')}</label>
            <Dropdown
              selection
              options={STATUS_OPTIONS}
              value={filters.status}
              onChange={(_, { value }) => setFilters({ ...filters, status: value })}
            />
          </Form.Field>
          <Form.Input
            width={3}
            label={t('video_task.filters.model')}
            value={filters.model}
            onChange={(_, { value }) => setFilters({ ...filters, model: value })}
          />
          <Form.Input
            width={3}
            label={t('video_task.filters.user_id')}
            value={filters.user_id}
            onChange={(_, { value }) => setFilters({ ...filters, user_id: value })}
          />
          <Form.Input
            width={4}
            label={t('video_task.filters.task_id')}
            value={filters.task_id}
            onChange={(_, { value }) => setFilters({ ...filters, task_id: value })}
          />
          <Form.Button width={3} label=' ' onClick={refresh}>
            {t('video_task.buttons.query')}
          </Form.Button>
        </Form.Group>
      </Form>

      <Button primary onClick={() => setCreateOpen(true)}>
        {t('video_task.buttons.create_debug')}
      </Button>
      <Button onClick={refresh} loading={loading}>
        {t('video_task.buttons.refresh')}
      </Button>

      <Table basic='very' compact size='small'>
        <Table.Header>
          <Table.Row>
            <Table.HeaderCell>{t('video_task.table.task_id')}</Table.HeaderCell>
            <Table.HeaderCell>{t('video_task.table.status')}</Table.HeaderCell>
            <Table.HeaderCell>{t('video_task.table.model')}</Table.HeaderCell>
            <Table.HeaderCell>{t('video_task.table.user')}</Table.HeaderCell>
            <Table.HeaderCell>{t('video_task.table.token')}</Table.HeaderCell>
            <Table.HeaderCell>{t('video_task.table.provider')}</Table.HeaderCell>
            <Table.HeaderCell>{t('video_task.table.created_at')}</Table.HeaderCell>
            <Table.HeaderCell>{t('video_task.table.actions')}</Table.HeaderCell>
          </Table.Row>
        </Table.Header>
        <Table.Body>
          {tasks.map((task) => (
            <Table.Row key={task.id}>
              <Table.Cell><code>{task.id}</code></Table.Cell>
              <Table.Cell>{renderStatus(task.status)}</Table.Cell>
              <Table.Cell>{task.model}</Table.Cell>
              <Table.Cell>{task.user_id}</Table.Cell>
              <Table.Cell>{task.token_id}</Table.Cell>
              <Table.Cell>{task.provider}</Table.Cell>
              <Table.Cell>{timestamp2string(task.created_at)}</Table.Cell>
              <Table.Cell>
                <Button size='mini' as={Link} to={`/video-task/${task.id}`}>
                  {t('video_task.buttons.view')}
                </Button>
                {isCancelable(task.status) && (
                  <Button size='mini' color='orange' onClick={() => cancelTask(task.id)}>
                    {t('video_task.buttons.cancel')}
                  </Button>
                )}
              </Table.Cell>
            </Table.Row>
          ))}
        </Table.Body>
      </Table>

      <Pagination
        activePage={activePage}
        onPageChange={(_, { activePage }) => setActivePage(activePage)}
        totalPages={Math.max(1, Math.ceil((tasks.length + 1) / ITEMS_PER_PAGE))}
      />

      <Modal open={createOpen} onClose={() => setCreateOpen(false)} size='small'>
        <Modal.Header>{t('video_task.debug.title')}</Modal.Header>
        <Modal.Content>
          <Form>
            <Form.TextArea
              label={t('video_task.debug.prompt')}
              value={form.prompt}
              onChange={(_, { value }) => setForm({ ...form, prompt: value })}
            />
            <Form.Input label='Model' value={form.model} onChange={(_, { value }) => setForm({ ...form, model: value })} />
            <Form.Input label='Duration' type='number' value={form.duration} onChange={(_, { value }) => setForm({ ...form, duration: Number(value) })} />
            <Form.Input label='Aspect ratio' value={form.aspect_ratio} onChange={(_, { value }) => setForm({ ...form, aspect_ratio: value })} />
            <Form.Input label='Resolution' value={form.resolution} onChange={(_, { value }) => setForm({ ...form, resolution: value })} />
            <Form.Field>
              <label>Test mode</label>
              <Dropdown selection options={TEST_MODE_OPTIONS} value={form.test_mode} onChange={(_, { value }) => setForm({ ...form, test_mode: value })} />
            </Form.Field>
          </Form>
        </Modal.Content>
        <Modal.Actions>
          <Button onClick={() => setCreateOpen(false)}>{t('video_task.buttons.close')}</Button>
          <Button primary onClick={createDebugTask}>{t('video_task.buttons.create')}</Button>
        </Modal.Actions>
      </Modal>
    </>
  );
};

export default VideoTasksTable;
```

- [ ] **Step 3: Add detail page**

Create `web/default/src/pages/VideoTask/VideoTaskDetail.js`:

```javascript
import React, { useEffect, useState } from 'react';
import { useParams } from 'react-router-dom';
import { Button, Card, Grid, Header, Label, Segment, Table } from 'semantic-ui-react';
import { useTranslation } from 'react-i18next';
import { API, copy, showError, showSuccess, timestamp2string } from '../../helpers';

function isTerminal(status) {
  return status === 'succeeded' || status === 'failed' || status === 'canceled';
}

function renderJSON(value) {
  if (!value) return '-';
  try {
    return JSON.stringify(JSON.parse(value), null, 2);
  } catch (_) {
    return value;
  }
}

const VideoTaskDetail = () => {
  const { t } = useTranslation();
  const { id } = useParams();
  const [task, setTask] = useState(null);
  const [events, setEvents] = useState([]);

  const loadTask = async () => {
    const res = await API.get(`/api/video_task/${id}`);
    const { success, message, data } = res.data;
    if (success) {
      setTask(data.task);
      setEvents(data.events || []);
    } else {
      showError(message);
    }
  };

  useEffect(() => {
    loadTask().then();
    const timer = setInterval(() => {
      if (!task || !isTerminal(task.status)) {
        loadTask().then();
      }
    }, 2000);
    return () => clearInterval(timer);
  }, [id, task?.status]);

  const copyTaskId = async () => {
    if (await copy(id)) showSuccess(t('video_task.messages.copy_success'));
  };

  const copyQueryCommand = async () => {
    const command = `curl http://localhost:3000/v1/videos/tasks/${id} -H "Authorization: Bearer <token>"`;
    if (await copy(command)) showSuccess(t('video_task.messages.copy_success'));
  };

  if (!task) return <Segment loading />;

  return (
    <div className='dashboard-container'>
      <Card fluid className='chart-card'>
        <Card.Content>
          <Card.Header className='header'>{t('video_task.detail.title')}</Card.Header>
          <Button onClick={copyTaskId}>{t('video_task.buttons.copy_task_id')}</Button>
          <Button onClick={copyQueryCommand}>{t('video_task.buttons.copy_query')}</Button>
          <Grid columns={2} stackable>
            <Grid.Column>
              <Header as='h4'>{t('video_task.detail.basic')}</Header>
              <Table compact basic='very'>
                <Table.Body>
                  <Table.Row><Table.Cell>ID</Table.Cell><Table.Cell><code>{task.id}</code></Table.Cell></Table.Row>
                  <Table.Row><Table.Cell>Status</Table.Cell><Table.Cell><Label>{task.status}</Label></Table.Cell></Table.Row>
                  <Table.Row><Table.Cell>Model</Table.Cell><Table.Cell>{task.model}</Table.Cell></Table.Row>
                  <Table.Row><Table.Cell>Prompt</Table.Cell><Table.Cell>{task.prompt}</Table.Cell></Table.Row>
                  <Table.Row><Table.Cell>Provider</Table.Cell><Table.Cell>{task.provider}</Table.Cell></Table.Row>
                  <Table.Row><Table.Cell>Provider Job</Table.Cell><Table.Cell>{task.provider_job_id || '-'}</Table.Cell></Table.Row>
                  <Table.Row><Table.Cell>Created</Table.Cell><Table.Cell>{timestamp2string(task.created_at)}</Table.Cell></Table.Row>
                </Table.Body>
              </Table>
            </Grid.Column>
            <Grid.Column>
              <Header as='h4'>{t('video_task.detail.payload')}</Header>
              <pre>{renderJSON(task.params_json)}</pre>
              <Header as='h4'>{t('video_task.detail.result')}</Header>
              <pre>{renderJSON(task.result_json || task.provider_response_json)}</pre>
              {task.error_code && <Label color='red'>{task.error_code}: {task.error_message}</Label>}
            </Grid.Column>
          </Grid>

          <Header as='h4'>{t('video_task.detail.events')}</Header>
          <Table compact basic='very'>
            <Table.Header>
              <Table.Row>
                <Table.HeaderCell>Time</Table.HeaderCell>
                <Table.HeaderCell>Event</Table.HeaderCell>
                <Table.HeaderCell>From</Table.HeaderCell>
                <Table.HeaderCell>To</Table.HeaderCell>
                <Table.HeaderCell>Message</Table.HeaderCell>
              </Table.Row>
            </Table.Header>
            <Table.Body>
              {events.map((event) => (
                <Table.Row key={event.id}>
                  <Table.Cell>{timestamp2string(event.created_at)}</Table.Cell>
                  <Table.Cell>{event.event_type}</Table.Cell>
                  <Table.Cell>{event.from_status || '-'}</Table.Cell>
                  <Table.Cell>{event.to_status}</Table.Cell>
                  <Table.Cell>{event.message}</Table.Cell>
                </Table.Row>
              ))}
            </Table.Body>
          </Table>
        </Card.Content>
      </Card>
    </div>
  );
};

export default VideoTaskDetail;
```

- [ ] **Step 4: Wire routes and navigation**

Modify `web/default/src/App.js` imports:

```javascript
import VideoTask from './pages/VideoTask';
import VideoTaskDetail from './pages/VideoTask/VideoTaskDetail';
```

Add routes near other private admin routes:

```jsx
<Route
  path='/video-task'
  element={
    <PrivateRoute>
      <VideoTask />
    </PrivateRoute>
  }
/>
<Route
  path='/video-task/:id'
  element={
    <PrivateRoute>
      <VideoTaskDetail />
    </PrivateRoute>
  }
/>
```

Modify `web/default/src/components/Header.js` by inserting this object in `headerButtons` near log/dashboard:

```javascript
{
  name: 'header.video_task',
  to: '/video-task',
  icon: 'film',
  admin: true,
},
```

- [ ] **Step 5: Add translations**

Add to `web/default/src/locales/zh/translation.json` under `header`:

```json
"video_task": "视频任务"
```

Add a top-level `video_task` object in `web/default/src/locales/zh/translation.json`:

```json
"video_task": {
  "title": "视频任务",
  "filters": {
    "status": "状态",
    "model": "模型",
    "user_id": "用户 ID",
    "task_id": "任务 ID"
  },
  "table": {
    "task_id": "任务 ID",
    "status": "状态",
    "model": "模型",
    "user": "用户",
    "token": "令牌",
    "provider": "Provider",
    "created_at": "创建时间",
    "actions": "操作"
  },
  "buttons": {
    "query": "查询",
    "refresh": "刷新",
    "view": "查看",
    "cancel": "取消",
    "create_debug": "新建调试任务",
    "create": "创建",
    "close": "关闭",
    "copy_task_id": "复制任务 ID",
    "copy_query": "复制查询命令"
  },
  "debug": {
    "title": "新建调试任务",
    "prompt": "Prompt"
  },
  "detail": {
    "title": "视频任务详情",
    "basic": "基础信息",
    "payload": "请求参数",
    "result": "结果 / Provider 响应",
    "events": "事件时间线"
  },
  "messages": {
    "create_success": "视频任务已创建",
    "cancel_success": "视频任务已取消",
    "copy_success": "已复制"
  }
}
```

Add equivalent English strings to `web/default/src/locales/en/translation.json`.

- [ ] **Step 6: Run frontend tests and build**

Run:

```bash
cd web/default
npm test -- --watchAll=false
npm run build
```

Expected: test command exits 0; build exits 0 and updates `web/build/default`.

- [ ] **Step 7: Commit frontend**

```bash
git add web/default/src/App.js web/default/src/components/Header.js web/default/src/pages/VideoTask web/default/src/locales/zh/translation.json web/default/src/locales/en/translation.json web/build/default
git commit -m "feat: add video task admin UI"
```

---

### Task 6: End-To-End Verification And Tracking

**Files:**

- Modify: `docs/framebridge-implementation-progress.md`

- [ ] **Step 1: Run full backend and frontend verification**

Run:

```bash
go test ./...
cd web/default
npm test -- --watchAll=false
npm run build
```

Expected: all commands exit 0.

- [ ] **Step 2: Start local service**

Run:

```bash
cd /Users/bytedance/my_workspace/VideoBridge
go build -o /tmp/framebridge/framebridge-dev-server main.go
/tmp/framebridge/framebridge-dev-server --log-dir ./logs
```

Expected: service listens on `http://localhost:3000`. If port 3000 is already in use by the old dev server, stop the old process first.

- [ ] **Step 3: Verify public API with curl**

Create a user token in the existing UI if one is not available. Then run:

```bash
TOKEN='<your-token>'

curl -sS http://localhost:3000/v1/videos/generations \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer ${TOKEN}" \
  -d '{
    "model": "fake-seedance-lite",
    "prompt": "一只机械手在桌面拼装芯片",
    "duration": 5,
    "aspect_ratio": "16:9",
    "resolution": "720p",
    "test_mode": "success"
  }'
```

Expected: JSON task object with `status=queued` and `object=video.task`.

Query the task:

```bash
TASK_ID='<returned-task-id>'
sleep 3
curl -sS "http://localhost:3000/v1/videos/tasks/${TASK_ID}" \
  -H "Authorization: Bearer ${TOKEN}"
```

Expected: `status=succeeded` and `result.video_url` exists.

- [ ] **Step 4: Verify failure path**

Run:

```bash
curl -sS http://localhost:3000/v1/videos/generations \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer ${TOKEN}" \
  -d '{
    "model": "fake-seedance-lite",
    "prompt": "模拟失败",
    "test_mode": "fail"
  }'
```

Expected after polling: `status=failed` and `error.code=fake_upstream_error`.

- [ ] **Step 5: Verify cancellation path**

Run:

```bash
curl -sS http://localhost:3000/v1/videos/generations \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer ${TOKEN}" \
  -d '{
    "model": "fake-seedance-lite",
    "prompt": "模拟慢任务",
    "test_mode": "slow"
  }'
```

Then cancel immediately:

```bash
TASK_ID='<returned-slow-task-id>'
curl -sS -X POST "http://localhost:3000/v1/videos/tasks/${TASK_ID}/cancel" \
  -H "Authorization: Bearer ${TOKEN}"
```

Expected: `status=canceled`. Poll again after 5 seconds; expected status remains `canceled`.

- [ ] **Step 6: Verify admin UI**

Open:

```text
http://localhost:3000/video-task
```

Expected:

- Admin sees `视频任务` navigation item.
- List shows success, failure, and canceled tasks.
- `queued/running` rows show cancel action.
- Detail page shows task facts, provider fields, and event timeline.
- Debug create form creates a task and routes to detail.

- [ ] **Step 7: Update progress document**

Modify `docs/framebridge-implementation-progress.md`:

- Set `最后更新：2026-05-22`.
- Set M0 row status to `已完成`.
- Set M1 row status to `已完成` after all M1 verification passes.
- Set T04, T05, T06, T07 rows to `已完成`.

- [ ] **Step 8: Commit verification and tracking**

```bash
git add docs/framebridge-implementation-progress.md
git commit -m "docs: mark M1 video task skeleton complete"
```

---

## Final Verification Checklist

Run from repo root:

```bash
go test ./...
cd web/default
npm test -- --watchAll=false
npm run build
git diff --check
```

Expected:

- Go tests pass.
- React tests pass.
- React production build succeeds.
- `git diff --check` emits no whitespace errors.
- Browser smoke test at `http://localhost:3000/video-task` matches the M1 admin prototype.

## Implementation Notes

- If frontend build updates hashed assets in `web/build/default`, include those files in the frontend commit.
- Do not commit `.superpowers/brainstorm` files.
- Keep unrelated existing changes separate from these M1 implementation commits.
