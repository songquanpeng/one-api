package controller

import (
	"bytes"
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

const (
	APITypeOpenAI = iota
	APITypeClaude
	APITypePaLM
)

func relayTextHelper(c *gin.Context, relayMode int) *OpenAIErrorWithStatusCode {
	channelType := c.GetInt("channel")
	tokenId := c.GetInt("token_id")
	userId := c.GetInt("id")
	consumeQuota := c.GetBool("consume_quota")
	group := c.GetString("group")
	var textRequest GeneralOpenAIRequest
	if consumeQuota || channelType == common.ChannelTypeAzure || channelType == common.ChannelTypePaLM {
		err := common.UnmarshalBodyReusable(c, &textRequest)
		if err != nil {
			return errorWrapper(err, "bind_request_body_failed", http.StatusBadRequest)
		}
	}
	if relayMode == RelayModeModerations && textRequest.Model == "" {
		textRequest.Model = "text-moderation-latest"
	}
	if relayMode == RelayModeEmbeddings && textRequest.Model == "" {
		textRequest.Model = c.Param("model")
	}
	// request validation
	if textRequest.Model == "" {
		return errorWrapper(errors.New("model is required"), "required_field_missing", http.StatusBadRequest)
	}
	switch relayMode {
	case RelayModeCompletions:
		if textRequest.Prompt == "" {
			return errorWrapper(errors.New("field prompt is required"), "required_field_missing", http.StatusBadRequest)
		}
	case RelayModeChatCompletions:
		if textRequest.Messages == nil || len(textRequest.Messages) == 0 {
			return errorWrapper(errors.New("field messages is required"), "required_field_missing", http.StatusBadRequest)
		}
	case RelayModeEmbeddings:
	case RelayModeModerations:
		if textRequest.Input == "" {
			return errorWrapper(errors.New("field input is required"), "required_field_missing", http.StatusBadRequest)
		}
	case RelayModeEdits:
		if textRequest.Instruction == "" {
			return errorWrapper(errors.New("field instruction is required"), "required_field_missing", http.StatusBadRequest)
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
		if modelMap[textRequest.Model] != "" {
			textRequest.Model = modelMap[textRequest.Model]
			isModelMapped = true
		}
	}
	apiType := APITypeOpenAI
	if strings.HasPrefix(textRequest.Model, "claude") {
		apiType = APITypeClaude
	}
	baseURL := common.ChannelBaseURLs[channelType]
	requestURL := c.Request.URL.String()
	if c.GetString("base_url") != "" {
		baseURL = c.GetString("base_url")
	}
	fullRequestURL := fmt.Sprintf("%s%s", baseURL, requestURL)
	switch apiType {
	case APITypeOpenAI:
		if channelType == common.ChannelTypeAzure {
			// https://learn.microsoft.com/en-us/azure/cognitive-services/openai/chatgpt-quickstart?pivots=rest-api&tabs=command-line#rest-api
			query := c.Request.URL.Query()
			apiVersion := query.Get("api-version")
			if apiVersion == "" {
				apiVersion = c.GetString("api_version")
			}
			requestURL := strings.Split(requestURL, "?")[0]
			requestURL = fmt.Sprintf("%s?api-version=%s", requestURL, apiVersion)
			baseURL = c.GetString("base_url")
			task := strings.TrimPrefix(requestURL, "/v1/")
			model_ := textRequest.Model
			model_ = strings.Replace(model_, ".", "", -1)
			// https://github.com/songquanpeng/one-api/issues/67
			model_ = strings.TrimSuffix(model_, "-0301")
			model_ = strings.TrimSuffix(model_, "-0314")
			model_ = strings.TrimSuffix(model_, "-0613")
			fullRequestURL = fmt.Sprintf("%s/openai/deployments/%s/%s", baseURL, model_, task)
		}
	case APITypeClaude:
		fullRequestURL = "https://api.anthropic.com/v1/complete"
		if baseURL != "" {
			fullRequestURL = fmt.Sprintf("%s/v1/complete", baseURL)
		}
	}
	var promptTokens int
	var completionTokens int
	switch relayMode {
	case RelayModeChatCompletions:
		promptTokens = countTokenMessages(textRequest.Messages, textRequest.Model)
	case RelayModeCompletions:
		promptTokens = countTokenInput(textRequest.Prompt, textRequest.Model)
	case RelayModeModerations:
		promptTokens = countTokenInput(textRequest.Input, textRequest.Model)
	}
	preConsumedTokens := common.PreConsumedQuota
	if textRequest.MaxTokens != 0 {
		preConsumedTokens = promptTokens + textRequest.MaxTokens
	}
	modelRatio := common.GetModelRatio(textRequest.Model)
	groupRatio := common.GetGroupRatio(group)
	ratio := modelRatio * groupRatio
	preConsumedQuota := int(float64(preConsumedTokens) * ratio)
	userQuota, err := model.CacheGetUserQuota(userId)
	if err != nil {
		return errorWrapper(err, "get_user_quota_failed", http.StatusInternalServerError)
	}
	if userQuota > 10*preConsumedQuota {
		// in this case, we do not pre-consume quota
		// because the user has enough quota
		preConsumedQuota = 0
	}
	if consumeQuota && preConsumedQuota > 0 {
		err := model.PreConsumeTokenQuota(tokenId, preConsumedQuota)
		if err != nil {
			return errorWrapper(err, "pre_consume_token_quota_failed", http.StatusForbidden)
		}
	}
	var requestBody io.Reader
	if isModelMapped {
		jsonStr, err := json.Marshal(textRequest)
		if err != nil {
			return errorWrapper(err, "marshal_text_request_failed", http.StatusInternalServerError)
		}
		requestBody = bytes.NewBuffer(jsonStr)
	} else {
		requestBody = c.Request.Body
	}
	switch apiType {
	case APITypeClaude:
		claudeRequest := requestOpenAI2Claude(textRequest)
		jsonStr, err := json.Marshal(claudeRequest)
		if err != nil {
			return errorWrapper(err, "marshal_text_request_failed", http.StatusInternalServerError)
		}
		requestBody = bytes.NewBuffer(jsonStr)
	}
	req, err := http.NewRequest(c.Request.Method, fullRequestURL, requestBody)
	if err != nil {
		return errorWrapper(err, "new_request_failed", http.StatusInternalServerError)
	}
	apiKey := c.Request.Header.Get("Authorization")
	apiKey = strings.TrimPrefix(apiKey, "Bearer ")
	switch apiType {
	case APITypeOpenAI:
		if channelType == common.ChannelTypeAzure {
			req.Header.Set("api-key", apiKey)
		} else {
			req.Header.Set("Authorization", c.Request.Header.Get("Authorization"))
		}
	case APITypeClaude:
		req.Header.Set("x-api-key", apiKey)
		anthropicVersion := c.Request.Header.Get("anthropic-version")
		if anthropicVersion == "" {
			anthropicVersion = "2023-06-01"
		}
		req.Header.Set("anthropic-version", anthropicVersion)
	}
	req.Header.Set("Content-Type", c.Request.Header.Get("Content-Type"))
	req.Header.Set("Accept", c.Request.Header.Get("Accept"))
	//req.Header.Set("Connection", c.Request.Header.Get("Connection"))
	client := &http.Client{}
	resp, err := client.Do(req)
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
	var textResponse TextResponse
	isStream := strings.HasPrefix(resp.Header.Get("Content-Type"), "text/event-stream")
	var streamResponseText string

	defer func() {
		if consumeQuota {
			quota := 0
			completionRatio := 1.0
			if strings.HasPrefix(textRequest.Model, "gpt-3.5") {
				completionRatio = 1.333333
			}
			if strings.HasPrefix(textRequest.Model, "gpt-4") {
				completionRatio = 2
			}
			if isStream {
				completionTokens = countTokenText(streamResponseText, textRequest.Model)
			} else {
				promptTokens = textResponse.Usage.PromptTokens
				completionTokens = textResponse.Usage.CompletionTokens
			}
			quota = promptTokens + int(float64(completionTokens)*completionRatio)
			quota = int(float64(quota) * ratio)
			if ratio != 0 && quota <= 0 {
				quota = 1
			}
			totalTokens := promptTokens + completionTokens
			if totalTokens == 0 {
				// in this case, must be some error happened
				// we cannot just return, because we may have to return the pre-consumed quota
				quota = 0
			}
			quotaDelta := quota - preConsumedQuota
			err := model.PostConsumeTokenQuota(tokenId, quotaDelta)
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
				model.RecordConsumeLog(userId, promptTokens, completionTokens, textRequest.Model, tokenName, quota, logContent)
				model.UpdateUserUsedQuotaAndRequestCount(userId, quota)
				channelId := c.GetInt("channel_id")
				model.UpdateChannelUsedQuota(channelId, quota)
			}
		}
	}()
	switch apiType {
	case APITypeOpenAI:
		if isStream {
			err, responseText := openaiStreamHandler(c, resp, relayMode)
			if err != nil {
				return err
			}
			streamResponseText = responseText
			return nil
		} else {
			err, usage := openaiHandler(c, resp, consumeQuota)
			if err != nil {
				return err
			}
			textResponse.Usage = *usage
			return nil
		}
	case APITypeClaude:
		if isStream {
			err, responseText := claudeStreamHandler(c, resp)
			if err != nil {
				return err
			}
			streamResponseText = responseText
			return nil
		} else {
			err, usage := claudeHandler(c, resp, promptTokens, textRequest.Model)
			if err != nil {
				return err
			}
			textResponse.Usage = *usage
			return nil
		}
	default:
		return errorWrapper(errors.New("unknown api type"), "unknown_api_type", http.StatusInternalServerError)
	}
}
