package controller

import (
	"net/http"
	"one-api/common"
	providersBase "one-api/providers/base"
	"one-api/types"

	"github.com/gin-gonic/gin"
)

func RelayImageGenerations(c *gin.Context) {

	var imageRequest types.ImageRequest

	if err := common.UnmarshalBodyReusable(c, &imageRequest); err != nil {
		common.AbortWithMessage(c, http.StatusBadRequest, err.Error())
		return
	}

	if imageRequest.Model == "" {
		imageRequest.Model = "dall-e-2"
	}

	if imageRequest.N == 0 {
		imageRequest.N = 1
	}

	if imageRequest.Size == "" {
		imageRequest.Size = "1024x1024"
	}

	if imageRequest.Quality == "" {
		imageRequest.Quality = "standard"
	}

	// 获取供应商
	provider, modelName, fail := getProvider(c, imageRequest.Model)
	if fail {
		return
	}
	imageRequest.Model = modelName

	imageGenerationsProvider, ok := provider.(providersBase.ImageGenerationsInterface)
	if !ok {
		common.AbortWithMessage(c, http.StatusNotImplemented, "channel not implemented")
		return
	}

	// 获取Input Tokens
	promptTokens, err := common.CountTokenImage(imageRequest)
	if err != nil {
		common.AbortWithMessage(c, http.StatusInternalServerError, err.Error())
		return
	}

	usage := &types.Usage{
		PromptTokens: promptTokens,
	}
	provider.SetUsage(usage)

	quotaInfo, errWithCode := generateQuotaInfo(c, imageRequest.Model, promptTokens)
	if errWithCode != nil {
		errorHelper(c, errWithCode)
		return
	}

	response, errWithCode := imageGenerationsProvider.CreateImageGenerations(&imageRequest)
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
