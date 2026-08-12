package controller

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"mime/multipart"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/songquanpeng/one-api/common/client"
	"github.com/songquanpeng/one-api/common/ctxkey"
	"github.com/songquanpeng/one-api/common/helper"
	"github.com/songquanpeng/one-api/common/logger"
	"github.com/songquanpeng/one-api/middleware"
	"github.com/songquanpeng/one-api/model"
	"github.com/songquanpeng/one-api/relay/adaptor/siftq"
	billingratio "github.com/songquanpeng/one-api/relay/billing/ratio"
	"github.com/songquanpeng/one-api/relay/channeltype"
	"github.com/songquanpeng/one-api/relay/meta"
)

const siftQResponseLimit = 8 << 20

type videoInputReference struct {
	FileID   string `json:"file_id,omitempty"`
	ImageURL string `json:"image_url,omitempty"`
}

type videoCreateRequest struct {
	Prompt         string               `json:"prompt"`
	InputReference *videoInputReference `json:"input_reference,omitempty"`
	Model          string               `json:"model,omitempty"`
	Seconds        any                  `json:"seconds,omitempty"`
	Size           string               `json:"size,omitempty"`
}

type videoError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type videoObject struct {
	ID          string      `json:"id"`
	Object      string      `json:"object"`
	Model       string      `json:"model"`
	Status      string      `json:"status"`
	Progress    int         `json:"progress"`
	CreatedAt   int64       `json:"created_at"`
	CompletedAt *int64      `json:"completed_at,omitempty"`
	Prompt      string      `json:"prompt,omitempty"`
	Seconds     string      `json:"seconds"`
	Size        string      `json:"size"`
	Error       *videoError `json:"error,omitempty"`
}

func CreateVideo(c *gin.Context) {
	request, imageData, err := decodeVideoCreateRequest(c)
	if err != nil {
		writeVideoError(c, http.StatusBadRequest, "invalid_request_error", "invalid_request", err.Error())
		return
	}
	providerRequest, seconds, size, err := convertVideoCreateRequest(request, imageData)
	if err != nil {
		writeVideoError(c, http.StatusBadRequest, "invalid_request_error", "invalid_request", err.Error())
		return
	}
	body, err := json.Marshal(providerRequest)
	if err != nil || len(body) > siftq.MaxRequestBytes {
		writeVideoError(c, http.StatusRequestEntityTooLarge, "invalid_request_error", "request_too_large", "video request exceeds the provider limit")
		return
	}

	channel, channelMeta, err := selectSiftQChannel(c, request.Model)
	if err != nil {
		writeVideoError(c, http.StatusServiceUnavailable, "server_error", "provider_unavailable", err.Error())
		return
	}
	quota, modelRatio, groupRatio := siftQQuota(c, seconds)
	if err := model.PreConsumeTokenQuota(c.GetInt(ctxkey.TokenId), quota); err != nil {
		writeVideoError(c, http.StatusForbidden, "insufficient_quota", "insufficient_quota", err.Error())
		return
	}
	refund := true
	defer func() {
		if refund {
			refundSiftQQuota(c, quota)
		}
	}()

	target, err := siftq.BuildURL(channelMeta.BaseURL, siftq.CreateVideoPath)
	if err != nil {
		writeVideoError(c, http.StatusBadRequest, "invalid_request_error", "invalid_base_url", err.Error())
		return
	}
	status, responseBody, err := doSiftQUpstream(c, channelMeta, http.MethodPost, target, bytes.NewReader(body))
	if err != nil {
		writeVideoError(c, http.StatusBadGateway, "server_error", "upstream_error", "SiftQ request failed")
		return
	}
	if status < 200 || status >= 300 {
		writeVideoUpstreamError(c, status, responseBody)
		return
	}
	var response siftq.CreateResponse
	if err := json.Unmarshal(responseBody, &response); err != nil || siftq.ValidateCreateResponse(&response) != nil {
		writeVideoError(c, http.StatusBadGateway, "server_error", "invalid_upstream_response", "SiftQ response is missing task_id")
		return
	}
	binding := &model.SiftQTaskBinding{
		TaskID: response.TaskID, ChannelID: channel.Id, UserID: c.GetInt(ctxkey.Id), TokenID: c.GetInt(ctxkey.TokenId),
		Model: request.Model, Prompt: request.Prompt, Seconds: seconds, Size: size,
	}
	if err := model.SaveSiftQTaskBinding(binding); err != nil {
		logger.Error(c.Request.Context(), "failed to save SiftQ task binding: "+err.Error())
		writeVideoError(c, http.StatusBadGateway, "server_error", "task_binding_failed", "video task could not be persisted")
		return
	}
	refund = false
	finalizeSiftQQuota(c, channel.Id, quota, modelRatio, groupRatio)
	c.JSON(http.StatusOK, videoFromTask(siftq.VideoTask{ID: response.TaskID, Status: "queued"}, binding))
}

