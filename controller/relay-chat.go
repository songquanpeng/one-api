package controller

import (
	"math"
	"net/http"
	"one-api/common"
	"one-api/common/requester"
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

	if chatRequest.MaxTokens < 0 || chatRequest.MaxTokens > math.MaxInt32/2 {
		common.AbortWithMessage(c, http.StatusBadRequest, "max_tokens is invalid")
		return
	}

	// 获取供应商
	provider, modelName, fail := getProvider(c, chatRequest.Model)
	if fail {
		return
	}
	chatRequest.Model = modelName

	chatProvider, ok := provider.(providersBase.ChatInterface)
	if !ok {
		common.AbortWithMessage(c, http.StatusNotImplemented, "channel not implemented")
		return
	}

	// 获取Input Tokens
	promptTokens := common.CountTokenMessages(chatRequest.Messages, chatRequest.Model)

	usage := &types.Usage{
		PromptTokens: promptTokens,
	}
	provider.SetUsage(usage)

	quotaInfo, errWithCode := generateQuotaInfo(c, chatRequest.Model, promptTokens)
	if errWithCode != nil {
		errorHelper(c, errWithCode)
		return
	}

	if chatRequest.Stream {
		var response requester.StreamReaderInterface[string]
		response, errWithCode = chatProvider.CreateChatCompletionStream(&chatRequest)
		if errWithCode != nil {
			errorHelper(c, errWithCode)
			return
		}
		errWithCode = responseStreamClient(c, response)
	} else {
		var response *types.ChatCompletionResponse
		response, errWithCode = chatProvider.CreateChatCompletion(&chatRequest)
		if errWithCode != nil {
			errorHelper(c, errWithCode)
			return
		}
		errWithCode = responseJsonClient(c, response)
	}

	// 如果报错，则退还配额
	if errWithCode != nil {
		quotaInfo.undo(c, errWithCode)
		return
	}

	quotaInfo.consume(c, usage)
}
