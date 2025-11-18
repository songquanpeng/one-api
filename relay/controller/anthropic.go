package controller

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/songquanpeng/one-api/common"
	"github.com/songquanpeng/one-api/common/config"
	"github.com/songquanpeng/one-api/common/logger"
	dbmodel "github.com/songquanpeng/one-api/model"
	"github.com/songquanpeng/one-api/relay"
	"github.com/songquanpeng/one-api/relay/adaptor/anthropic"
	"github.com/songquanpeng/one-api/relay/adaptor/openai"
	"github.com/songquanpeng/one-api/relay/billing"
	billingratio "github.com/songquanpeng/one-api/relay/billing/ratio"
	"github.com/songquanpeng/one-api/relay/meta"
	"github.com/songquanpeng/one-api/relay/model"
)

// RelayAnthropicHelper handles native Anthropic API requests (anthropic -> anthropic passthrough)
func RelayAnthropicHelper(c *gin.Context) *model.ErrorWithStatusCode {
	ctx := c.Request.Context()
	meta := meta.GetByContext(c)

	logger.Infof(ctx, "Anthropic request received - URL: %s", c.Request.URL.String())

	// get & validate anthropic request
	anthropicRequest, err := getAndValidateAnthropicRequest(c)
	if err != nil {
		logger.Errorf(ctx, "getAndValidateAnthropicRequest failed: %s", err.Error())
		return openai.ErrorWrapper(err, "invalid_anthropic_request", http.StatusBadRequest)
	}
	logger.Debugf(ctx, "Parsed anthropic request - Model: %s, Stream: %v, Messages: %d",
		anthropicRequest.Model, anthropicRequest.Stream, len(anthropicRequest.Messages))
	meta.IsStream = anthropicRequest.Stream

	// map model name
	meta.OriginModelName = anthropicRequest.Model
	mappedModel, _ := getMappedModelName(anthropicRequest.Model, meta.ModelMapping)
	anthropicRequest.Model = mappedModel
	meta.ActualModelName = anthropicRequest.Model

	// estimate token usage for anthropic request
	promptTokens := estimateAnthropicTokens(anthropicRequest)
	meta.PromptTokens = promptTokens

	// get model ratio & group ratio
	modelRatio := billingratio.GetModelRatio(anthropicRequest.Model, meta.ChannelType)
	groupRatio := billingratio.GetGroupRatio(meta.Group)
	ratio := modelRatio * groupRatio

	// pre-consume quota
	preConsumedQuota, bizErr := preConsumeQuotaForAnthropic(ctx, anthropicRequest, promptTokens, ratio, meta)
	if bizErr != nil {
		logger.Warnf(ctx, "preConsumeQuota failed: %+v", *bizErr)
		return bizErr
	}

	logger.Debugf(ctx, "Meta info - APIType: %d, ChannelType: %d, BaseURL: %s", meta.APIType, meta.ChannelType, meta.BaseURL)

	adaptor := relay.GetAdaptor(meta.APIType)
	if adaptor == nil {
		logger.Errorf(ctx, "Failed to get adaptor for API type: %d", meta.APIType)
		return openai.ErrorWrapper(fmt.Errorf("invalid api type: %d", meta.APIType), "invalid_api_type", http.StatusBadRequest)
	}
	logger.Debugf(ctx, "Using adaptor: %s", adaptor.GetChannelName())
	adaptor.Init(meta)

	// get request body - for anthropic passthrough, we directly use the request body
	requestBody, err := getAnthropicRequestBody(c, anthropicRequest)
	if err != nil {
		return openai.ErrorWrapper(err, "convert_anthropic_request_failed", http.StatusInternalServerError)
	}

	// do request
	resp, err := adaptor.DoRequest(c, meta, requestBody)
	if err != nil {
		logger.Errorf(ctx, "DoRequest failed: %s", err.Error())
		return openai.ErrorWrapper(err, "do_request_failed", http.StatusInternalServerError)
	}
	logger.Debugf(ctx, "Received response - Status: %d", resp.StatusCode)

	if isErrorHappened(meta, resp) {
		logger.Errorf(ctx, "Error detected in response - Status: %d", resp.StatusCode)

		// Read response body for detailed error logging
		if resp.Body != nil {
			bodyBytes, readErr := io.ReadAll(resp.Body)
			if readErr == nil {
				logger.Errorf(ctx, "Error response body: %s", string(bodyBytes))
				// Recreate the body for further processing
				resp.Body = io.NopCloser(bytes.NewReader(bodyBytes))
			}
		}

		billing.ReturnPreConsumedQuota(ctx, preConsumedQuota, meta.TokenId)
		return RelayErrorHandler(resp)
	}

	// do response - for anthropic native requests, we need to handle the response directly
	usage, respErr := handleAnthropicResponse(c, resp, meta)
	if respErr != nil {
		logger.Errorf(ctx, "respErr is not nil: %+v", respErr)
		billing.ReturnPreConsumedQuota(ctx, preConsumedQuota, meta.TokenId)
		return respErr
	}

	logger.Infof(ctx, "Anthropic request completed - Usage: %+v", usage)

	// post-consume quota - for anthropic, we create a placeholder GeneralOpenAIRequest
	placeholderRequest := &model.GeneralOpenAIRequest{
		Model: anthropicRequest.Model,
	}
	go postConsumeQuota(ctx, usage, meta, placeholderRequest, ratio, preConsumedQuota, modelRatio, groupRatio, false)
	return nil
}

