package controller

import (
	"net/http"
	"one-api/common"
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

	// 获取供应商
	provider, modelName, fail := getProvider(c, imageEditRequest.Model)
	if fail {
		return
	}
	imageEditRequest.Model = modelName

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

	usage := &types.Usage{
		PromptTokens: promptTokens,
	}
	provider.SetUsage(usage)

	quotaInfo, errWithCode := generateQuotaInfo(c, imageEditRequest.Model, promptTokens)
	if errWithCode != nil {
		errorHelper(c, errWithCode)
		return
	}

	response, errWithCode := imageVariations.CreateImageVariations(&imageEditRequest)
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
