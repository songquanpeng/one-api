package controller

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/songquanpeng/one-api/common/ctxkey"
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

func decodeVideoTaskAdminResponse(t *testing.T, body *bytes.Buffer) map[string]any {
	t.Helper()
	var resp map[string]any
	if err := json.Unmarshal(body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	return resp
}

func TestGetAllVideoTasksFiltersAndNormalizesPage(t *testing.T) {
	setupControllerVideoTaskDB(t)
	gin.SetMode(gin.TestMode)

	matching := &model.VideoTask{
		TaskID:        "task_admin_list_match",
		UserID:        7,
		Model:         "fake-admin-model",
		Prompt:        "matching prompt",
		Status:        model.VideoTaskStatusQueued,
		Provider:      model.VideoTaskProviderFake,
		BillingStatus: model.VideoTaskBillingPending,
	}
	if err := model.CreateVideoTask(matching); err != nil {
		t.Fatalf("create matching task: %v", err)
	}
	other := &model.VideoTask{
		TaskID:        "task_admin_list_other",
		UserID:        8,
		Model:         "fake-admin-model",
		Prompt:        "other prompt",
		Status:        model.VideoTaskStatusQueued,
		Provider:      model.VideoTaskProviderFake,
		BillingStatus: model.VideoTaskBillingPending,
	}
	if err := model.CreateVideoTask(other); err != nil {
		t.Fatalf("create other task: %v", err)
	}

	router := gin.New()
	router.GET("/api/video_task/", GetAllVideoTasks)
	req := httptest.NewRequest(http.MethodGet, "/api/video_task/?p=-1&status=queued&model=fake-admin-model&user_id=7&task_id=task_admin_list_match", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}
	resp := decodeVideoTaskAdminResponse(t, rec.Body)
	if resp["success"] != true {
		t.Fatalf("success = %v, message = %v", resp["success"], resp["message"])
	}
	data, ok := resp["data"].([]any)
	if !ok {
		t.Fatalf("data type = %T", resp["data"])
	}
	if len(data) != 1 {
		t.Fatalf("task count = %d, want 1", len(data))
	}
	task := data[0].(map[string]any)
	if task["id"] != matching.TaskID {
		t.Fatalf("task id = %v, want %s", task["id"], matching.TaskID)
	}
}

func TestGetAdminVideoTaskReturnsTaskAndEvents(t *testing.T) {
	setupControllerVideoTaskDB(t)
	gin.SetMode(gin.TestMode)

	task := &model.VideoTask{
		TaskID:        "task_admin_detail",
		UserID:        7,
		Model:         "fake-admin-model",
		Prompt:        "detail prompt",
		Status:        model.VideoTaskStatusQueued,
		Provider:      model.VideoTaskProviderFake,
		BillingStatus: model.VideoTaskBillingPending,
	}
	if err := model.CreateVideoTask(task); err != nil {
		t.Fatalf("create task: %v", err)
	}

	router := gin.New()
	router.GET("/api/video_task/:id", GetAdminVideoTask)
	req := httptest.NewRequest(http.MethodGet, "/api/video_task/"+task.TaskID, nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}
	resp := decodeVideoTaskAdminResponse(t, rec.Body)
	if resp["success"] != true {
		t.Fatalf("success = %v, message = %v", resp["success"], resp["message"])
	}
	data := resp["data"].(map[string]any)
	if data["task"].(map[string]any)["id"] != task.TaskID {
		t.Fatalf("task id = %v, want %s", data["task"].(map[string]any)["id"], task.TaskID)
	}
	events := data["events"].([]any)
	if len(events) == 0 {
		t.Fatalf("events must include task.created")
	}
}

func TestCreateAdminDebugVideoTaskCreatesQueuedFakeTask(t *testing.T) {
	setupControllerVideoTaskDB(t)
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.POST("/api/video_task/debug", func(c *gin.Context) {
		c.Set(ctxkey.Id, 42)
		CreateAdminDebugVideoTask(c)
	})
	body := bytes.NewBufferString(`{"prompt":"debug prompt","test_mode":"slow"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/video_task/debug", body)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}
	resp := decodeVideoTaskAdminResponse(t, rec.Body)
	if resp["success"] != true {
		t.Fatalf("success = %v, message = %v", resp["success"], resp["message"])
	}
	taskData := resp["data"].(map[string]any)
	taskID := taskData["id"].(string)
	task, err := model.GetVideoTaskByTaskID(taskID)
	if err != nil {
		t.Fatalf("get task: %v", err)
	}
	if task.UserID != 42 {
		t.Fatalf("user_id = %d, want 42", task.UserID)
	}
	if task.TokenID != 0 {
		t.Fatalf("token_id = %d, want 0", task.TokenID)
	}
	if task.Status != model.VideoTaskStatusQueued {
		t.Fatalf("status = %s, want queued", task.Status)
	}
	if task.Provider != model.VideoTaskProviderFake {
		t.Fatalf("provider = %s, want fake", task.Provider)
	}
	if err := model.CancelVideoTask(taskID, "test"); err != nil {
		t.Fatalf("cancel created task: %v", err)
	}
}

func TestCancelAdminVideoTaskReturnsRefreshedTask(t *testing.T) {
	setupControllerVideoTaskDB(t)
	gin.SetMode(gin.TestMode)

	task := &model.VideoTask{
		TaskID:        "task_admin_cancel",
		UserID:        7,
		Model:         "fake-admin-model",
		Prompt:        "cancel prompt",
		Status:        model.VideoTaskStatusQueued,
		Provider:      model.VideoTaskProviderFake,
		BillingStatus: model.VideoTaskBillingPending,
	}
	if err := model.CreateVideoTask(task); err != nil {
		t.Fatalf("create task: %v", err)
	}

	router := gin.New()
	router.POST("/api/video_task/:id/cancel", CancelAdminVideoTask)
	req := httptest.NewRequest(http.MethodPost, "/api/video_task/"+task.TaskID+"/cancel", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}
	resp := decodeVideoTaskAdminResponse(t, rec.Body)
	if resp["success"] != true {
		t.Fatalf("success = %v, message = %v", resp["success"], resp["message"])
	}
	taskData := resp["data"].(map[string]any)
	if taskData["status"] != model.VideoTaskStatusCanceled {
		t.Fatalf("status = %v, want canceled", taskData["status"])
	}
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
