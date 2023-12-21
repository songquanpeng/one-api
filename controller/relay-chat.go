package controller

import (
	"context"
	"math"
	"net/http"
	"one-api/common"
	"one-api/model"
	providersBase "one-api/providers/base"
	"one-api/types"

	"github.com/gin-gonic/gin"
)

func RelayChat(c *gin.Context) {

	var chatRequest types.ChatCompletionRequest
	if err := common.UnmarshalBodyReusable(c, &chatRequest); err != nil {
		common.AbortWithMessage(c, http.StatusBadRequest, err.Error())
		return
	}

	channel, pass := fetchChannel(c, chatRequest.Model)
	if pass {
		return
	}

	if chatRequest.MaxTokens < 0 || chatRequest.MaxTokens > math.MaxInt32/2 {
		common.AbortWithMessage(c, http.StatusBadRequest, "max_tokens is invalid")
		return
	}

	// 解析模型映射
	var isModelMapped bool
	modelMap, err := parseModelMapping(channel.GetModelMapping())
	if err != nil {
		common.AbortWithMessage(c, http.StatusInternalServerError, err.Error())
		return
	}
	if modelMap != nil && modelMap[chatRequest.Model] != "" {
		chatRequest.Model = modelMap[chatRequest.Model]
		isModelMapped = true
	}

	// 获取供应商
	provider, pass := getProvider(c, channel.Type, common.RelayModeChatCompletions)
	if pass {
		return
	}
	chatProvider, ok := provider.(providersBase.ChatInterface)
	if !ok {
		common.AbortWithMessage(c, http.StatusNotImplemented, "channel not implemented")
		return
	}

	// 获取Input Tokens
	promptTokens := common.CountTokenMessages(chatRequest.Messages, chatRequest.Model)

	var quotaInfo *QuotaInfo
	var errWithCode *types.OpenAIErrorWithStatusCode
	var usage *types.Usage
	quotaInfo, errWithCode = generateQuotaInfo(c, chatRequest.Model, promptTokens)
	if errWithCode != nil {
		errorHelper(c, errWithCode)
		return
	}

	usage, errWithCode = chatProvider.ChatAction(&chatRequest, isModelMapped, promptTokens)

	// 如果报错，则退还配额
	if errWithCode != nil {
		tokenId := c.GetInt("token_id")
		if quotaInfo.HandelStatus {
			go func(ctx context.Context) {
				// return pre-consumed quota
				err := model.PostConsumeTokenQuota(tokenId, -quotaInfo.preConsumedQuota)
				if err != nil {
					common.LogError(ctx, "error return pre-consumed quota: "+err.Error())
				}
			}(c.Request.Context())
		}
		errorHelper(c, errWithCode)
		return
	} else {
		tokenName := c.GetString("token_name")
		// 如果没有报错，则消费配额
		go func(ctx context.Context) {
			err = quotaInfo.completedQuotaConsumption(usage, tokenName, ctx)
			if err != nil {
				common.LogError(ctx, err.Error())
			}
		}(c.Request.Context())
	}
}
