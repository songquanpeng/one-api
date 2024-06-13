package controller

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/songquanpeng/one-api/common"
	"github.com/songquanpeng/one-api/common/audit"
	"github.com/songquanpeng/one-api/common/client"
	"github.com/songquanpeng/one-api/common/config"
	"github.com/songquanpeng/one-api/common/ctxkey"
	"github.com/songquanpeng/one-api/common/helper"
	"github.com/songquanpeng/one-api/relay/adaptor/openai"
	"github.com/songquanpeng/one-api/relay/billing"
	"github.com/songquanpeng/one-api/relay/channeltype"
	"github.com/songquanpeng/one-api/relay/meta"
	relaymodel "github.com/songquanpeng/one-api/relay/model"
	"github.com/songquanpeng/one-api/relay/relaymode"
)

func (rl *defaultRelay) RelayAudioHelper(c *gin.Context, relayMode int) *relaymodel.ErrorWithStatusCode {
	meta := meta.GetByContext(c)
	audioModel := "whisper-1"

	tokenId := c.GetInt(ctxkey.TokenId)
	channelType := c.GetInt(ctxkey.Channel)
	// channelId := c.GetInt(ctxkey.ChannelId)
	userId := c.GetInt(ctxkey.Id)
	group := c.GetString(ctxkey.Group)
	// tokenName := c.GetString(ctxkey.TokenName)

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

	var (
		modelRatio       float64
		groupRatio       float64
		ratio            float64
		quota            int64
		preConsumeQuota  int64
		preConsumedQuota int64
		bizErr           *relaymodel.ErrorWithStatusCode
	)

	if rl.Bookkeeper != nil {
		modelRatio = rl.ModelRatio(audioModel)
		groupRatio = rl.GroupRation(group)
		ratio = modelRatio * groupRatio
	}

	switch relayMode {
	// speech 类型，消费的配额直接根据输入的文本长度计算
	case relaymode.AudioSpeech:
		preConsumeQuota = int64(float64(len(ttsRequest.Input)) * ratio)
		quota = preConsumeQuota
	// 其他类型，假设消费的配额是预设的配额的 ratio 倍
	default:
		preConsumeQuota = int64(float64(config.PreConsumedQuota) * ratio)
	}
	if rl.Bookkeeper != nil {
		preConsumedQuota, bizErr = rl.PreConsumeQuota(c, preConsumeQuota, userId, tokenId)
		if bizErr != nil {
			return bizErr
		}
	}

	succeed := false
	defer func() {
		if succeed {
			return
		}
		if rl.Bookkeeper != nil {
			rl.Bookkeeper.RefundQuota(c.Request.Context(), preConsumedQuota, tokenId)
		}
	}()

	// map model name
	modelMapping := c.GetString(ctxkey.ModelMapping)
	if modelMapping != "" {
		modelMap := make(map[string]string)
		err := json.Unmarshal([]byte(modelMapping), &modelMap)
		if err != nil {
			return openai.ErrorWrapper(err, "unmarshal_model_mapping_failed", http.StatusInternalServerError)
		}
		if modelMap[audioModel] != "" {
			audioModel = modelMap[audioModel]
		}
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
	if _, err := io.Copy(requestBody, c.Request.Body); err != nil {
		return openai.ErrorWrapper(err, "new_request_body_failed", http.StatusInternalServerError)
	}
	c.Request.Body = io.NopCloser(bytes.NewBuffer(requestBody.Bytes()))
	responseFormat := c.DefaultPostForm("response_format", "json")

	req, err := http.NewRequest(c.Request.Method, fullRequestURL, requestBody)
	if err != nil {
		return openai.ErrorWrapper(err, "new_request_failed", http.StatusInternalServerError)
	}
	if config.UpstreamAuditEnabled {
		audit.Logger().
			WithField("stage", "upstream request").
			WithField("raw", audit.B64encode(requestBody.Bytes())).
			WithField("requestid", c.GetString(helper.RequestIdKey)).
			WithFields(meta.ToLogrusFields()).
			Info("upstream request")
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

	resp, err := client.HTTPClient.Do(req)
	if err != nil {
		return openai.ErrorWrapper(err, "do_request_failed", http.StatusInternalServerError)
	}
	if config.UpstreamAuditEnabled {
		buf := audit.CaptureHTTPResponseBody(resp)
		defer func() {
			audit.Logger().
				WithField("stage", "upstream response").
				WithField("raw", audit.B64encode(buf.Bytes())).
				WithField("requestid", c.GetString(helper.RequestIdKey)).
				WithFields(meta.ToLogrusFields()).
				Info("upstream response")
		}()
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
	succeed = true
	// quotaDelta := quota - preConsumedQuota
	defer func(ctx context.Context) {
		// go billing.PostConsumeQuota(ctx, tokenId, quotaDelta, quota, userId, channelId, modelRatio, groupRatio, audioModel, tokenName)
		// post-consume quota
		if rl.Bookkeeper != nil {
			// go postConsumeQuota(c, usage, meta, textRequest, ratio, preConsumedQuota, modelRatio, groupRatio)

			logContent := fmt.Sprintf("模型倍率 %.2f，分组倍率 %.2f", modelRatio, groupRatio)
			consumeLog := &billing.ConsumeLog{
				UserId:           meta.UserId,
				ChannelId:        meta.ChannelId,
				ModelName:        audioModel,
				TokenName:        c.GetString(ctxkey.TokenName),
				TokenId:          meta.TokenId,
				Quota:            quota,
				Content:          logContent,
				PromptTokens:     int(preConsumeQuota),
				CompletionTokens: 0,
				PreConsumedQuota: preConsumedQuota,
			}
			rl.Bookkeeper.Consume(c, consumeLog)
		}
	}(c.Request.Context())

	for k, v := range resp.Header {
		c.Writer.Header().Set(k, v[0])
	}
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