func RetrieveVideo(c *gin.Context) {
	task, binding, _, ok := loadVideoTask(c, c.Param("video_id"))
	if !ok {
		return
	}
	c.JSON(http.StatusOK, videoFromTask(task, binding))
}

func ListVideos(c *gin.Context) {
	limit := 20
	if raw := c.Query("limit"); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil || parsed < 1 || parsed > 100 {
			writeVideoError(c, http.StatusBadRequest, "invalid_request_error", "invalid_limit", "limit must be between 1 and 100")
			return
		}
		limit = parsed
	}
	order := c.DefaultQuery("order", "desc")
	if order != "asc" && order != "desc" {
		writeVideoError(c, http.StatusBadRequest, "invalid_request_error", "invalid_order", "order must be asc or desc")
		return
	}
	bindings, err := model.ListSiftQTaskBindings(c.GetInt(ctxkey.Id))
	if err != nil {
		writeVideoError(c, http.StatusInternalServerError, "server_error", "list_failed", "failed to load video tasks")
		return
	}
	objects := make([]videoObject, 0, len(bindings))
	for index := range bindings {
		task, _, _, ok := loadVideoTask(c, bindings[index].TaskID)
		if !ok {
			return
		}
		objects = append(objects, videoFromTask(task, &bindings[index]))
	}
	sort.Slice(objects, func(i, j int) bool {
		if order == "asc" {
			return objects[i].CreatedAt < objects[j].CreatedAt
		}
		return objects[i].CreatedAt > objects[j].CreatedAt
	})
	if after := c.Query("after"); after != "" {
		for index := range objects {
			if objects[index].ID == after {
				objects = objects[index+1:]
				break
			}
		}
	}
	if len(objects) > limit {
		objects = objects[:limit]
	}
	c.JSON(http.StatusOK, gin.H{"object": "list", "data": objects})
}

func DeleteVideo(c *gin.Context) {
	taskID := c.Param("video_id")
	_, channelMeta, err := selectSiftQTaskChannel(c, taskID)
	if err != nil {
		writeVideoError(c, http.StatusNotFound, "invalid_request_error", "video_not_found", "video was not found")
		return
	}
	target, err := siftq.BuildURL(channelMeta.BaseURL, siftq.CreateVideoPath+"/"+url.PathEscape(taskID))
	if err != nil {
		writeVideoError(c, http.StatusBadRequest, "invalid_request_error", "invalid_base_url", err.Error())
		return
	}
	status, body, err := doSiftQUpstream(c, channelMeta, http.MethodDelete, target, nil)
	if err != nil {
		writeVideoError(c, http.StatusBadGateway, "server_error", "upstream_error", "SiftQ request failed")
		return
	}
	if status < 200 || status >= 300 {
		writeVideoUpstreamError(c, status, body)
		return
	}
	var response siftq.DeleteResponse
	if err := json.Unmarshal(body, &response); err != nil || siftq.ValidateDeleteResponse(&response, taskID) != nil {
		writeVideoError(c, http.StatusBadGateway, "server_error", "invalid_upstream_response", "SiftQ returned a malformed delete response")
		return
	}
	if err := model.DeleteSiftQTaskBinding(c.GetInt(ctxkey.Id), taskID); err != nil {
		logger.Error(c.Request.Context(), "failed to delete SiftQ task binding: "+err.Error())
	}
	c.JSON(http.StatusOK, gin.H{"id": taskID, "object": "video.deleted", "deleted": true})
}

