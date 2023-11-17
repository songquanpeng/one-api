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
	"path"

	"github.com/gin-gonic/gin"
)

func relayAudioHelper(c *gin.Context, relayMode int) *OpenAIErrorWithStatusCode {
	audioModel := "whisper-1"

	tokenId := c.GetInt("token_id")
	channelType := c.GetInt("channel")
	channelId := c.GetInt("channel_id")
	userId := c.GetInt("id")
	group := c.GetString("group")

	// Get last path of request URL
	// Example: v1/audio/speech -> speech
	requestPath := path.Base(c.Request.URL.Path) // speech

	var ttsRequest TextToSpeechRequest

	if requestPath == "speech" {
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

	preConsumedTokens := common.PreConsumedQuota
	modelRatio := common.GetModelRatio(audioModel)
	groupRatio := common.GetGroupRatio(group)
	ratio := modelRatio * groupRatio
	preConsumedQuota := int(float64(preConsumedTokens) * ratio)
	userQuota, err := model.CacheGetUserQuota(userId)

	if err != nil {
		return errorWrapper(err, "get_user_quota_failed", http.StatusInternalServerError)
	}

	quota := 0

	// Check if user quota is enough
	if requestPath == "speech" {
		quota = int(float64(len(ttsRequest.Input)) * modelRatio * groupRatio)

		fmt.Print(len(ttsRequest.Input), quota)

		if quota > userQuota {
			return errorWrapper(errors.New("user quota is not enough"), "insufficient_user_quota", http.StatusForbidden)
		}
	} else {
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
	requestBody := c.Request.Body

	req, err := http.NewRequest(c.Request.Method, fullRequestURL, requestBody)
	if err != nil {
		return errorWrapper(err, "new_request_failed", http.StatusInternalServerError)
	}
	req.Header.Set("Authorization", c.Request.Header.Get("Authorization"))
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

	responseBody, err := io.ReadAll(resp.Body)

	if err != nil {
		return errorWrapper(err, "read_response_body_failed", http.StatusInternalServerError)
	}
	err = resp.Body.Close()
	if err != nil {
		return errorWrapper(err, "close_response_body_failed", http.StatusInternalServerError)
	}

	if requestPath == "speech" {
		defer func(ctx context.Context) {
			go func(quota int) {
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
					model.RecordConsumeLog(ctx, userId, channelId, 0, 0, audioModel, tokenName, quota, logContent)
					model.UpdateUserUsedQuotaAndRequestCount(userId, quota)
					channelId := c.GetInt("channel_id")
					model.UpdateChannelUsedQuota(channelId, quota)
				}
			}(quota)
		}(c.Request.Context())
	} else {
		var whisperResponse WhisperResponse

		defer func(ctx context.Context) {
			go func() {
				quota := countTokenText(whisperResponse.Text, audioModel)
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
					model.RecordConsumeLog(ctx, userId, channelId, 0, 0, audioModel, tokenName, quota, logContent)
					model.UpdateUserUsedQuotaAndRequestCount(userId, quota)
					channelId := c.GetInt("channel_id")
					model.UpdateChannelUsedQuota(channelId, quota)
				}
			}()
		}(c.Request.Context())

		err = json.Unmarshal(responseBody, &whisperResponse)

		if err != nil {
			return errorWrapper(err, "unmarshal_response_body_failed", http.StatusInternalServerError)
		}
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
