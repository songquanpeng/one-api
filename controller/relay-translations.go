package controller

import (
	"net/http"
	"one-api/common"
	providersBase "one-api/providers/base"
	"one-api/types"

	"github.com/gin-gonic/gin"
)

func RelayTranslations(c *gin.Context) {

	var audioRequest types.AudioRequest

	if err := common.UnmarshalBodyReusable(c, &audioRequest); err != nil {
		common.AbortWithMessage(c, http.StatusBadRequest, err.Error())
		return
	}

	// 获取供应商
	provider, modelName, fail := getProvider(c, audioRequest.Model)
	if fail {
		return
	}
	audioRequest.Model = modelName

	translationProvider, ok := provider.(providersBase.TranslationInterface)
	if !ok {
		common.AbortWithMessage(c, http.StatusNotImplemented, "channel not implemented")
		return
	}

	// 获取Input Tokens
	promptTokens := 0

	usage := &types.Usage{
		PromptTokens: promptTokens,
	}
	provider.SetUsage(usage)

	quotaInfo, errWithCode := generateQuotaInfo(c, audioRequest.Model, promptTokens)
	if errWithCode != nil {
		errorHelper(c, errWithCode)
		return
	}

	response, errWithCode := translationProvider.CreateTranslation(&audioRequest)
	if errWithCode != nil {
		errorHelper(c, errWithCode)
		return
	}
	errWithCode = responseCustom(c, response)

	// 如果报错，则退还配额
	if errWithCode != nil {
		quotaInfo.undo(c, errWithCode)
		return
	}

	quotaInfo.consume(c, usage)
}
