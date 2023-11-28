package controller

import (
	"context"
	"errors"
	"net/http"
	"one-api/common"
	"one-api/model"
	"one-api/providers"
	"one-api/types"

	"github.com/gin-gonic/gin"
)

func relayCompletionHelper(c *gin.Context) *types.OpenAIErrorWithStatusCode {

	// 获取请求参数
	channelType := c.GetInt("channel")
	channelId := c.GetInt("channel_id")
	tokenId := c.GetInt("token_id")
	userId := c.GetInt("id")
	// consumeQuota := c.GetBool("consume_quota")
	group := c.GetString("group")

	// 获取 Provider
	completionProvider := GetCompletionProvider(channelType, c)
	if completionProvider == nil {
		return types.ErrorWrapper(errors.New("API not implemented"), "api_not_implemented", http.StatusNotImplemented)
	}

	// 获取请求体
	var completionRequest types.CompletionRequest
	err := common.UnmarshalBodyReusable(c, &completionRequest)
	if err != nil {
		return types.ErrorWrapper(err, "bind_request_body_failed", http.StatusBadRequest)
	}

	// 检查模型映射
	isModelMapped := false
	modelMap, err := parseModelMapping(c.GetString("model_mapping"))
	if err != nil {
		return types.ErrorWrapper(err, "unmarshal_model_mapping_failed", http.StatusInternalServerError)
	}
	if modelMap != nil && modelMap[completionRequest.Model] != "" {
		completionRequest.Model = modelMap[completionRequest.Model]
		isModelMapped = true
	}

	// 开始计算Tokens
	var promptTokens int
	promptTokens = common.CountTokenInput(completionRequest.Prompt, completionRequest.Model)

	// 计算预付费配额
	quotaInfo := &QuotaInfo{
		modelName:    completionRequest.Model,
		promptTokens: promptTokens,
		userId:       userId,
		channelId:    channelId,
		tokenId:      tokenId,
	}
	quotaInfo.initQuotaInfo(group)
	quota_err := quotaInfo.preQuotaConsumption()
	if quota_err != nil {
		return quota_err
	}

	usage, openAIErrorWithStatusCode := completionProvider.CompleteResponse(&completionRequest, isModelMapped, promptTokens)

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

func GetCompletionProvider(channelType int, c *gin.Context) providers.CompletionProviderAction {
	switch channelType {
	case common.ChannelTypeOpenAI:
		return providers.CreateOpenAIProvider(c, "")
	case common.ChannelTypeAzure:
		return providers.CreateAzureProvider(c)
	}

	baseURL := common.ChannelBaseURLs[channelType]
	if c.GetString("base_url") != "" {
		baseURL = c.GetString("base_url")
	}

	if baseURL != "" {
		return providers.CreateOpenAIProvider(c, baseURL)
	}

	return nil
}
