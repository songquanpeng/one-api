package controller

import (
	"context"
	"errors"
	"net/http"
	"one-api/common"
	"one-api/model"
	"one-api/providers"
	providers_base "one-api/providers/base"
	"one-api/types"

	"github.com/gin-gonic/gin"
)

func relayTextHelper(c *gin.Context, relayMode int) *types.OpenAIErrorWithStatusCode {
	// 获取请求参数
	channelType := c.GetInt("channel")
	channelId := c.GetInt("channel_id")
	tokenId := c.GetInt("token_id")
	userId := c.GetInt("id")
	group := c.GetString("group")

	// 获取 Provider
	provider := providers.GetProvider(channelType, c)
	if provider == nil {
		return types.ErrorWrapper(errors.New("channel not implemented"), "channel_not_implemented", http.StatusNotImplemented)
	}

	modelMap, err := parseModelMapping(c.GetString("model_mapping"))
	if err != nil {
		return types.ErrorWrapper(err, "unmarshal_model_mapping_failed", http.StatusInternalServerError)
	}

	var promptTokens int
	quotaInfo := &QuotaInfo{
		modelName:    "",
		promptTokens: promptTokens,
		userId:       userId,
		channelId:    channelId,
		tokenId:      tokenId,
	}

	var usage *types.Usage
	var openAIErrorWithStatusCode *types.OpenAIErrorWithStatusCode

	switch relayMode {
	case RelayModeChatCompletions:
		usage, openAIErrorWithStatusCode = handleChatCompletions(c, provider, modelMap, quotaInfo, group)
	case RelayModeCompletions:
		usage, openAIErrorWithStatusCode = handleCompletions(c, provider, modelMap, quotaInfo, group)
	case RelayModeEmbeddings:
		usage, openAIErrorWithStatusCode = handleEmbeddings(c, provider, modelMap, quotaInfo, group)
	default:
		return types.ErrorWrapper(errors.New("invalid relay mode"), "invalid_relay_mode", http.StatusBadRequest)
	}

	if openAIErrorWithStatusCode != nil {
		if quotaInfo.preConsumedQuota != 0 {
			go func(ctx context.Context) {
				// return pre-consumed quota
				err := model.PostConsumeTokenQuota(tokenId, -quotaInfo.preConsumedQuota)
				if err != nil {
					common.LogError(ctx, "error return pre-consumed quota: "+err.Error())
				}
			}(c.Request.Context())
		}
		return openAIErrorWithStatusCode
	}

	tokenName := c.GetString("token_name")
	defer func(ctx context.Context) {
		go func() {
			err = quotaInfo.completedQuotaConsumption(usage, tokenName, ctx)
			if err != nil {
				common.LogError(ctx, err.Error())
			}
		}()
	}(c.Request.Context())

	return nil
}

func handleChatCompletions(c *gin.Context, provider providers_base.ProviderInterface, modelMap map[string]string, quotaInfo *QuotaInfo, group string) (*types.Usage, *types.OpenAIErrorWithStatusCode) {
	var chatRequest types.ChatCompletionRequest
	isModelMapped := false
	chatProvider, ok := provider.(providers_base.ChatInterface)
	if !ok {
		return nil, types.ErrorWrapper(errors.New("channel not implemented"), "channel_not_implemented", http.StatusNotImplemented)
	}
	err := common.UnmarshalBodyReusable(c, &chatRequest)
	if err != nil {
		return nil, types.ErrorWrapper(err, "bind_request_body_failed", http.StatusBadRequest)
	}
	if modelMap != nil && modelMap[chatRequest.Model] != "" {
		chatRequest.Model = modelMap[chatRequest.Model]
		isModelMapped = true
	}
	promptTokens := common.CountTokenMessages(chatRequest.Messages, chatRequest.Model)

	quotaInfo.modelName = chatRequest.Model
	quotaInfo.initQuotaInfo(group)
	quota_err := quotaInfo.preQuotaConsumption()
	if quota_err != nil {
		return nil, quota_err
	}
	return chatProvider.ChatAction(&chatRequest, isModelMapped, promptTokens)
}

func handleCompletions(c *gin.Context, provider providers_base.ProviderInterface, modelMap map[string]string, quotaInfo *QuotaInfo, group string) (*types.Usage, *types.OpenAIErrorWithStatusCode) {
	var completionRequest types.CompletionRequest
	isModelMapped := false
	completionProvider, ok := provider.(providers_base.CompletionInterface)
	if !ok {
		return nil, types.ErrorWrapper(errors.New("channel not implemented"), "channel_not_implemented", http.StatusNotImplemented)
	}
	err := common.UnmarshalBodyReusable(c, &completionRequest)
	if err != nil {
		return nil, types.ErrorWrapper(err, "bind_request_body_failed", http.StatusBadRequest)
	}
	if modelMap != nil && modelMap[completionRequest.Model] != "" {
		completionRequest.Model = modelMap[completionRequest.Model]
		isModelMapped = true
	}
	promptTokens := common.CountTokenInput(completionRequest.Prompt, completionRequest.Model)

	quotaInfo.modelName = completionRequest.Model
	quotaInfo.initQuotaInfo(group)
	quota_err := quotaInfo.preQuotaConsumption()
	if quota_err != nil {
		return nil, quota_err
	}
	return completionProvider.CompleteAction(&completionRequest, isModelMapped, promptTokens)
}

func handleEmbeddings(c *gin.Context, provider providers_base.ProviderInterface, modelMap map[string]string, quotaInfo *QuotaInfo, group string) (*types.Usage, *types.OpenAIErrorWithStatusCode) {
	var embeddingsRequest types.EmbeddingRequest
	isModelMapped := false
	embeddingsProvider, ok := provider.(providers_base.EmbeddingsInterface)
	if !ok {
		return nil, types.ErrorWrapper(errors.New("channel not implemented"), "channel_not_implemented", http.StatusNotImplemented)
	}
	err := common.UnmarshalBodyReusable(c, &embeddingsRequest)
	if err != nil {
		return nil, types.ErrorWrapper(err, "bind_request_body_failed", http.StatusBadRequest)
	}
	if modelMap != nil && modelMap[embeddingsRequest.Model] != "" {
		embeddingsRequest.Model = modelMap[embeddingsRequest.Model]
		isModelMapped = true
	}
	promptTokens := common.CountTokenInput(embeddingsRequest.Input, embeddingsRequest.Model)

	quotaInfo.modelName = embeddingsRequest.Model
	quotaInfo.initQuotaInfo(group)
	quota_err := quotaInfo.preQuotaConsumption()
	if quota_err != nil {
		return nil, quota_err
	}
	return embeddingsProvider.EmbeddingsAction(&embeddingsRequest, isModelMapped, promptTokens)
}