func RetrieveVideoContent(c *gin.Context) {
	task, _, _, ok := loadVideoTask(c, c.Param("video_id"))
	if !ok {
		return
	}
	if task.Status != "succeeded" || task.Content == nil || task.Content.URL == "" {
		writeVideoError(c, http.StatusConflict, "invalid_request_error", "video_not_ready", "video content is not ready")
		return
	}
	parsed, err := url.Parse(task.Content.URL)
	if err != nil || parsed.Host == "" || (parsed.Scheme != "http" && parsed.Scheme != "https") {
		writeVideoError(c, http.StatusBadGateway, "server_error", "invalid_content_url", "SiftQ returned an invalid video URL")
		return
	}
	request, err := http.NewRequestWithContext(c.Request.Context(), http.MethodGet, parsed.String(), nil)
	if err != nil {
		writeVideoError(c, http.StatusBadGateway, "server_error", "download_failed", "failed to create video download request")
		return
	}
	response, err := client.HTTPClient.Do(request)
	if err != nil || response == nil {
		writeVideoError(c, http.StatusBadGateway, "server_error", "download_failed", "failed to download video content")
		return
	}
	defer response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		writeVideoError(c, http.StatusBadGateway, "server_error", "download_failed", "video content URL returned an error")
		return
	}
	contentType := response.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "video/mp4"
	}
	c.DataFromReader(http.StatusOK, response.ContentLength, contentType, response.Body, map[string]string{
		"Content-Disposition": fmt.Sprintf("attachment; filename=\"%s.mp4\"", task.ID),
	})
}

func decodeVideoCreateRequest(c *gin.Context) (*videoCreateRequest, string, error) {
	request := &videoCreateRequest{}
	if strings.HasPrefix(c.ContentType(), "multipart/form-data") {
		if err := c.Request.ParseMultipartForm(siftq.MaxRequestBytes); err != nil {
			return nil, "", err
		}
		request.Prompt = c.PostForm("prompt")
		request.Model = c.PostForm("model")
		request.Seconds = c.PostForm("seconds")
		request.Size = c.PostForm("size")
		file, header, err := c.Request.FormFile("input_reference")
		if err == nil {
			defer file.Close()
			dataURI, err := multipartImageDataURI(file, header)
			if err != nil {
				return nil, "", err
			}
			return request, dataURI, nil
		}
		if !errors.Is(err, http.ErrMissingFile) {
			return nil, "", err
		}
		return request, c.PostForm("input_reference"), nil
	}
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, siftq.MaxRequestBytes+1)
	decoder := json.NewDecoder(c.Request.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(request); err != nil {
		return nil, "", err
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return nil, "", errors.New("request must contain exactly one JSON object")
	}
	if request.InputReference != nil {
		if request.InputReference.FileID != "" {
			return nil, "", errors.New("input_reference.file_id is not supported; use image_url or upload a file")
		}
		return request, request.InputReference.ImageURL, nil
	}
	return request, "", nil
}

func multipartImageDataURI(file multipart.File, header *multipart.FileHeader) (string, error) {
	data, err := io.ReadAll(io.LimitReader(file, (20<<20)+1))
	if err != nil {
		return "", err
	}
	if len(data) > 20<<20 {
		return "", errors.New("input_reference exceeds 20 MB")
	}
	mimeType := http.DetectContentType(data)
	if !strings.HasPrefix(mimeType, "image/") {
		return "", fmt.Errorf("input_reference %q is not an image", header.Filename)
	}
	return "data:" + mimeType + ";base64," + base64.StdEncoding.EncodeToString(data), nil
}

