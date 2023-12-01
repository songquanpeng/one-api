package controller

import (
	"context"
	"net/http"
	"one-api/common"
	"one-api/model"
	providersBase "one-api/providers/base"
	"one-api/types"

	"github.com/gin-gonic/gin"
)

func RelayCompletions(c *gin.Context) {

	var completionRequest types.CompletionRequest
	if err := common.UnmarshalBodyReusable(c, &completionRequest); err != nil {
		common.AbortWithMessage(c, http.StatusBadRequest, err.Error())
		return
	}

	channel, pass := fetchChannel(c, completionRequest.Model)
	if pass {
		return
	}

	// 写入渠道信息
	setChannelToContext(c, channel)

	// 解析模型映射
	var isModelMapped bool
	modelMap, err := parseModelMapping(c.GetString("model_mapping"))
	if err != nil {
		common.AbortWithMessage(c, http.StatusInternalServerError, err.Error())
		return
	}
	if modelMap != nil && modelMap[completionRequest.Model] != "" {
		completionRequest.Model = modelMap[completionRequest.Model]
		isModelMapped = true
	}

	// 获取供应商
	provider, pass := getProvider(c, channel.Type, common.RelayModeCompletions)
	if pass {
		return
	}
	completionProvider, ok := provider.(providersBase.CompletionInterface)
	if !ok {
		common.AbortWithMessage(c, http.StatusNotImplemented, "channel not implemented")
		return
	}

	// 获取Input Tokens
	promptTokens := common.CountTokenInput(completionRequest.Prompt, completionRequest.Model)

	var quotaInfo *QuotaInfo
	var errWithCode *types.OpenAIErrorWithStatusCode
	var usage *types.Usage
	quotaInfo, errWithCode = generateQuotaInfo(c, completionRequest.Model, promptTokens)
	if errWithCode != nil {
		errorHelper(c, errWithCode)
		return
	}

	usage, errWithCode = completionProvider.CompleteAction(&completionRequest, isModelMapped, promptTokens)

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
