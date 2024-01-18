package controller

import (
	"net/http"
	"one-api/common"
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

	// 获取供应商
	provider, modelName, fail := getProvider(c, embeddingsRequest.Model)
	if fail {
		return
	}
	embeddingsRequest.Model = modelName

	embeddingsProvider, ok := provider.(providersBase.EmbeddingsInterface)
	if !ok {
		common.AbortWithMessage(c, http.StatusNotImplemented, "channel not implemented")
		return
	}

	// 获取Input Tokens
	promptTokens := common.CountTokenInput(embeddingsRequest.Input, embeddingsRequest.Model)

	usage := &types.Usage{
		PromptTokens: promptTokens,
	}
	provider.SetUsage(usage)

	quotaInfo, errWithCode := generateQuotaInfo(c, embeddingsRequest.Model, promptTokens)
	if errWithCode != nil {
		errorHelper(c, errWithCode)
		return
	}

	response, errWithCode := embeddingsProvider.CreateEmbeddings(&embeddingsRequest)
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
