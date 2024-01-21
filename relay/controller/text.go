package controller

import (
	"context"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"math"
	"net/http"
	"one-api/common"
	"one-api/common/logger"
	"one-api/model"
	"one-api/relay/channel/openai"
	"one-api/relay/constant"
	"one-api/relay/util"
	"strings"
)

func RelayTextHelper(c *gin.Context, relayMode int) *openai.ErrorWithStatusCode {
	ctx := c.Request.Context()
	meta := util.GetRelayMeta(c)
	var textRequest openai.GeneralOpenAIRequest
	err := common.UnmarshalBodyReusable(c, &textRequest)
	if err != nil {
		return openai.ErrorWrapper(err, "bind_request_body_failed", http.StatusBadRequest)
	}
	if relayMode == constant.RelayModeModerations && textRequest.Model == "" {
		textRequest.Model = "text-moderation-latest"
	}
	if relayMode == constant.RelayModeEmbeddings && textRequest.Model == "" {
		textRequest.Model = c.Param("model")
	}
	err = util.ValidateTextRequest(&textRequest, relayMode)
	if err != nil {
		return openai.ErrorWrapper(err, "invalid_text_request", http.StatusBadRequest)
	}
	var isModelMapped bool
	textRequest.Model, isModelMapped = util.GetMappedModelName(textRequest.Model, meta.ModelMapping)
	apiType := constant.ChannelType2APIType(meta.ChannelType)
	fullRequestURL, err := GetRequestURL(c.Request.URL.String(), apiType, relayMode, meta, &textRequest)
	if err != nil {
		logger.Error(ctx, fmt.Sprintf("util.GetRequestURL failed: %s", err.Error()))
		return openai.ErrorWrapper(fmt.Errorf("util.GetRequestURL failed"), "get_request_url_failed", http.StatusInternalServerError)
	}
	var promptTokens int
	var completionTokens int
	switch relayMode {
	case constant.RelayModeChatCompletions:
		promptTokens = openai.CountTokenMessages(textRequest.Messages, textRequest.Model)
	case constant.RelayModeCompletions:
		promptTokens = openai.CountTokenInput(textRequest.Prompt, textRequest.Model)
	case constant.RelayModeModerations:
		promptTokens = openai.CountTokenInput(textRequest.Input, textRequest.Model)
	}
	preConsumedTokens := common.PreConsumedQuota
	if textRequest.MaxTokens != 0 {
		preConsumedTokens = promptTokens + textRequest.MaxTokens
	}
	modelRatio := common.GetModelRatio(textRequest.Model)
	groupRatio := common.GetGroupRatio(meta.Group)
	ratio := modelRatio * groupRatio
	preConsumedQuota := int(float64(preConsumedTokens) * ratio)
	userQuota, err := model.CacheGetUserQuota(meta.UserId)
	if err != nil {
		return openai.ErrorWrapper(err, "get_user_quota_failed", http.StatusInternalServerError)
	}
	if userQuota-preConsumedQuota < 0 {
		return openai.ErrorWrapper(errors.New("user quota is not enough"), "insufficient_user_quota", http.StatusForbidden)
	}
	err = model.CacheDecreaseUserQuota(meta.UserId, preConsumedQuota)
	if err != nil {
		return openai.ErrorWrapper(err, "decrease_user_quota_failed", http.StatusInternalServerError)
	}
	if userQuota > 100*preConsumedQuota {
		// in this case, we do not pre-consume quota
		// because the user has enough quota
		preConsumedQuota = 0
		logger.Info(c.Request.Context(), fmt.Sprintf("user %d has enough quota %d, trusted and no need to pre-consume", meta.UserId, userQuota))
	}
	if preConsumedQuota > 0 {
		err := model.PreConsumeTokenQuota(meta.TokenId, preConsumedQuota)
		if err != nil {
			return openai.ErrorWrapper(err, "pre_consume_token_quota_failed", http.StatusForbidden)
		}
	}
	requestBody, err := GetRequestBody(c, textRequest, isModelMapped, apiType, relayMode)
	if err != nil {
		return openai.ErrorWrapper(err, "get_request_body_failed", http.StatusInternalServerError)
	}
	var req *http.Request
	var resp *http.Response
	isStream := textRequest.Stream

	if apiType != constant.APITypeXunfei { // cause xunfei use websocket
		req, err = http.NewRequest(c.Request.Method, fullRequestURL, requestBody)
		if err != nil {
			return openai.ErrorWrapper(err, "new_request_failed", http.StatusInternalServerError)
		}
		SetupRequestHeaders(c, req, apiType, meta, isStream)
		resp, err = util.HTTPClient.Do(req)
		if err != nil {
			return openai.ErrorWrapper(err, "do_request_failed", http.StatusInternalServerError)
		}
		err = req.Body.Close()
		if err != nil {
			return openai.ErrorWrapper(err, "close_request_body_failed", http.StatusInternalServerError)
		}
		err = c.Request.Body.Close()
		if err != nil {
			return openai.ErrorWrapper(err, "close_request_body_failed", http.StatusInternalServerError)
		}
		isStream = isStream || strings.HasPrefix(resp.Header.Get("Content-Type"), "text/event-stream")

		if resp.StatusCode != http.StatusOK {
			util.ReturnPreConsumedQuota(ctx, preConsumedQuota, meta.TokenId)
			return util.RelayErrorHandler(resp)
		}
	}

	var respErr *openai.ErrorWithStatusCode
	var usage *openai.Usage

	defer func(ctx context.Context) {
		// Why we use defer here? Because if error happened, we will have to return the pre-consumed quota.
		if respErr != nil {
			logger.Errorf(ctx, "respErr is not nil: %+v", respErr)
			util.ReturnPreConsumedQuota(ctx, preConsumedQuota, meta.TokenId)
			return
		}
		if usage == nil {
			logger.Error(ctx, "usage is nil, which is unexpected")
			return
		}

		go func() {
			quota := 0
			completionRatio := common.GetCompletionRatio(textRequest.Model)
			promptTokens = usage.PromptTokens
			completionTokens = usage.CompletionTokens
			quota = int(math.Ceil((float64(promptTokens) + float64(completionTokens)*completionRatio) * ratio))
			if ratio != 0 && quota <= 0 {
				quota = 1
			}
			totalTokens := promptTokens + completionTokens
			if totalTokens == 0 {
				// in this case, must be some error happened
				// we cannot just return, because we may have to return the pre-consumed quota
				quota = 0
			}
			quotaDelta := quota - preConsumedQuota
			err := model.PostConsumeTokenQuota(meta.TokenId, quotaDelta)
			if err != nil {
				logger.Error(ctx, "error consuming token remain quota: "+err.Error())
			}
			err = model.CacheUpdateUserQuota(meta.UserId)
			if err != nil {
				logger.Error(ctx, "error update user quota cache: "+err.Error())
			}
			if quota != 0 {
				logContent := fmt.Sprintf("模型倍率 %.2f，分组倍率 %.2f", modelRatio, groupRatio)
				model.RecordConsumeLog(ctx, meta.UserId, meta.ChannelId, promptTokens, completionTokens, textRequest.Model, meta.TokenName, quota, logContent)
				model.UpdateUserUsedQuotaAndRequestCount(meta.UserId, quota)
				model.UpdateChannelUsedQuota(meta.ChannelId, quota)
			}
		}()
	}(ctx)
	usage, respErr = DoResponse(c, &textRequest, resp, relayMode, apiType, isStream, promptTokens)
	if respErr != nil {
		return respErr
	}
	return nil
}
