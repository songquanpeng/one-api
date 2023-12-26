package controller

import (
	"context"
	"net/http"
	"one-api/common"
	"one-api/model"
	providersBase "one-api/providers/base"
	"one-api/types"
	"strings"

	"github.com/gin-gonic/gin"
)

func RelayEmbeddings(c *gin.Context) {

	var embeddingsRequest types.EmbeddingRequest
	if strings.HasSuffix(c.Request.URL.Path, "embeddings") {
		embeddingsRequest.Model = c.Param("model")
	}

	if err := common.UnmarshalBodyReusable(c, &embeddingsRequest); err != nil {
		common.AbortWithMessage(c, http.StatusBadRequest, err.Error())
		return
	}

	channel, pass := fetchChannel(c, embeddingsRequest.Model)
	if pass {
		return
	}

	// 解析模型映射
	var isModelMapped bool
	modelMap, err := parseModelMapping(channel.GetModelMapping())
	if err != nil {
		common.AbortWithMessage(c, http.StatusInternalServerError, err.Error())
		return
	}
	if modelMap != nil && modelMap[embeddingsRequest.Model] != "" {
		embeddingsRequest.Model = modelMap[embeddingsRequest.Model]
		isModelMapped = true
	}

	// 获取供应商
	provider, pass := getProvider(c, channel, common.RelayModeEmbeddings)
	if pass {
		return
	}
	embeddingsProvider, ok := provider.(providersBase.EmbeddingsInterface)
	if !ok {
		common.AbortWithMessage(c, http.StatusNotImplemented, "channel not implemented")
		return
	}

	// 获取Input Tokens
	promptTokens := common.CountTokenInput(embeddingsRequest.Input, embeddingsRequest.Model)

	var quotaInfo *QuotaInfo
	var errWithCode *types.OpenAIErrorWithStatusCode
	var usage *types.Usage
	quotaInfo, errWithCode = generateQuotaInfo(c, embeddingsRequest.Model, promptTokens)
	if errWithCode != nil {
		errorHelper(c, errWithCode)
		return
	}

	usage, errWithCode = embeddingsProvider.EmbeddingsAction(&embeddingsRequest, isModelMapped, promptTokens)

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