func getAndValidateAnthropicRequest(c *gin.Context) (*anthropic.Request, error) {
	ctx := c.Request.Context()
	anthropicRequest := &anthropic.Request{}
	err := common.UnmarshalBodyReusable(c, anthropicRequest)
	if err != nil {
		logger.Errorf(ctx, "Failed to unmarshal Anthropic request: %s", err.Error())
		return nil, err
	}

	logger.Debugf(ctx, "Successfully parsed Anthropic request with %d messages", len(anthropicRequest.Messages))

	// Basic validation only - minimal intervention
	if anthropicRequest.Model == "" {
		return nil, fmt.Errorf("model is required")
	}
	if len(anthropicRequest.Messages) == 0 {
		return nil, fmt.Errorf("messages are required")
	}
	if anthropicRequest.MaxTokens == 0 {
		anthropicRequest.MaxTokens = 4096 // default max tokens
	}

	// Add detailed logging for debugging without filtering messages
	logger.Debugf(ctx, "Request details - Model: %s, MaxTokens: %d, Messages: %d",
		anthropicRequest.Model, anthropicRequest.MaxTokens, len(anthropicRequest.Messages))

	for i, message := range anthropicRequest.Messages {
		contentArray := message.Content.ToContentArray()
		logger.Debugf(ctx, "Message[%d] - Role: %s, Content blocks: %d", i, message.Role, len(contentArray))

		for j, content := range contentArray {
			if content.Type == "tool_use" {
				logger.Debugf(ctx, "  Content[%d] - Type: %s, ID: %s, Name: %s", j, content.Type, content.Id, content.Name)
			} else if content.Type == "tool_result" {
				logger.Debugf(ctx, "  Content[%d] - Type: %s, ToolUseId: %s, Content length: %d",
					j, content.Type, content.ToolUseId, len(content.Content))
			} else {
				logger.Debugf(ctx, "  Content[%d] - Type: %s, Text length: %d", j, content.Type, len(content.Text))
			}
		}
	}

	return anthropicRequest, nil
}

