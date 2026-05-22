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
