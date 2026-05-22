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
		"result_json":            resultJSON,
		"provider_response_json": resultJSON,
		"video_url":              "https://example.com/fake-video.mp4",
	})
}