func getAnthropicRequestBody(c *gin.Context, anthropicRequest *anthropic.Request) (io.Reader, error) {
	// For anthropic native requests, we marshal the request back to JSON
	jsonData, err := json.Marshal(anthropicRequest)
	if err != nil {
		logger.Errorf(c.Request.Context(), "Failed to marshal anthropic request: %s", err.Error())
		return nil, err
	}

	ctx := c.Request.Context()
	logger.Debugf(ctx, "Final request body size: %d bytes", len(jsonData))
	logger.Debugf(ctx, "Final request body: %s", string(jsonData))

	return bytes.NewBuffer(jsonData), nil
}

const (
	// CHARS_PER_TOKEN represents the rough character-to-token ratio for Anthropic models
	// This is a conservative estimate: approximately 1 token per 4 characters
	CHARS_PER_TOKEN = 4
)

func estimateAnthropicTokens(request *anthropic.Request) int {
	// Simple token estimation for Anthropic requests
	// This is a rough estimation, actual implementation might need more sophisticated logic
	totalTokens := 0

	// Count tokens in system prompt
	if !request.System.IsEmpty() {
		systemText := request.System.String()
		totalTokens += len(systemText) / CHARS_PER_TOKEN // rough estimate: 1 token per 4 characters
	}

	// Count tokens in messages
	for _, message := range request.Messages {
		for _, content := range message.Content.ToContentArray() {
			if content.Type == "text" {
				totalTokens += len(content.Text) / CHARS_PER_TOKEN
			}
		}
	}

	return totalTokens
}

func handleAnthropicResponse(c *gin.Context, resp *http.Response, meta *meta.Meta) (*model.Usage, *model.ErrorWithStatusCode) {
	// For anthropic native requests, use direct handlers to maintain Anthropic format
	if meta.IsStream {
		// Handle streaming response - note: DirectStreamHandler returns (error, usage)
		err, usage := anthropic.DirectStreamHandler(c, resp)
		return usage, err
	} else {
		// Handle non-streaming response - note: DirectHandler returns (error, usage)
		err, usage := anthropic.DirectHandler(c, resp, meta.PromptTokens, meta.ActualModelName)
		return usage, err
	}
}

func preConsumeQuotaForAnthropic(ctx context.Context, request *anthropic.Request, promptTokens int, ratio float64, meta *meta.Meta) (int64, *model.ErrorWithStatusCode) {
	// Use the same quota logic as text requests but adapted for Anthropic
	preConsumedTokens := config.PreConsumedQuota + int64(promptTokens)
	if request.MaxTokens != 0 {
		preConsumedTokens += int64(request.MaxTokens)
	}
	preConsumedQuota := int64(float64(preConsumedTokens) * ratio)

	userQuota, err := dbmodel.CacheGetUserQuota(ctx, meta.UserId)
	if err != nil {
		return preConsumedQuota, openai.ErrorWrapper(err, "get_user_quota_failed", http.StatusInternalServerError)
	}

	if userQuota-preConsumedQuota < 0 {
		return preConsumedQuota, openai.ErrorWrapper(fmt.Errorf("user quota is not enough"), "insufficient_user_quota", http.StatusForbidden)
	}

	err = dbmodel.CacheDecreaseUserQuota(meta.UserId, preConsumedQuota)
	if err != nil {
		return preConsumedQuota, openai.ErrorWrapper(err, "decrease_user_quota_failed", http.StatusInternalServerError)
	}

	if userQuota > 100*preConsumedQuota {
		// in this case, we do not pre-consume quota
		// because the user has enough quota
		preConsumedQuota = 0
		logger.Info(ctx, fmt.Sprintf("user %d has enough quota %d, trusted and no need to pre-consume", meta.UserId, userQuota))
	}

	if preConsumedQuota > 0 {
		err := dbmodel.PreConsumeTokenQuota(meta.TokenId, preConsumedQuota)
		if err != nil {
			return preConsumedQuota, openai.ErrorWrapper(err, "pre_consume_token_quota_failed", http.StatusForbidden)
		}
	}

	return preConsumedQuota, nil
}