func convertVideoCreateRequest(request *videoCreateRequest, imageData string) (*siftq.CreateRequest, int, string, error) {
	request.Prompt = strings.TrimSpace(request.Prompt)
	if request.Prompt == "" || len([]rune(request.Prompt)) > 7000 {
		return nil, 0, "", errors.New("prompt must contain between 1 and 7000 characters")
	}
	if request.Model == "" {
		request.Model = siftq.ModelName
	}
	if request.Model != siftq.ModelName {
		return nil, 0, "", fmt.Errorf("model must be %s", siftq.ModelName)
	}
	seconds, err := parseVideoSeconds(request.Seconds)
	if err != nil {
		return nil, 0, "", err
	}
	resolution, ratio, size, err := mapVideoSize(request.Size)
	if err != nil {
		return nil, 0, "", err
	}
	content := []siftq.ContentItem{{Type: "text", Text: request.Prompt}}
	if imageData != "" {
		content = append(content, siftq.ContentItem{Type: "image_url", ImageURL: &siftq.URLValue{URL: imageData}, Role: "first_frame"})
		ratio = "adaptive"
	}
	providerRequest := &siftq.CreateRequest{Model: siftq.ModelName, Content: content, Resolution: &resolution, Duration: &seconds, Ratio: &ratio}
	if err := siftq.ValidateCreateRequest(providerRequest, false); err != nil {
		return nil, 0, "", err
	}
	return providerRequest, seconds, size, nil
}

func parseVideoSeconds(value any) (int, error) {
	if value == nil || value == "" {
		return 4, nil
	}
	var raw string
	switch typed := value.(type) {
	case string:
		raw = typed
	case float64:
		if typed != math.Trunc(typed) {
			return 0, errors.New("seconds must be an integer from 4 through 15")
		}
		raw = strconv.Itoa(int(typed))
	default:
		return 0, errors.New("seconds must be a string or integer")
	}
	seconds, err := strconv.Atoi(raw)
	if err != nil || seconds < 4 || seconds > 15 {
		return 0, errors.New("seconds must be an integer from 4 through 15")
	}
	return seconds, nil
}

func mapVideoSize(size string) (string, string, string, error) {
	if size == "" {
		size = "720x1280"
	}
	switch size {
	case "720x1280":
		return "768P", "9:16", size, nil
	case "1280x720":
		return "768P", "16:9", size, nil
	case "1024x1792":
		return "2K", "9:16", size, nil
	case "1792x1024":
		return "2K", "16:9", size, nil
	default:
		return "", "", "", errors.New("size must be 720x1280, 1280x720, 1024x1792, or 1792x1024")
	}
}

func loadVideoTask(c *gin.Context, taskID string) (siftq.VideoTask, *model.SiftQTaskBinding, *meta.Meta, bool) {
	binding, channelMeta, err := selectSiftQTaskChannel(c, taskID)
	if err != nil {
		writeVideoError(c, http.StatusNotFound, "invalid_request_error", "video_not_found", "video was not found")
		return siftq.VideoTask{}, nil, nil, false
	}
	target, err := siftq.BuildURL(channelMeta.BaseURL, siftq.QueryVideoPath+"/"+url.PathEscape(taskID))
	if err != nil {
		writeVideoError(c, http.StatusBadRequest, "invalid_request_error", "invalid_base_url", err.Error())
		return siftq.VideoTask{}, nil, nil, false
	}
	status, body, err := doSiftQUpstream(c, channelMeta, http.MethodGet, target, nil)
	if err != nil {
		writeVideoError(c, http.StatusBadGateway, "server_error", "upstream_error", "SiftQ request failed")
		return siftq.VideoTask{}, nil, nil, false
	}
	if status < 200 || status >= 300 {
		writeVideoUpstreamError(c, status, body)
		return siftq.VideoTask{}, nil, nil, false
	}
	var response siftq.QueryResponse
	if err := json.Unmarshal(body, &response); err != nil || siftq.ValidateTask(&response.Task) != nil || response.Task.ID != taskID {
		writeVideoError(c, http.StatusBadGateway, "server_error", "invalid_upstream_response", "SiftQ returned a malformed task response")
		return siftq.VideoTask{}, nil, nil, false
	}
	return response.Task, binding, channelMeta, true
}

