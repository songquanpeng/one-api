package controller

import (
	"net/http"
	"one-api/common"
	providersBase "one-api/providers/base"
	"one-api/types"

	"github.com/gin-gonic/gin"
)

func RelaySpeech(c *gin.Context) {

	var speechRequest types.SpeechAudioRequest

	if err := common.UnmarshalBodyReusable(c, &speechRequest); err != nil {
		common.AbortWithMessage(c, http.StatusBadRequest, err.Error())
		return
	}

	// 获取供应商
	provider, modelName, fail := getProvider(c, speechRequest.Model)
	if fail {
		return
	}
	speechRequest.Model = modelName

	speechProvider, ok := provider.(providersBase.SpeechInterface)
	if !ok {
		common.AbortWithMessage(c, http.StatusNotImplemented, "channel not implemented")
		return
	}

	// 获取Input Tokens
	promptTokens := len(speechRequest.Input)

	usage := &types.Usage{
		PromptTokens: promptTokens,
	}
	provider.SetUsage(usage)

	quotaInfo, errWithCode := generateQuotaInfo(c, speechRequest.Model, promptTokens)
	if errWithCode != nil {
		errorHelper(c, errWithCode)
		return
	}

	response, errWithCode := speechProvider.CreateSpeech(&speechRequest)
	if errWithCode != nil {
		errorHelper(c, errWithCode)
		return
	}
	errWithCode = responseMultipart(c, response)

	// 如果报错，则退还配额
	if errWithCode != nil {
		quotaInfo.undo(c, errWithCode)
		return
	}

	quotaInfo.consume(c, usage)
}
