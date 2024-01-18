package controller

import (
	"net/http"
	"one-api/common"
	providersBase "one-api/providers/base"
	"one-api/types"

	"github.com/gin-gonic/gin"
)

func RelayModerations(c *gin.Context) {

	var moderationRequest types.ModerationRequest

	if err := common.UnmarshalBodyReusable(c, &moderationRequest); err != nil {
		common.AbortWithMessage(c, http.StatusBadRequest, err.Error())
		return
	}

	if moderationRequest.Model == "" {
		moderationRequest.Model = "text-moderation-stable"
	}

	// 获取供应商
	provider, modelName, fail := getProvider(c, moderationRequest.Model)
	if fail {
		return
	}
	moderationRequest.Model = modelName

	moderationProvider, ok := provider.(providersBase.ModerationInterface)
	if !ok {
		common.AbortWithMessage(c, http.StatusNotImplemented, "channel not implemented")
		return
	}

	// 获取Input Tokens
	promptTokens := common.CountTokenInput(moderationRequest.Input, moderationRequest.Model)

	usage := &types.Usage{
		PromptTokens: promptTokens,
	}
	provider.SetUsage(usage)

	quotaInfo, errWithCode := generateQuotaInfo(c, moderationRequest.Model, promptTokens)
	if errWithCode != nil {
		errorHelper(c, errWithCode)
		return
	}

	response, errWithCode := moderationProvider.CreateModeration(&moderationRequest)
	if errWithCode != nil {
		errorHelper(c, errWithCode)
		return
	}
	errWithCode = responseJsonClient(c, response)

	// 如果报错，则退还配额
	if errWithCode != nil {
		quotaInfo.undo(c, errWithCode)
		return
	}

	quotaInfo.consume(c, usage)
}