func videoFromTask(task siftq.VideoTask, binding *model.SiftQTaskBinding) videoObject {
	status, progress := "queued", 0
	switch task.Status {
	case "running":
		status, progress = "in_progress", 50
	case "succeeded":
		status, progress = "completed", 100
	case "failed", "cancelled":
		status, progress = "failed", 100
	}
	createdAt := binding.CreatedAt
	if task.CreatedAt != 0 {
		createdAt = task.CreatedAt
	}
	result := videoObject{ID: task.ID, Object: "video", Model: binding.Model, Status: status, Progress: progress, CreatedAt: createdAt, Prompt: binding.Prompt, Seconds: strconv.Itoa(binding.Seconds), Size: binding.Size}
	if result.Model == "" {
		result.Model = siftq.ModelName
	}
	if task.Duration > 0 {
		result.Seconds = strconv.Itoa(task.Duration)
	}
	if status == "completed" {
		completed := task.UpdatedAt
		if completed == 0 {
			completed = createdAt
		}
		result.CompletedAt = &completed
	}
	if status == "failed" {
		result.Error = &videoError{Code: "generation_failed", Message: "video generation failed"}
		if task.Error != nil {
			result.Error.Code, result.Error.Message = task.Error.Code, task.Error.Message
		}
	}
	return result
}

func selectSiftQChannel(c *gin.Context, modelName string) (*model.Channel, *meta.Meta, error) {
	if modelName == "" {
		modelName = siftq.ModelName
	}
	userID := c.GetInt(ctxkey.Id)
	group, err := model.CacheGetUserGroup(userID)
	if err != nil {
		return nil, nil, err
	}
	c.Set(ctxkey.Group, group)
	var channel *model.Channel
	if value, exists := c.Get(ctxkey.SpecificChannelId); exists {
		channelID, err := strconv.Atoi(value.(string))
		if err != nil {
			return nil, nil, errors.New("specific channel id is invalid")
		}
		channel, err = model.GetChannelById(channelID, true)
		if err != nil {
			return nil, nil, errors.New("specific channel was not found")
		}
	} else {
		channel, err = model.CacheGetRandomSatisfiedChannel(group, modelName, false)
		if err != nil {
			return nil, nil, fmt.Errorf("no SiftQ channel is available for %s", modelName)
		}
	}
	if channel.Type != channeltype.SiftQ || channel.Status != model.ChannelStatusEnabled {
		return nil, nil, errors.New("selected channel is not an enabled SiftQ channel")
	}
	middleware.SetupContextForSelectedChannel(c, channel, modelName)
	return channel, meta.GetByContext(c), nil
}

func selectSiftQTaskChannel(c *gin.Context, taskID string) (*model.SiftQTaskBinding, *meta.Meta, error) {
	if strings.TrimSpace(taskID) == "" || len(taskID) > 128 {
		return nil, nil, errors.New("invalid task id")
	}
	binding, err := model.GetSiftQTaskBinding(c.GetInt(ctxkey.Id), taskID)
	if err != nil {
		return nil, nil, err
	}
	channel, err := model.GetChannelById(binding.ChannelID, true)
	if err != nil || channel.Type != channeltype.SiftQ || channel.Status != model.ChannelStatusEnabled {
		return nil, nil, errors.New("bound SiftQ channel is unavailable")
	}
	group, err := model.CacheGetUserGroup(c.GetInt(ctxkey.Id))
	if err != nil {
		return nil, nil, err
	}
	c.Set(ctxkey.Group, group)
	middleware.SetupContextForSelectedChannel(c, channel, siftq.ModelName)
	return binding, meta.GetByContext(c), nil
}

