package controller

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/songquanpeng/one-api/common/config"
	"github.com/songquanpeng/one-api/common/ctxkey"
	"github.com/songquanpeng/one-api/common/helper"
	"github.com/songquanpeng/one-api/common/logger"
	"github.com/songquanpeng/one-api/model"
	relaymodel "github.com/songquanpeng/one-api/relay/model"
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
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data":    tasks,
	})
}

func GetAdminVideoTask(c *gin.Context) {
	task, err := model.GetVideoTaskByTaskID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "视频任务不存在",
		})
		return
	}
	events, err := model.GetVideoTaskEvents(task.TaskID)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data": gin.H{
			"task":   task,
			"events": events,
		},
	})
}

func CreateAdminDebugVideoTask(c *gin.Context) {
	var req createVideoTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	normalizeCreateVideoTaskRequest(&req)
	if req.Prompt == "" {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "prompt is required",
		})
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
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	go runFakeVideoTask(task.TaskID, defaultFakeVideoTaskTimings(req.TestMode))
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data":    task,
	})
}

func CancelAdminVideoTask(c *gin.Context) {
	taskID := c.Param("id")
	if err := model.CancelVideoTask(taskID, "admin"); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	task, err := model.GetVideoTaskByTaskID(taskID)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data":    task,
	})
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

func logFakeVideoTaskTransitionError(taskID string, nextStatus string, err error) {
	if err == nil || errors.Is(err, model.ErrVideoTaskStateConflict) {
		return
	}
	logger.SysError(fmt.Sprintf("fake video task %s transition to %s failed: %s", taskID, nextStatus, err.Error()))
}

func fakeVideoTaskErrorResponse(code string, message string, details string) string {
	response := map[string]any{
		"error": map[string]string{
			"code":    code,
			"message": message,
			"details": details,
		},
	}
	raw, err := json.Marshal(response)
	if err != nil {
		return fmt.Sprintf(`{"error":{"code":%q,"message":%q}}`, code, message)
	}
	return string(raw)
}

func runFakeVideoTask(taskID string, timings fakeVideoTaskTimings) {
	time.Sleep(timings.StartDelay)
	task, err := model.GetVideoTaskByTaskID(taskID)
	if err != nil || model.IsVideoTaskTerminal(task.Status) {
		return
	}
	if err := model.TransitionVideoTask(taskID, model.VideoTaskStatusRunning, "task.started", "fake worker started", map[string]any{"provider": "fake"}); err != nil {
		logFakeVideoTaskTransitionError(taskID, model.VideoTaskStatusRunning, err)
		return
	}
	time.Sleep(timings.CompleteDelay)
	task, err = model.GetVideoTaskByTaskID(taskID)
	if err != nil || model.IsVideoTaskTerminal(task.Status) {
		return
	}

	var params createVideoTaskRequest
	if err := json.Unmarshal([]byte(task.ParamsJSON), &params); err != nil {
		errorResponse := fakeVideoTaskErrorResponse("invalid_video_task_request", "invalid video task params", err.Error())
		err = model.TransitionVideoTask(taskID, model.VideoTaskStatusFailed, "task.failed", "fake worker failed", map[string]any{
			"error_code":             "invalid_video_task_request",
			"error_message":          "invalid video task params",
			"provider_response_json": errorResponse,
		})
		logFakeVideoTaskTransitionError(taskID, model.VideoTaskStatusFailed, err)
		return
	}
	if params.TestMode == "fail" {
		errorResponse := `{"error":{"code":"fake_upstream_error","message":"Fake provider simulated failure"}}`
		err = model.TransitionVideoTask(taskID, model.VideoTaskStatusFailed, "task.failed", "fake worker failed", map[string]any{
			"error_code":             "fake_upstream_error",
			"error_message":          "Fake provider simulated failure",
			"provider_response_json": errorResponse,
		})
		logFakeVideoTaskTransitionError(taskID, model.VideoTaskStatusFailed, err)
		return
	}

	resultJSON := `{"video_url":"https://example.com/fake-video.mp4","cover_url":"https://example.com/fake-cover.jpg","duration":5,"metadata":{"fake":true}}`
	err = model.TransitionVideoTask(taskID, model.VideoTaskStatusSucceeded, "task.succeeded", "fake worker succeeded", map[string]any{
		"result_json":            resultJSON,
		"provider_response_json": resultJSON,
		"video_url":              "https://example.com/fake-video.mp4",
	})
	logFakeVideoTaskTransitionError(taskID, model.VideoTaskStatusSucceeded, err)
}
