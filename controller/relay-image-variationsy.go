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

func RelayImageVariations(c *gin.Context) {

	var imageEditRequest types.ImageEditRequest

	if err := common.UnmarshalBodyReusable(c, &imageEditRequest); err != nil {
		common.AbortWithMessage(c, http.StatusBadRequest, err.Error())
		return
	}

	if imageEditRequest.Model == "" {
		imageEditRequest.Model = "dall-e-2"
	}

	if imageEditRequest.Size == "" {
		imageEditRequest.Size = "1024x1024"
	}

	channel, pass := fetchChannel(c, imageEditRequest.Model)
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
	if modelMap != nil && modelMap[imageEditRequest.Model] != "" {
		imageEditRequest.Model = modelMap[imageEditRequest.Model]
		isModelMapped = true
	}

	// 获取供应商
	provider, pass := getProvider(c, channel.Type, common.RelayModeImagesVariations)
	if pass {
		return
	}
	imageVariations, ok := provider.(providersBase.ImageVariationsInterface)
	if !ok {
		common.AbortWithMessage(c, http.StatusNotImplemented, "channel not implemented")
		return
	}

	// 获取Input Tokens
	promptTokens, err := common.CountTokenImage(imageEditRequest)
	if err != nil {
		common.AbortWithMessage(c, http.StatusInternalServerError, err.Error())
		return
	}

	var quotaInfo *QuotaInfo
	var errWithCode *types.OpenAIErrorWithStatusCode
	var usage *types.Usage
	quotaInfo, errWithCode = generateQuotaInfo(c, imageEditRequest.Model, promptTokens)
	if errWithCode != nil {
		errorHelper(c, errWithCode)
		return
	}

	usage, errWithCode = imageVariations.ImageVariationsAction(&imageEditRequest, isModelMapped, promptTokens)

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
