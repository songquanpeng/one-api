package controller

import (
	"bufio"
	"bytes"
	"context"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/songquanpeng/one-api/common"
	"github.com/songquanpeng/one-api/common/client"
	"github.com/songquanpeng/one-api/common/config"
	"github.com/songquanpeng/one-api/common/ctxkey"
	"github.com/songquanpeng/one-api/common/logger"
	"github.com/songquanpeng/one-api/model"
	"github.com/songquanpeng/one-api/relay/adaptor/openai"
	"github.com/songquanpeng/one-api/relay/billing"
	billingratio "github.com/songquanpeng/one-api/relay/billing/ratio"
	"github.com/songquanpeng/one-api/relay/channeltype"
	"github.com/songquanpeng/one-api/relay/meta"
	relaymodel "github.com/songquanpeng/one-api/relay/model"
	"github.com/songquanpeng/one-api/relay/relaymode"
)

type minimaxSpeechRequest struct {
	Model        string              `json:"model"`
	Text         string              `json:"text"`
	Stream       bool                `json:"stream"`
	OutputFormat string              `json:"output_format"`
	VoiceSetting minimaxVoiceSetting `json:"voice_setting"`
	AudioSetting minimaxAudioSetting `json:"audio_setting"`
}

type minimaxVoiceSetting struct {
	VoiceID string  `json:"voice_id"`
	Speed   float64 `json:"speed"`
}

type minimaxAudioSetting struct {
	SampleRate int    `json:"sample_rate"`
	Bitrate    int    `json:"bitrate"`
	Format     string `json:"format"`
	Channel    int    `json:"channel"`
}

type minimaxSpeechResponse struct {
	Data struct {
		Audio string `json:"audio"`
	} `json:"data"`
	BaseResp struct {
		StatusCode int    `json:"status_code"`
		StatusMsg  string `json:"status_msg"`
	} `json:"base_resp"`
}

func buildMinimaxSpeechRequest(ttsRequest openai.TextToSpeechRequest) ([]byte, string, error) {
	format := strings.ToLower(ttsRequest.ResponseFormat)
	if format == "" {
		format = "mp3"
	}
	switch format {
	case "mp3", "wav", "flac", "pcm":
	default:
		return nil, "", fmt.Errorf("unsupported audio response format %q", format)
	}
	speed := ttsRequest.Speed
	if speed == 0 {
		speed = 1
	}
	request := minimaxSpeechRequest{
		Model:        ttsRequest.Model,
		Text:         ttsRequest.Input,
		Stream:       false,
		OutputFormat: "hex",
		VoiceSetting: minimaxVoiceSetting{VoiceID: ttsRequest.Voice, Speed: speed},
		AudioSetting: minimaxAudioSetting{SampleRate: 32000, Bitrate: 128000, Format: format, Channel: 1},
	}
	body, err := json.Marshal(request)
	return body, format, err
}

func decodeMinimaxSpeechResponse(body []byte) ([]byte, error) {
	var response minimaxSpeechResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, err
	}
	if response.BaseResp.StatusCode != 0 {
		return nil, fmt.Errorf("upstream speech error %d: %s", response.BaseResp.StatusCode, response.BaseResp.StatusMsg)
	}
	if response.Data.Audio == "" {
		return nil, errors.New("upstream speech response did not include audio")
	}
	audio := strings.TrimSpace(response.Data.Audio)
	if decoded, err := hex.DecodeString(audio); err == nil {
		return decoded, nil
	}
	if decoded, err := base64.StdEncoding.DecodeString(audio); err == nil {
		return decoded, nil
	}
	return nil, errors.New("upstream speech audio is not valid hex or base64")
}

func minimaxSpeechURL(baseURL string) string {
	baseURL = strings.TrimRight(baseURL, "/")
	baseURL = strings.TrimSuffix(baseURL, "/v1")
	if strings.Contains(baseURL, "api.minimax.chat") {
		baseURL = strings.Replace(baseURL, "api.minimax.chat", "api.minimax.io", 1)
	}
	return baseURL + "/v1/t2a_v2"
}

