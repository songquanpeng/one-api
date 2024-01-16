package controller

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"io"
	"math"
	"net/http"
	"one-api/common"
	"one-api/model"
	"one-api/relay/channel/aiproxy"
	"one-api/relay/channel/ali"
	"one-api/relay/channel/anthropic"
	"one-api/relay/channel/baidu"
	"one-api/relay/channel/google"
	"one-api/relay/channel/openai"
	"one-api/relay/channel/tencent"
	"one-api/relay/channel/xunfei"
	"one-api/relay/channel/zhipu"
	"one-api/relay/constant"
	"one-api/relay/util"
	"strings"
)

const (
	APITypeOpenAI = iota
	APITypeClaude
	APITypePaLM
	APITypeBaidu
	APITypeZhipu
	APITypeAli
	APITypeXunfei
	APITypeAIProxyLibrary
	APITypeTencent
	APITypeGemini
)

func RelayTextHelper(c *gin.Context, relayMode int) *openai.ErrorWithStatusCode {
	channelType := c.GetInt("channel")
	channelId := c.GetInt("channel_id")
	tokenId := c.GetInt("token_id")
	userId := c.GetInt("id")
	group := c.GetString("group")
	var textRequest openai.GeneralOpenAIRequest
	err := common.UnmarshalBodyReusable(c, &textRequest)
	if err != nil {
		return openai.ErrorWrapper(err, "bind_request_body_failed", http.StatusBadRequest)
	}
	if textRequest.MaxTokens < 0 || textRequest.MaxTokens > math.MaxInt32/2 {
		return openai.ErrorWrapper(errors.New("max_tokens is invalid"), "invalid_max_tokens", http.StatusBadRequest)
	}
	if relayMode == constant.RelayModeModerations && textRequest.Model == "" {
		textRequest.Model = "text-moderation-latest"
	}
	if relayMode == constant.RelayModeEmbeddings && textRequest.Model == "" {
		textRequest.Model = c.Param("model")
	}
	// request validation
	if textRequest.Model == "" {
		return openai.ErrorWrapper(errors.New("model is required"), "required_field_missing", http.StatusBadRequest)
	}
	switch relayMode {
	case constant.RelayModeCompletions:
		if textRequest.Prompt == "" {
			return openai.ErrorWrapper(errors.New("field prompt is required"), "required_field_missing", http.StatusBadRequest)
		}
	case constant.RelayModeChatCompletions:
		if textRequest.Messages == nil || len(textRequest.Messages) == 0 {
			return openai.ErrorWrapper(errors.New("field messages is required"), "required_field_missing", http.StatusBadRequest)
		}
	case constant.RelayModeEmbeddings:
	case constant.RelayModeModerations:
		if textRequest.Input == "" {
			return openai.ErrorWrapper(errors.New("field input is required"), "required_field_missing", http.StatusBadRequest)
		}
	case constant.RelayModeEdits:
		if textRequest.Instruction == "" {
			return openai.ErrorWrapper(errors.New("field instruction is required"), "required_field_missing", http.StatusBadRequest)
		}
	}
	// map model name
	modelMapping := c.GetString("model_mapping")
	isModelMapped := false
	if modelMapping != "" && modelMapping != "{}" {
		modelMap := make(map[string]string)
		err := json.Unmarshal([]byte(modelMapping), &modelMap)
		if err != nil {
			return openai.ErrorWrapper(err, "unmarshal_model_mapping_failed", http.StatusInternalServerError)
		}
		if modelMap[textRequest.Model] != "" {
			textRequest.Model = modelMap[textRequest.Model]
			isModelMapped = true
		}
	}
	apiType := APITypeOpenAI
	switch channelType {
	case common.ChannelTypeAnthropic:
		apiType = APITypeClaude
	case common.ChannelTypeBaidu:
		apiType = APITypeBaidu
	case common.ChannelTypePaLM:
		apiType = APITypePaLM
	case common.ChannelTypeZhipu:
		apiType = APITypeZhipu
	case common.ChannelTypeAli:
		apiType = APITypeAli
	case common.ChannelTypeXunfei:
		apiType = APITypeXunfei
	case common.ChannelTypeAIProxyLibrary:
		apiType = APITypeAIProxyLibrary
	case common.ChannelTypeTencent:
		apiType = APITypeTencent
	case common.ChannelTypeGemini:
		apiType = APITypeGemini
	}
	baseURL := common.ChannelBaseURLs[channelType]
	requestURL := c.Request.URL.String()
	if c.GetString("base_url") != "" {
		baseURL = c.GetString("base_url")
	}
	fullRequestURL := util.GetFullRequestURL(baseURL, requestURL, channelType)
	switch apiType {
	case APITypeOpenAI:
		if channelType == common.ChannelTypeAzure {
			// https://learn.microsoft.com/en-us/azure/cognitive-services/openai/chatgpt-quickstart?pivots=rest-api&tabs=command-line#rest-api
			apiVersion := util.GetAPIVersion(c)
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

			requestURL = fmt.Sprintf("/openai/deployments/%s/%s", model_, task)
			fullRequestURL = util.GetFullRequestURL(baseURL, requestURL, channelType)
		}
	case APITypeClaude:
		fullRequestURL = "https://api.anthropic.com/v1/complete"
		if baseURL != "" {
			fullRequestURL = fmt.Sprintf("%s/v1/complete", baseURL)
		}
	case APITypeBaidu:
		switch textRequest.Model {
		case "ERNIE-Bot":
			fullRequestURL = "https://aip.baidubce.com/rpc/2.0/ai_custom/v1/wenxinworkshop/chat/completions"
		case "ERNIE-Bot-turbo":
			fullRequestURL = "https://aip.baidubce.com/rpc/2.0/ai_custom/v1/wenxinworkshop/chat/eb-instant"
		case "ERNIE-Bot-4":
			fullRequestURL = "https://aip.baidubce.com/rpc/2.0/ai_custom/v1/wenxinworkshop/chat/completions_pro"
		case "BLOOMZ-7B":
			fullRequestURL = "https://aip.baidubce.com/rpc/2.0/ai_custom/v1/wenxinworkshop/chat/bloomz_7b1"
		case "Embedding-V1":
			fullRequestURL = "https://aip.baidubce.com/rpc/2.0/ai_custom/v1/wenxinworkshop/embeddings/embedding-v1"
		}
		apiKey := c.Request.Header.Get("Authorization")
		apiKey = strings.TrimPrefix(apiKey, "Bearer ")
		var err error
		if apiKey, err = baidu.GetAccessToken(apiKey); err != nil {
			return openai.ErrorWrapper(err, "invalid_baidu_config", http.StatusInternalServerError)
		}
		fullRequestURL += "?access_token=" + apiKey
	case APITypePaLM:
		fullRequestURL = "https://generativelanguage.googleapis.com/v1beta2/models/chat-bison-001:generateMessage"
		if baseURL != "" {
			fullRequestURL = fmt.Sprintf("%s/v1beta2/models/chat-bison-001:generateMessage", baseURL)
		}
	case APITypeGemini:
		requestBaseURL := "https://generativelanguage.googleapis.com"
		if baseURL != "" {
			requestBaseURL = baseURL
		}
		version := "v1"
		if c.GetString("api_version") != "" {
			version = c.GetString("api_version")
		}
		action := "generateContent"
		if textRequest.Stream {
			action = "streamGenerateContent"
		}
		fullRequestURL = fmt.Sprintf("%s/%s/models/%s:%s", requestBaseURL, version, textRequest.Model, action)
	case APITypeZhipu:
		method := "invoke"
		if textRequest.Stream {
			method = "sse-invoke"
		}
		fullRequestURL = fmt.Sprintf("https://open.bigmodel.cn/api/paas/v4/model-api/%s/%s", textRequest.Model, method)
	case APITypeAli:
		fullRequestURL = "https://dashscope.aliyuncs.com/api/v1/services/aigc/text-generation/generation"
		if relayMode == constant.RelayModeEmbeddings {
			fullRequestURL = "https://dashscope.aliyuncs.com/api/v1/services/embeddings/text-embedding/text-embedding"
		}
	case APITypeTencent:
		fullRequestURL = "https://hunyuan.cloud.tencent.com/hyllm/v1/chat/completions"
	case APITypeAIProxyLibrary:
		fullRequestURL = fmt.Sprintf("%s/api/library/ask", baseURL)
	}
	var promptTokens int
	var completionTokens int
	switch relayMode {
	case constant.RelayModeChatCompletions:
		promptTokens = openai.CountTokenMessages(textRequest.Messages, textRequest.Model)
	case constant.RelayModeCompletions:
		promptTokens = openai.CountTokenInput(textRequest.Prompt, textRequest.Model)
	case constant.RelayModeModerations:
		promptTokens = openai.CountTokenInput(textRequest.Input, textRequest.Model)
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
		return openai.ErrorWrapper(err, "get_user_quota_failed", http.StatusInternalServerError)
	}
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
		common.LogInfo(c.Request.Context(), fmt.Sprintf("user %d has enough quota %d, trusted and no need to pre-consume", userId, userQuota))
	}
	if preConsumedQuota > 0 {
		err := model.PreConsumeTokenQuota(tokenId, preConsumedQuota)
		if err != nil {
			return openai.ErrorWrapper(err, "pre_consume_token_quota_failed", http.StatusForbidden)
		}
	}
	var requestBody io.Reader
	if isModelMapped {
		jsonStr, err := json.Marshal(textRequest)
		if err != nil {
			return openai.ErrorWrapper(err, "marshal_text_request_failed", http.StatusInternalServerError)
		}
		requestBody = bytes.NewBuffer(jsonStr)
	} else {
		requestBody = c.Request.Body
	}
	switch apiType {
	case APITypeClaude:
		claudeRequest := anthropic.ConvertRequest(textRequest)
		jsonStr, err := json.Marshal(claudeRequest)
		if err != nil {
			return openai.ErrorWrapper(err, "marshal_text_request_failed", http.StatusInternalServerError)
		}
		requestBody = bytes.NewBuffer(jsonStr)
	case APITypeBaidu:
		var jsonData []byte
		var err error
		switch relayMode {
		case constant.RelayModeEmbeddings:
			baiduEmbeddingRequest := baidu.ConvertEmbeddingRequest(textRequest)
			jsonData, err = json.Marshal(baiduEmbeddingRequest)
		default:
			baiduRequest := baidu.ConvertRequest(textRequest)
			jsonData, err = json.Marshal(baiduRequest)
		}
		if err != nil {
			return openai.ErrorWrapper(err, "marshal_text_request_failed", http.StatusInternalServerError)
		}
		requestBody = bytes.NewBuffer(jsonData)
	case APITypePaLM:
		palmRequest := google.ConvertPaLMRequest(textRequest)
		jsonStr, err := json.Marshal(palmRequest)
		if err != nil {
			return openai.ErrorWrapper(err, "marshal_text_request_failed", http.StatusInternalServerError)
		}
		requestBody = bytes.NewBuffer(jsonStr)
	case APITypeGemini:
		geminiChatRequest := google.ConvertGeminiRequest(textRequest)
		jsonStr, err := json.Marshal(geminiChatRequest)
		if err != nil {
			return openai.ErrorWrapper(err, "marshal_text_request_failed", http.StatusInternalServerError)
		}
		requestBody = bytes.NewBuffer(jsonStr)
	case APITypeZhipu:
		zhipuRequest := zhipu.ConvertRequest(textRequest)
		jsonStr, err := json.Marshal(zhipuRequest)
		if err != nil {
			return openai.ErrorWrapper(err, "marshal_text_request_failed", http.StatusInternalServerError)
		}
		requestBody = bytes.NewBuffer(jsonStr)
	case APITypeAli:
		var jsonStr []byte
		var err error
		switch relayMode {
		case constant.RelayModeEmbeddings:
			aliEmbeddingRequest := ali.ConvertEmbeddingRequest(textRequest)
			jsonStr, err = json.Marshal(aliEmbeddingRequest)
		default:
			aliRequest := ali.ConvertRequest(textRequest)
			jsonStr, err = json.Marshal(aliRequest)
		}
		if err != nil {
			return openai.ErrorWrapper(err, "marshal_text_request_failed", http.StatusInternalServerError)
		}
		requestBody = bytes.NewBuffer(jsonStr)
	case APITypeTencent:
		apiKey := c.Request.Header.Get("Authorization")
		apiKey = strings.TrimPrefix(apiKey, "Bearer ")
		appId, secretId, secretKey, err := tencent.ParseConfig(apiKey)
		if err != nil {
			return openai.ErrorWrapper(err, "invalid_tencent_config", http.StatusInternalServerError)
		}
		tencentRequest := tencent.ConvertRequest(textRequest)
		tencentRequest.AppId = appId
		tencentRequest.SecretId = secretId
		jsonStr, err := json.Marshal(tencentRequest)
		if err != nil {
			return openai.ErrorWrapper(err, "marshal_text_request_failed", http.StatusInternalServerError)
		}
		sign := tencent.GetSign(*tencentRequest, secretKey)
		c.Request.Header.Set("Authorization", sign)
		requestBody = bytes.NewBuffer(jsonStr)
	case APITypeAIProxyLibrary:
		aiProxyLibraryRequest := aiproxy.ConvertRequest(textRequest)
		aiProxyLibraryRequest.LibraryId = c.GetString("library_id")
		jsonStr, err := json.Marshal(aiProxyLibraryRequest)
		if err != nil {
			return openai.ErrorWrapper(err, "marshal_text_request_failed", http.StatusInternalServerError)
		}
		requestBody = bytes.NewBuffer(jsonStr)
	}

	var req *http.Request
	var resp *http.Response
	isStream := textRequest.Stream

	if apiType != APITypeXunfei { // cause xunfei use websocket
		req, err = http.NewRequest(c.Request.Method, fullRequestURL, requestBody)
		if err != nil {
			return openai.ErrorWrapper(err, "new_request_failed", http.StatusInternalServerError)
		}
		apiKey := c.Request.Header.Get("Authorization")
		apiKey = strings.TrimPrefix(apiKey, "Bearer ")
		switch apiType {
		case APITypeOpenAI:
			if channelType == common.ChannelTypeAzure {
				req.Header.Set("api-key", apiKey)
			} else {
				req.Header.Set("Authorization", c.Request.Header.Get("Authorization"))
				if channelType == common.ChannelTypeOpenRouter {
					req.Header.Set("HTTP-Referer", "https://github.com/songquanpeng/one-api")
					req.Header.Set("X-Title", "One API")
				}
			}
		case APITypeClaude:
			req.Header.Set("x-api-key", apiKey)
			anthropicVersion := c.Request.Header.Get("anthropic-version")
			if anthropicVersion == "" {
				anthropicVersion = "2023-06-01"
			}
			req.Header.Set("anthropic-version", anthropicVersion)
		case APITypeZhipu:
			token := zhipu.GetToken(apiKey)
			req.Header.Set("Authorization", token)
		case APITypeAli:
			req.Header.Set("Authorization", "Bearer "+apiKey)
			if textRequest.Stream {
				req.Header.Set("X-DashScope-SSE", "enable")
			}
			if c.GetString("plugin") != "" {
				req.Header.Set("X-DashScope-Plugin", c.GetString("plugin"))
			}
		case APITypeTencent:
			req.Header.Set("Authorization", apiKey)
		case APITypePaLM:
			req.Header.Set("x-goog-api-key", apiKey)
		case APITypeGemini:
			req.Header.Set("x-goog-api-key", apiKey)
		default:
			req.Header.Set("Authorization", "Bearer "+apiKey)
		}
		req.Header.Set("Content-Type", c.Request.Header.Get("Content-Type"))
		req.Header.Set("Accept", c.Request.Header.Get("Accept"))
		if isStream && c.Request.Header.Get("Accept") == "" {
			req.Header.Set("Accept", "text/event-stream")
		}
		//req.Header.Set("Connection", c.Request.Header.Get("Connection"))
		resp, err = util.HTTPClient.Do(req)
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
		isStream = isStream || strings.HasPrefix(resp.Header.Get("Content-Type"), "text/event-stream")

		if resp.StatusCode != http.StatusOK {
			if preConsumedQuota != 0 {
				go func(ctx context.Context) {
					// return pre-consumed quota
					err := model.PostConsumeTokenQuota(tokenId, -preConsumedQuota)
					if err != nil {
						common.LogError(ctx, "error return pre-consumed quota: "+err.Error())
					}
				}(c.Request.Context())
			}
			return util.RelayErrorHandler(resp)
		}
	}

	var textResponse openai.SlimTextResponse
	tokenName := c.GetString("token_name")

	defer func(ctx context.Context) {
		// c.Writer.Flush()
		go func() {
			quota := 0
			completionRatio := common.GetCompletionRatio(textRequest.Model)
			promptTokens = textResponse.Usage.PromptTokens
			completionTokens = textResponse.Usage.CompletionTokens
			quota = int(math.Ceil((float64(promptTokens) + float64(completionTokens)*completionRatio) * ratio))
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
				common.LogError(ctx, "error consuming token remain quota: "+err.Error())
			}
			err = model.CacheUpdateUserQuota(userId)
			if err != nil {
				common.LogError(ctx, "error update user quota cache: "+err.Error())
			}
			if quota != 0 {
				logContent := fmt.Sprintf("模型倍率 %.2f，分组倍率 %.2f", modelRatio, groupRatio)
				model.RecordConsumeLog(ctx, userId, channelId, promptTokens, completionTokens, textRequest.Model, tokenName, quota, logContent)
				model.UpdateUserUsedQuotaAndRequestCount(userId, quota)
				model.UpdateChannelUsedQuota(channelId, quota)
			}

		}()
	}(c.Request.Context())
	switch apiType {
	case APITypeOpenAI:
		if isStream {
			err, responseText := openai.StreamHandler(c, resp, relayMode)
			if err != nil {
				return err
			}
			textResponse.Usage.PromptTokens = promptTokens
			textResponse.Usage.CompletionTokens = openai.CountTokenText(responseText, textRequest.Model)
			return nil
		} else {
			err, usage := openai.Handler(c, resp, promptTokens, textRequest.Model)
			if err != nil {
				return err
			}
			if usage != nil {
				textResponse.Usage = *usage
			}
			return nil
		}
	case APITypeClaude:
		if isStream {
			err, responseText := anthropic.StreamHandler(c, resp)
			if err != nil {
				return err
			}
			textResponse.Usage.PromptTokens = promptTokens
			textResponse.Usage.CompletionTokens = openai.CountTokenText(responseText, textRequest.Model)
			return nil
		} else {
			err, usage := anthropic.Handler(c, resp, promptTokens, textRequest.Model)
			if err != nil {
				return err
			}
			if usage != nil {
				textResponse.Usage = *usage
			}
			return nil
		}
	case APITypeBaidu:
		if isStream {
			err, usage := baidu.StreamHandler(c, resp)
			if err != nil {
				return err
			}
			if usage != nil {
				textResponse.Usage = *usage
			}
			return nil
		} else {
			var err *openai.ErrorWithStatusCode
			var usage *openai.Usage
			switch relayMode {
			case constant.RelayModeEmbeddings:
				err, usage = baidu.EmbeddingHandler(c, resp)
			default:
				err, usage = baidu.Handler(c, resp)
			}
			if err != nil {
				return err
			}
			if usage != nil {
				textResponse.Usage = *usage
			}
			return nil
		}
	case APITypePaLM:
		if textRequest.Stream { // PaLM2 API does not support stream
			err, responseText := google.PaLMStreamHandler(c, resp)
			if err != nil {
				return err
			}
			textResponse.Usage.PromptTokens = promptTokens
			textResponse.Usage.CompletionTokens = openai.CountTokenText(responseText, textRequest.Model)
			return nil
		} else {
			err, usage := google.PaLMHandler(c, resp, promptTokens, textRequest.Model)
			if err != nil {
				return err
			}
			if usage != nil {
				textResponse.Usage = *usage
			}
			return nil
		}
	case APITypeGemini:
		if textRequest.Stream {
			err, responseText := google.StreamHandler(c, resp)
			if err != nil {
				return err
			}
			textResponse.Usage.PromptTokens = promptTokens
			textResponse.Usage.CompletionTokens = openai.CountTokenText(responseText, textRequest.Model)
			return nil
		} else {
			err, usage := google.GeminiHandler(c, resp, promptTokens, textRequest.Model)
			if err != nil {
				return err
			}
			if usage != nil {
				textResponse.Usage = *usage
			}
			return nil
		}
	case APITypeZhipu:
		if isStream {
			err, usage := zhipu.StreamHandler(c, resp)
			if err != nil {
				return err
			}
			if usage != nil {
				textResponse.Usage = *usage
			}
			// zhipu's API does not return prompt tokens & completion tokens
			textResponse.Usage.PromptTokens = textResponse.Usage.TotalTokens
			return nil
		} else {
			err, usage := zhipu.Handler(c, resp)
			if err != nil {
				return err
			}
			if usage != nil {
				textResponse.Usage = *usage
			}
			// zhipu's API does not return prompt tokens & completion tokens
			textResponse.Usage.PromptTokens = textResponse.Usage.TotalTokens
			return nil
		}
	case APITypeAli:
		if isStream {
			err, usage := ali.StreamHandler(c, resp)
			if err != nil {
				return err
			}
			if usage != nil {
				textResponse.Usage = *usage
			}
			return nil
		} else {
			var err *openai.ErrorWithStatusCode
			var usage *openai.Usage
			switch relayMode {
			case constant.RelayModeEmbeddings:
				err, usage = ali.EmbeddingHandler(c, resp)
			default:
				err, usage = ali.Handler(c, resp)
			}
			if err != nil {
				return err
			}
			if usage != nil {
				textResponse.Usage = *usage
			}
			return nil
		}
	case APITypeXunfei:
		auth := c.Request.Header.Get("Authorization")
		auth = strings.TrimPrefix(auth, "Bearer ")
		splits := strings.Split(auth, "|")
		if len(splits) != 3 {
			return openai.ErrorWrapper(errors.New("invalid auth"), "invalid_auth", http.StatusBadRequest)
		}
		var err *openai.ErrorWithStatusCode
		var usage *openai.Usage
		if isStream {
			err, usage = xunfei.StreamHandler(c, textRequest, splits[0], splits[1], splits[2])
		} else {
			err, usage = xunfei.Handler(c, textRequest, splits[0], splits[1], splits[2])
		}
		if err != nil {
			return err
		}
		if usage != nil {
			textResponse.Usage = *usage
		}
		return nil
	case APITypeAIProxyLibrary:
		if isStream {
			err, usage := aiproxy.StreamHandler(c, resp)
			if err != nil {
				return err
			}
			if usage != nil {
				textResponse.Usage = *usage
			}
			return nil
		} else {
			err, usage := aiproxy.Handler(c, resp)
			if err != nil {
				return err
			}
			if usage != nil {
				textResponse.Usage = *usage
			}
			return nil
		}
	case APITypeTencent:
		if isStream {
			err, responseText := tencent.StreamHandler(c, resp)
			if err != nil {
				return err
			}
			textResponse.Usage.PromptTokens = promptTokens
			textResponse.Usage.CompletionTokens = openai.CountTokenText(responseText, textRequest.Model)
			return nil
		} else {
			err, usage := tencent.Handler(c, resp)
			if err != nil {
				return err
			}
			if usage != nil {
				textResponse.Usage = *usage
			}
			return nil
		}
	default:
		return openai.ErrorWrapper(errors.New("unknown api type"), "unknown_api_type", http.StatusInternalServerError)
	}
}
