package controller

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"one-api/common"
	"one-api/model"
	"strings"

	"github.com/gin-gonic/gin"
)

func isWithinRange(element string, value int) bool {
	if _, ok := common.DalleGenerationImageAmounts[element]; !ok {
		return false
	}
	min := common.DalleGenerationImageAmounts[element][0]
	max := common.DalleGenerationImageAmounts[element][1]

	return value >= min && value <= max
}

func relayImageHelper(c *gin.Context, relayMode int) *OpenAIErrorWithStatusCode {
	imageModel := "dall-e-2"
	imageSize := "1024x1024"

	tokenId := c.GetInt("token_id")
	channelType := c.GetInt("channel")
	channelId := c.GetInt("channel_id")
	userId := c.GetInt("id")
	group := c.GetString("group")

	var imageRequest ImageRequest
	err := common.UnmarshalBodyReusable(c, &imageRequest)
	if err != nil {
		return errorWrapper(err, "bind_request_body_failed", http.StatusBadRequest)
	}

	if imageRequest.N == 0 {
		imageRequest.N = 1
	}

	// Size validation
	if imageRequest.Size != "" {
		imageSize = imageRequest.Size
	}

	// Model validation
	if imageRequest.Model != "" {
		imageModel = imageRequest.Model
	}

	imageCostRatio, hasValidSize := common.DalleSizeRatios[imageModel][imageSize]

	// Check if model is supported
	if hasValidSize {
		if imageRequest.Quality == "hd" && imageModel == "dall-e-3" {
			if imageSize == "1024x1024" {
				imageCostRatio *= 2
			} else {
				imageCostRatio *= 1.5
			}
		}
	} else {
		return errorWrapper(errors.New("size not supported for this image model"), "size_not_supported", http.StatusBadRequest)
	}

	// Prompt validation
	if imageRequest.Prompt == "" {
		return errorWrapper(errors.New("prompt is required"), "prompt_missing", http.StatusBadRequest)
	}

	// Check prompt length
	if len(imageRequest.Prompt) > common.DalleImagePromptLengthLimitations[imageModel] {
		return errorWrapper(errors.New("prompt is too long"), "prompt_too_long", http.StatusBadRequest)
	}

	// Number of generated images validation
	if isWithinRange(imageModel, imageRequest.N) == false {
		// channel not azure
		if channelType != common.ChannelTypeAzure {
			return errorWrapper(errors.New("invalid value of n"), "n_not_within_range", http.StatusBadRequest)
		}
	}

	// map model name
	modelMapping := c.GetString("model_mapping")
	isModelMapped := false
	if modelMapping != "" {
		modelMap := make(map[string]string)
		err := json.Unmarshal([]byte(modelMapping), &modelMap)
		if err != nil {
			return errorWrapper(err, "unmarshal_model_mapping_failed", http.StatusInternalServerError)
		}
		if modelMap[imageModel] != "" {
			imageModel = modelMap[imageModel]
			isModelMapped = true
		}
	}
	baseURL := common.ChannelBaseURLs[channelType]
	requestURL := c.Request.URL.String()
	if c.GetString("base_url") != "" {
		baseURL = c.GetString("base_url")
	}
	fullRequestURL := getFullRequestURL(baseURL, requestURL, channelType)
	if channelType == common.ChannelTypeAzure {
		// https://learn.microsoft.com/en-us/azure/ai-services/openai/dall-e-quickstart?tabs=dalle3%2Ccommand-line&pivots=rest-api
		apiVersion := GetAPIVersion(c)
		// https://{resource_name}.openai.azure.com/openai/deployments/dall-e-3/images/generations?api-version=2023-06-01-preview
		fullRequestURL = fmt.Sprintf("%s/openai/deployments/%s/images/generations?api-version=%s", baseURL, imageModel, apiVersion)
	}

	var requestBody io.Reader
	if isModelMapped || channelType == common.ChannelTypeAzure { // make Azure channel request body
		jsonStr, err := json.Marshal(imageRequest)
		if err != nil {
			return errorWrapper(err, "marshal_text_request_failed", http.StatusInternalServerError)
		}
		requestBody = bytes.NewBuffer(jsonStr)
	} else {
		requestBody = c.Request.Body
	}

	modelRatio := common.GetModelRatio(imageModel)
	groupRatio := common.GetGroupRatio(group)
	ratio := modelRatio * groupRatio
	userQuota, err := model.CacheGetUserQuota(userId)

	quota := int(ratio*imageCostRatio*1000) * imageRequest.N

	if userQuota-quota < 0 {
		return errorWrapper(errors.New("user quota is not enough"), "insufficient_user_quota", http.StatusForbidden)
	}

	req, err := http.NewRequest(c.Request.Method, fullRequestURL, requestBody)
	if err != nil {
		return errorWrapper(err, "new_request_failed", http.StatusInternalServerError)
	}
	token := c.Request.Header.Get("Authorization")
	if channelType == common.ChannelTypeAzure { // Azure authentication
		token = strings.TrimPrefix(token, "Bearer ")
		req.Header.Set("api-key", token)
	} else {
		req.Header.Set("Authorization", token)
	}

	req.Header.Set("Content-Type", c.Request.Header.Get("Content-Type"))
	req.Header.Set("Accept", c.Request.Header.Get("Accept"))

	resp, err := httpClient.Do(req)
	if err != nil {
		return errorWrapper(err, "do_request_failed", http.StatusInternalServerError)
	}

	err = req.Body.Close()
	if err != nil {
		return errorWrapper(err, "close_request_body_failed", http.StatusInternalServerError)
	}
	err = c.Request.Body.Close()
	if err != nil {
		return errorWrapper(err, "close_request_body_failed", http.StatusInternalServerError)
	}
	var textResponse ImageResponse

	defer func(ctx context.Context) {
		if resp.StatusCode != http.StatusOK {
			return
		}
		err := model.PostConsumeTokenQuota(tokenId, quota)
		if err != nil {
			common.SysError("error consuming token remain quota: " + err.Error())
		}
		err = model.CacheUpdateUserQuota(userId)
		if err != nil {
			common.SysError("error update user quota cache: " + err.Error())
		}
		if quota != 0 {
			tokenName := c.GetString("token_name")
			logContent := fmt.Sprintf("模型倍率 %.2f，分组倍率 %.2f", modelRatio, groupRatio)
			model.RecordConsumeLog(ctx, userId, channelId, 0, 0, imageModel, tokenName, quota, logContent)
			model.UpdateUserUsedQuotaAndRequestCount(userId, quota)
			channelId := c.GetInt("channel_id")
			model.UpdateChannelUsedQuota(channelId, quota)
		}
	}(c.Request.Context())

	responseBody, err := io.ReadAll(resp.Body)

	if err != nil {
		return errorWrapper(err, "read_response_body_failed", http.StatusInternalServerError)
	}
	err = resp.Body.Close()
	if err != nil {
		return errorWrapper(err, "close_response_body_failed", http.StatusInternalServerError)
	}
	err = json.Unmarshal(responseBody, &textResponse)
	if err != nil {
		return errorWrapper(err, "unmarshal_response_body_failed", http.StatusInternalServerError)
	}

	resp.Body = io.NopCloser(bytes.NewBuffer(responseBody))

	for k, v := range resp.Header {
		c.Writer.Header().Set(k, v[0])
	}
	c.Writer.WriteHeader(resp.StatusCode)

	_, err = io.Copy(c.Writer, resp.Body)
	if err != nil {
		return errorWrapper(err, "copy_response_body_failed", http.StatusInternalServerError)
	}
	err = resp.Body.Close()
	if err != nil {
		return errorWrapper(err, "close_response_body_failed", http.StatusInternalServerError)
	}
	return nil
}
