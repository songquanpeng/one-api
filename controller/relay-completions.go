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

func RelayCompletions(c *gin.Context) {

	var completionRequest types.CompletionRequest
	if err := common.UnmarshalBodyReusable(c, &completionRequest); err != nil {
		common.AbortWithMessage(c, http.StatusBadRequest, err.Error())
		return
	}

	if completionRequest.MaxTokens < 0 || completionRequest.MaxTokens > math.MaxInt32/2 {
		common.AbortWithMessage(c, http.StatusBadRequest, "max_tokens is invalid")
		return
	}

	// 获取供应商
	provider, modelName, fail := getProvider(c, completionRequest.Model)
	if fail {
		return
	}
	completionRequest.Model = modelName

	completionProvider, ok := provider.(providersBase.CompletionInterface)
	if !ok {
		common.AbortWithMessage(c, http.StatusNotImplemented, "channel not implemented")
		return
	}

	// 获取Input Tokens
	promptTokens := common.CountTokenInput(completionRequest.Prompt, completionRequest.Model)

	usage := &types.Usage{
		PromptTokens: promptTokens,
	}
	provider.SetUsage(usage)

	quotaInfo, errWithCode := generateQuotaInfo(c, completionRequest.Model, promptTokens)
	if errWithCode != nil {
		errorHelper(c, errWithCode)
		return
	}

	if completionRequest.Stream {
		var response requester.StreamReaderInterface[string]
		response, errWithCode = completionProvider.CreateCompletionStream(&completionRequest)
		if errWithCode != nil {
			errorHelper(c, errWithCode)
			return
		}
		errWithCode = responseStreamClient(c, response)
	} else {
		var response *types.CompletionResponse
		response, errWithCode = completionProvider.CreateCompletion(&completionRequest)
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