func doSiftQUpstream(c *gin.Context, channelMeta *meta.Meta, method, target string, body io.Reader) (int, []byte, error) {
	request, err := siftq.NewUpstreamRequest(c, channelMeta, method, target, body)
	if err != nil {
		return 0, nil, err
	}
	response, err := client.HTTPClient.Do(request)
	if err != nil {
		return 0, nil, err
	}
	if response == nil {
		return 0, nil, errors.New("SiftQ response is nil")
	}
	defer response.Body.Close()
	responseBody, err := io.ReadAll(io.LimitReader(response.Body, siftQResponseLimit+1))
	if err != nil {
		return response.StatusCode, nil, err
	}
	if len(responseBody) > siftQResponseLimit {
		return response.StatusCode, nil, errors.New("SiftQ response exceeds size limit")
	}
	return response.StatusCode, responseBody, nil
}

func writeVideoUpstreamError(c *gin.Context, status int, body []byte) {
	var envelope siftq.ErrorEnvelope
	if err := json.Unmarshal(body, &envelope); err == nil && envelope.Error.Message != "" {
		writeVideoError(c, status, videoErrorType(status), envelope.Error.Type, envelope.Error.Message)
		return
	}
	writeVideoError(c, status, videoErrorType(status), "upstream_error", fmt.Sprintf("SiftQ returned HTTP %d", status))
}

func writeVideoError(c *gin.Context, status int, errorType, code, message string) {
	c.JSON(status, gin.H{"error": gin.H{"message": helper.MessageWithRequestId(message, c.GetString(helper.RequestIdKey)), "type": errorType, "code": code}})
}

func videoErrorType(status int) string {
	if status == http.StatusTooManyRequests {
		return "rate_limit_error"
	}
	if status >= 400 && status < 500 {
		return "invalid_request_error"
	}
	return "server_error"
}

func siftQQuota(c *gin.Context, duration int) (int64, float64, float64) {
	modelRatio := billingratio.GetModelRatio(siftq.ModelName, channeltype.SiftQ)
	groupRatio := billingratio.GetGroupRatio(c.GetString(ctxkey.Group))
	quota := int64(math.Ceil(float64(duration) * modelRatio * groupRatio * 1000))
	return quota, modelRatio, groupRatio
}

func refundSiftQQuota(c *gin.Context, quota int64) {
	if quota == 0 {
		return
	}
	if err := model.PostConsumeTokenQuota(c.GetInt(ctxkey.TokenId), -quota); err != nil {
		logger.Error(c.Request.Context(), "failed to refund SiftQ quota: "+err.Error())
	}
	_ = model.CacheUpdateUserQuota(c.Request.Context(), c.GetInt(ctxkey.Id))
}

func finalizeSiftQQuota(c *gin.Context, channelID int, quota int64, modelRatio, groupRatio float64) {
	_ = model.CacheUpdateUserQuota(c.Request.Context(), c.GetInt(ctxkey.Id))
	if quota == 0 {
		return
	}
	model.RecordConsumeLog(c.Request.Context(), &model.Log{
		UserId: c.GetInt(ctxkey.Id), ChannelId: channelID, ModelName: siftq.ModelName,
		TokenName: c.GetString(ctxkey.TokenName), Quota: int(quota),
		Content: fmt.Sprintf("视频时长倍率：%.2f × %.2f", modelRatio, groupRatio),
	})
	model.UpdateUserUsedQuotaAndRequestCount(c.GetInt(ctxkey.Id), quota)
	model.UpdateChannelUsedQuota(channelID, quota)
}
