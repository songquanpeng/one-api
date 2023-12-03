package controller

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"io"
	"net/http"
	"one-api/common"
	"one-api/model"
	"strings"
)

func relayAudioHelper(c *gin.Context, relayMode int) *OpenAIErrorWithStatusCode {
	audioModel := "whisper-1"

	tokenId := c.GetInt("token_id")
	channelType := c.GetInt("channel")
	channelId := c.GetInt("channel_id")
	userId := c.GetInt("id")
	group := c.GetString("group")
	tokenName := c.GetString("token_name")

	var ttsRequest TextToSpeechRequest
	if relayMode == RelayModeAudioSpeech {
		// Read JSON
		err := common.UnmarshalBodyReusable(c, &ttsRequest)
		// Check if JSON is valid
		if err != nil {
			return errorWrapper(err, "invalid_json", http.StatusBadRequest)
		}
		audioModel = ttsRequest.Model
		// Check if text is too long 4096
		if len(ttsRequest.Input) > 4096 {
			return errorWrapper(errors.New("input is too long (over 4096 characters)"), "text_too_long", http.StatusBadRequest)
		}
	}

	modelRatio := common.GetModelRatio(audioModel)
	groupRatio := common.GetGroupRatio(group)
	ratio := modelRatio * groupRatio
	var quota int
	var preConsumedQuota int
	switch relayMode {
	case RelayModeAudioSpeech:
		preConsumedQuota = int(float64(len(ttsRequest.Input)) * ratio)
		quota = preConsumedQuota
	default:
		preConsumedQuota = int(float64(common.PreConsumedQuota) * ratio)
	}
	userQuota, err := model.CacheGetUserQuota(userId)
	if err != nil {
		return errorWrapper(err, "get_user_quota_failed", http.StatusInternalServerError)
	}

	// Check if user quota is enough
	if userQuota-preConsumedQuota < 0 {
		return errorWrapper(errors.New("user quota is not enough"), "insufficient_user_quota", http.StatusForbidden)
	}
	err = model.CacheDecreaseUserQuota(userId, preConsumedQuota)
	if err != nil {
		return errorWrapper(err, "decrease_user_quota_failed", http.StatusInternalServerError)
	}
	if userQuota > 100*preConsumedQuota {
		// in this case, we do not pre-consume quota
		// because the user has enough quota
		preConsumedQuota = 0
	}
	if preConsumedQuota > 0 {
		err := model.PreConsumeTokenQuota(tokenId, preConsumedQuota)
		if err != nil {
			return errorWrapper(err, "pre_consume_token_quota_failed", http.StatusForbidden)
		}
	}

	// map model name
	modelMapping := c.GetString("model_mapping")
	if modelMapping != "" {
		modelMap := make(map[string]string)
		err := json.Unmarshal([]byte(modelMapping), &modelMap)
		if err != nil {
			return errorWrapper(err, "unmarshal_model_mapping_failed", http.StatusInternalServerError)
		}
		if modelMap[audioModel] != "" {
			audioModel = modelMap[audioModel]
		}
	}

	baseURL := common.ChannelBaseURLs[channelType]
	requestURL := c.Request.URL.String()
	if c.GetString("base_url") != "" {
		baseURL = c.GetString("base_url")
	}

	fullRequestURL := getFullRequestURL(baseURL, requestURL, channelType)
	if relayMode == RelayModeAudioTranscription && channelType == common.ChannelTypeAzure {
		// https://learn.microsoft.com/en-us/azure/ai-services/openai/whisper-quickstart?tabs=command-line#rest-api
		apiVersion := GetAPIVersion(c)
		fullRequestURL = fmt.Sprintf("%s/openai/deployments/%s/audio/transcriptions?api-version=%s", baseURL, audioModel, apiVersion)
	}

	requestBody := c.Request.Body

	req, err := http.NewRequest(c.Request.Method, fullRequestURL, requestBody)
	if err != nil {
		return errorWrapper(err, "new_request_failed", http.StatusInternalServerError)
	}

	if relayMode == RelayModeAudioTranscription && channelType == common.ChannelTypeAzure {
		// https://learn.microsoft.com/en-us/azure/ai-services/openai/whisper-quickstart?tabs=command-line#rest-api
		apiKey := c.Request.Header.Get("Authorization")
		apiKey = strings.TrimPrefix(apiKey, "Bearer ")
		req.Header.Set("api-key", apiKey)
		req.ContentLength = c.Request.ContentLength
	} else {
		req.Header.Set("Authorization", c.Request.Header.Get("Authorization"))
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

	if relayMode != RelayModeAudioSpeech {
		responseBody, err := io.ReadAll(resp.Body)
		if err != nil {
			return errorWrapper(err, "read_response_body_failed", http.StatusInternalServerError)
		}
		err = resp.Body.Close()
		if err != nil {
			return errorWrapper(err, "close_response_body_failed", http.StatusInternalServerError)
		}
		var whisperResponse WhisperResponse
		err = json.Unmarshal(responseBody, &whisperResponse)
		if err != nil {
			return errorWrapper(err, "unmarshal_response_body_failed", http.StatusInternalServerError)
		}
		quota = countTokenText(whisperResponse.Text, audioModel)
		resp.Body = io.NopCloser(bytes.NewBuffer(responseBody))
	}
	if resp.StatusCode != http.StatusOK {
		if preConsumedQuota > 0 {
			// we need to roll back the pre-consumed quota
			defer func(ctx context.Context) {
				go func() {
					// negative means add quota back for token & user
					err := model.PostConsumeTokenQuota(tokenId, -preConsumedQuota)
					if err != nil {
						common.LogError(ctx, fmt.Sprintf("error rollback pre-consumed quota: %s", err.Error()))
					}
				}()
			}(c.Request.Context())
		}
		return relayErrorHandler(resp)
	}
	quotaDelta := quota - preConsumedQuota
	defer func(ctx context.Context) {
		go postConsumeQuota(ctx, tokenId, quotaDelta, quota, userId, channelId, modelRatio, groupRatio, audioModel, tokenName)
	}(c.Request.Context())

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
