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

func TestVideoTaskCreateForcesQueuedStatus(t *testing.T) {
	setupVideoTaskTestDB(t)

	task := &VideoTask{
		TaskID:         "task_create_forces_queued",
		UserID:         1,
		TokenID:        2,
		Model:          "fake-seedance-lite",
		Prompt:         "hello video",
		Status:         VideoTaskStatusSucceeded,
		Provider:       VideoTaskProviderFake,
		ProviderStatus: VideoTaskStatusSucceeded,
		BillingStatus:  VideoTaskBillingPending,
	}
	if err := CreateVideoTask(task); err != nil {
		t.Fatalf("create task: %v", err)
	}

	reloaded, err := GetVideoTaskByTaskID(task.TaskID)
	if err != nil {
		t.Fatalf("get task: %v", err)
	}
	if reloaded.Status != VideoTaskStatusQueued {
		t.Fatalf("status = %s, want %s", reloaded.Status, VideoTaskStatusQueued)
	}
	if reloaded.ProviderStatus != VideoTaskStatusQueued {
		t.Fatalf("provider status = %s, want %s", reloaded.ProviderStatus, VideoTaskStatusQueued)
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
		Status:        VideoTaskStatusQueued,
		Provider:      VideoTaskProviderFake,
		BillingStatus: VideoTaskBillingPending,
	}
	if err := CreateVideoTask(task); err != nil {
		t.Fatalf("create task: %v", err)
	}
	if err := TransitionVideoTask(task.TaskID, VideoTaskStatusRunning, "task.started", "fake worker started", nil); err != nil {
		t.Fatalf("transition running: %v", err)
	}
	if err := TransitionVideoTask(task.TaskID, VideoTaskStatusSucceeded, "task.succeeded", "fake worker succeeded", nil); err != nil {
		t.Fatalf("transition succeeded: %v", err)
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

	runningTask := &VideoTask{
		TaskID:        "task_cancel_running",
		UserID:        1,
		TokenID:       2,
		Model:         "fake-seedance-lite",
		Prompt:        "hello video",
		Status:        VideoTaskStatusQueued,
		Provider:      VideoTaskProviderFake,
		BillingStatus: VideoTaskBillingPending,
	}
	if err := CreateVideoTask(runningTask); err != nil {
		t.Fatalf("create running task: %v", err)
	}
	if err := TransitionVideoTask(runningTask.TaskID, VideoTaskStatusRunning, "task.started", "fake worker started", nil); err != nil {
		t.Fatalf("transition running task: %v", err)
	}
	if err := CancelVideoTask(runningTask.TaskID, "admin"); err != nil {
		t.Fatalf("cancel running task: %v", err)
	}
	reloadedRunning, err := GetVideoTaskByTaskID(runningTask.TaskID)
	if err != nil {
		t.Fatalf("get running task: %v", err)
	}
	if reloadedRunning.Status != VideoTaskStatusCanceled {
		t.Fatalf("running status = %s, want %s", reloadedRunning.Status, VideoTaskStatusCanceled)
	}
}
