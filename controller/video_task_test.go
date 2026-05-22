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
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("get sql db: %v", err)
	}
	sqlDB.SetMaxOpenConns(1)
	model.DB = db
}

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

func TestRunFakeVideoTaskDoesNotAdvanceCanceledRunningTask(t *testing.T) {
	setupControllerVideoTaskDB(t)

	task := &model.VideoTask{
		TaskID:        "task_fake_canceled_running",
		UserID:        1,
		TokenID:       1,
		Model:         "fake-seedance-lite",
		Prompt:        "cancel while running prompt",
		ParamsJSON:    `{"test_mode":"success"}`,
		Status:        model.VideoTaskStatusQueued,
		Provider:      model.VideoTaskProviderFake,
		BillingStatus: model.VideoTaskBillingPending,
	}
	if err := model.CreateVideoTask(task); err != nil {
		t.Fatalf("create task: %v", err)
	}

	done := make(chan struct{})
	go func() {
		defer close(done)
		runFakeVideoTask(task.TaskID, fakeVideoTaskTimings{StartDelay: time.Millisecond, CompleteDelay: 50 * time.Millisecond})
	}()

	deadline := time.After(time.Second)
	for {
		reloaded, err := model.GetVideoTaskByTaskID(task.TaskID)
		if err != nil {
			t.Fatalf("get task: %v", err)
		}
		if reloaded.Status == model.VideoTaskStatusRunning {
			break
		}
		select {
		case <-deadline:
			t.Fatalf("task did not reach running status")
		default:
			time.Sleep(time.Millisecond)
		}
	}

	if err := model.CancelVideoTask(task.TaskID, "test"); err != nil {
		t.Fatalf("cancel task: %v", err)
	}

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatalf("fake worker did not finish")
	}

	reloaded, err := model.GetVideoTaskByTaskID(task.TaskID)
	if err != nil {
		t.Fatalf("get task: %v", err)
	}
	if reloaded.Status != model.VideoTaskStatusCanceled {
		t.Fatalf("status = %s, want %s", reloaded.Status, model.VideoTaskStatusCanceled)
	}
}

func TestRunFakeVideoTaskMalformedParamsFailsTask(t *testing.T) {
	setupControllerVideoTaskDB(t)

	task := &model.VideoTask{
		TaskID:        "task_fake_malformed_params",
		UserID:        1,
		TokenID:       1,
		Model:         "fake-seedance-lite",
		Prompt:        "malformed params prompt",
		ParamsJSON:    `{"test_mode":`,
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
	if reloaded.ErrorCode != "invalid_video_task_request" {
		t.Fatalf("error_code = %s, want invalid_video_task_request", reloaded.ErrorCode)
	}
	if reloaded.ErrorMessage != "invalid video task params" {
		t.Fatalf("error_message = %s, want invalid video task params", reloaded.ErrorMessage)
	}
	if reloaded.ProviderResponseJSON == "" {
		t.Fatalf("provider_response_json must be set")
	}
}