func audioContentType(format string) string {
	switch format {
	case "wav":
		return "audio/wav"
	case "flac":
		return "audio/flac"
	case "pcm":
		return "audio/L16"
	default:
		return "audio/mpeg"
	}
}

func RelayAudioHelper(c *gin.Context, relayMode int) *relaymodel.ErrorWithStatusCode {
	ctx := c.Request.Context()
	meta := meta.GetByContext(c)
	audioModel := "whisper-1"

	tokenId := c.GetInt(ctxkey.TokenId)
	channelType := c.GetInt(ctxkey.Channel)
	channelId := c.GetInt(ctxkey.ChannelId)
	userId := c.GetInt(ctxkey.Id)
	group := c.GetString(ctxkey.Group)
	tokenName := c.GetString(ctxkey.TokenName)

	var ttsRequest openai.TextToSpeechRequest
	if relayMode == relaymode.AudioSpeech {
		// Read JSON
		err := common.UnmarshalBodyReusable(c, &ttsRequest)
		// Check if JSON is valid
		if err != nil {
			return openai.ErrorWrapper(err, "invalid_json", http.StatusBadRequest)
		}
		audioModel = ttsRequest.Model
		// Check if text is too long 4096
		if len(ttsRequest.Input) > 4096 {
			return openai.ErrorWrapper(errors.New("input is too long (over 4096 characters)"), "text_too_long", http.StatusBadRequest)
		}
	}

	modelRatio := billingratio.GetModelRatio(audioModel, channelType)
	groupRatio := billingratio.GetGroupRatio(group)
	ratio := modelRatio * groupRatio
	var quota int64
	var preConsumedQuota int64
	switch relayMode {
	case relaymode.AudioSpeech:
		preConsumedQuota = int64(float64(len(ttsRequest.Input)) * ratio)
		quota = preConsumedQuota
	default:
		preConsumedQuota = int64(float64(config.PreConsumedQuota) * ratio)
	}
	userQuota, err := model.CacheGetUserQuota(ctx, userId)
	if err != nil {
		return openai.ErrorWrapper(err, "get_user_quota_failed", http.StatusInternalServerError)
	}

	// Check if user quota is enough
	if userQuota-preConsumedQuota < 0 {
		return openai.ErrorWrapper(errors.New("user quota is not enough"), "insufficient_user_quota", http.StatusForbidden)
	}
	err = model.CacheDecreaseUserQuota(userId, preConsumedQuota)
	if err != nil {
		return openai.ErrorWrapper(err, "decrease_user_quota_failed", http.StatusInternalServerError)
	}
	if userQuota > 100*preConsumedQuota {
		// in this case, we do not pre-consume quota
		// because the user has enough quota
		preConsumedQuota = 0
	}
	if preConsumedQuota > 0 {
		err := model.PreConsumeTokenQuota(tokenId, preConsumedQuota)
		if err != nil {
			return openai.ErrorWrapper(err, "pre_consume_token_quota_failed", http.StatusForbidden)
		}
	}
	succeed := false
	defer func() {
		if succeed {
			return
		}
		if preConsumedQuota > 0 {
			// we need to roll back the pre-consumed quota
			defer func(ctx context.Context) {
				go func() {
					// negative means add quota back for token & user
					err := model.PostConsumeTokenQuota(tokenId, -preConsumedQuota)
					if err != nil {
						logger.Error(ctx, fmt.Sprintf("error rollback pre-consumed quota: %s", err.Error()))
					}
				}()
			}(c.Request.Context())
		}
	}()

	// map model name
	modelMapping := c.GetStringMapString(ctxkey.ModelMapping)
	if modelMapping != nil && modelMapping[audioModel] != "" {
		audioModel = modelMapping[audioModel]
	}

	baseURL := channeltype.ChannelBaseURLs[channelType]
	requestURL := c.Request.URL.String()
	if c.GetString(ctxkey.BaseURL) != "" {
		baseURL = c.GetString(ctxkey.BaseURL)
	}

	fullRequestURL := openai.GetFullRequestURL(baseURL, requestURL, channelType)
	if channelType == channeltype.Azure {
		apiVersion := meta.Config.APIVersion
		if relayMode == relaymode.AudioTranscription {
			// https://learn.microsoft.com/en-us/azure/ai-services/openai/whisper-quickstart?tabs=command-line#rest-api
			fullRequestURL = fmt.Sprintf("%s/openai/deployments/%s/audio/transcriptions?api-version=%s", baseURL, audioModel, apiVersion)
		} else if relayMode == relaymode.AudioSpeech {
			// https://learn.microsoft.com/en-us/azure/ai-services/openai/text-to-speech-quickstart?tabs=command-line#rest-api
			fullRequestURL = fmt.Sprintf("%s/openai/deployments/%s/audio/speech?api-version=%s", baseURL, audioModel, apiVersion)
		}
	}

	requestBody := &bytes.Buffer{}
	_, err = io.Copy(requestBody, c.Request.Body)
	if err != nil {
		return openai.ErrorWrapper(err, "new_request_body_failed", http.StatusInternalServerError)
	}
	c.Request.Body = io.NopCloser(bytes.NewBuffer(requestBody.Bytes()))
	responseFormat := c.DefaultPostForm("response_format", "json")
	if relayMode == relaymode.AudioSpeech && ttsRequest.ResponseFormat != "" {
		responseFormat = strings.ToLower(ttsRequest.ResponseFormat)
	}
	if relayMode == relaymode.AudioSpeech && channelType == channeltype.Minimax {
		ttsRequest.Model = audioModel
		requestBodyBytes, format, requestErr := buildMinimaxSpeechRequest(ttsRequest)
		if requestErr != nil {
			return openai.ErrorWrapper(requestErr, "invalid_request_error", http.StatusBadRequest)
		}
		requestBody = bytes.NewBuffer(requestBodyBytes)
		responseFormat = format
		fullRequestURL = minimaxSpeechURL(baseURL)
	}

	req, err := http.NewRequest(c.Request.Method, fullRequestURL, requestBody)
	if err != nil {
		return openai.ErrorWrapper(err, "new_request_failed", http.StatusInternalServerError)
	}

	if (relayMode == relaymode.AudioTranscription || relayMode == relaymode.AudioSpeech) && channelType == channeltype.Azure {
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
	if relayMode == relaymode.AudioSpeech && channelType == channeltype.Minimax {
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", "application/json")
	}

	resp, err := client.HTTPClient.Do(req)
	if err != nil {
		return openai.ErrorWrapper(err, "do_request_failed", http.StatusInternalServerError)
	}

	err = req.Body.Close()
	if err != nil {
		return openai.ErrorWrapper(err, "close_request_body_failed", http.StatusInternalServerError)
	}
	err = c.Request.Body.Close()
	if err != nil {
		return openai.ErrorWrapper(err, "close_request_body_failed", http.StatusInternalServerError)
	}

	if relayMode != relaymode.AudioSpeech {
		responseBody, err := io.ReadAll(resp.Body)
		if err != nil {
			return openai.ErrorWrapper(err, "read_response_body_failed", http.StatusInternalServerError)
		}
		err = resp.Body.Close()
		if err != nil {
			return openai.ErrorWrapper(err, "close_response_body_failed", http.StatusInternalServerError)
		}

		var openAIErr openai.SlimTextResponse
		if err = json.Unmarshal(responseBody, &openAIErr); err == nil {
			if openAIErr.Error.Message != "" {
				return openai.ErrorWrapper(fmt.Errorf("type %s, code %v, message %s", openAIErr.Error.Type, openAIErr.Error.Code, openAIErr.Error.Message), "request_error", http.StatusInternalServerError)
			}
		}

		var text string
		switch responseFormat {
		case "json":
			text, err = getTextFromJSON(responseBody)
		case "text":
			text, err = getTextFromText(responseBody)
		case "srt":
			text, err = getTextFromSRT(responseBody)
		case "verbose_json":
			text, err = getTextFromVerboseJSON(responseBody)
		case "vtt":
			text, err = getTextFromVTT(responseBody)
		default:
			return openai.ErrorWrapper(errors.New("unexpected_response_format"), "unexpected_response_format", http.StatusInternalServerError)
		}
		if err != nil {
			return openai.ErrorWrapper(err, "get_text_from_body_err", http.StatusInternalServerError)
		}
		quota = int64(openai.CountTokenText(text, audioModel))
		resp.Body = io.NopCloser(bytes.NewBuffer(responseBody))
	}
	if resp.StatusCode != http.StatusOK {
		return RelayErrorHandler(resp)
	}
	for k, v := range resp.Header {
		c.Writer.Header().Set(k, v[0])
	}
	if relayMode == relaymode.AudioSpeech && channelType == channeltype.Minimax {
		responseBody, readErr := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		if readErr != nil {
			return openai.ErrorWrapper(readErr, "read_response_body_failed", http.StatusInternalServerError)
		}
		audio, decodeErr := decodeMinimaxSpeechResponse(responseBody)
		if decodeErr != nil {
			return openai.ErrorWrapper(decodeErr, "invalid_response", http.StatusBadGateway)
		}
		succeed = true
		quotaDelta := quota - preConsumedQuota
		defer func(ctx context.Context) {
			go billing.PostConsumeQuota(ctx, tokenId, quotaDelta, quota, userId, channelId, modelRatio, groupRatio, audioModel, tokenName)
		}(c.Request.Context())
		c.Writer.Header().Set("Content-Type", audioContentType(responseFormat))
		c.Writer.Header().Set("Content-Length", fmt.Sprintf("%d", len(audio)))
		c.Writer.WriteHeader(resp.StatusCode)
		_, err = c.Writer.Write(audio)
		if err != nil {
			return openai.ErrorWrapper(err, "copy_response_body_failed", http.StatusInternalServerError)
		}
		return nil
	}
	succeed = true
	quotaDelta := quota - preConsumedQuota
	defer func(ctx context.Context) {
		go billing.PostConsumeQuota(ctx, tokenId, quotaDelta, quota, userId, channelId, modelRatio, groupRatio, audioModel, tokenName)
	}(c.Request.Context())
	c.Writer.WriteHeader(resp.StatusCode)

	_, err = io.Copy(c.Writer, resp.Body)
	if err != nil {
		return openai.ErrorWrapper(err, "copy_response_body_failed", http.StatusInternalServerError)
	}
	err = resp.Body.Close()
	if err != nil {
		return openai.ErrorWrapper(err, "close_response_body_failed", http.StatusInternalServerError)
	}
	return nil
}

func getTextFromVTT(body []byte) (string, error) {
	return getTextFromSRT(body)
}

func getTextFromVerboseJSON(body []byte) (string, error) {
	var whisperResponse openai.WhisperVerboseJSONResponse
	if err := json.Unmarshal(body, &whisperResponse); err != nil {
		return "", fmt.Errorf("unmarshal_response_body_failed err :%w", err)
	}
	return whisperResponse.Text, nil
}

func getTextFromSRT(body []byte) (string, error) {
	scanner := bufio.NewScanner(strings.NewReader(string(body)))
	var builder strings.Builder
	var textLine bool
	for scanner.Scan() {
		line := scanner.Text()
		if textLine {
			builder.WriteString(line)
			textLine = false
			continue
		} else if strings.Contains(line, "-->") {
			textLine = true
			continue
		}
	}
	if err := scanner.Err(); err != nil {
		return "", err
	}
	return builder.String(), nil
}

func getTextFromText(body []byte) (string, error) {
	return strings.TrimSuffix(string(body), "\n"), nil
}

func getTextFromJSON(body []byte) (string, error) {
	var whisperResponse openai.WhisperJSONResponse
	if err := json.Unmarshal(body, &whisperResponse); err != nil {
		return "", fmt.Errorf("unmarshal_response_body_failed err :%w", err)
	}
	return whisperResponse.Text, nil
}
