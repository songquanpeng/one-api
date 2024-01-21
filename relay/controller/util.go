package controller

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"io"
	"net/http"
	"one-api/common"
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

func GetRequestURL(requestURL string, apiType int, relayMode int, meta *util.RelayMeta, textRequest *openai.GeneralOpenAIRequest) (string, error) {
	fullRequestURL := util.GetFullRequestURL(meta.BaseURL, requestURL, meta.ChannelType)
	switch apiType {
	case constant.APITypeOpenAI:
		if meta.ChannelType == common.ChannelTypeAzure {
			// https://learn.microsoft.com/en-us/azure/cognitive-services/openai/chatgpt-quickstart?pivots=rest-api&tabs=command-line#rest-api
			requestURL := strings.Split(requestURL, "?")[0]
			requestURL = fmt.Sprintf("%s?api-version=%s", requestURL, meta.APIVersion)
			task := strings.TrimPrefix(requestURL, "/v1/")
			model_ := textRequest.Model
			model_ = strings.Replace(model_, ".", "", -1)
			// https://github.com/songquanpeng/one-api/issues/67
			model_ = strings.TrimSuffix(model_, "-0301")
			model_ = strings.TrimSuffix(model_, "-0314")
			model_ = strings.TrimSuffix(model_, "-0613")

			requestURL = fmt.Sprintf("/openai/deployments/%s/%s", model_, task)
			fullRequestURL = util.GetFullRequestURL(meta.BaseURL, requestURL, meta.ChannelType)
		}
	case constant.APITypeClaude:
		fullRequestURL = fmt.Sprintf("%s/v1/complete", meta.BaseURL)
	case constant.APITypeBaidu:
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
		var accessToken string
		var err error
		if accessToken, err = baidu.GetAccessToken(meta.APIKey); err != nil {
			return "", fmt.Errorf("failed to get baidu access token: %w", err)
		}
		fullRequestURL += "?access_token=" + accessToken
	case constant.APITypePaLM:
		fullRequestURL = fmt.Sprintf("%s/v1beta2/models/chat-bison-001:generateMessage", meta.BaseURL)
	case constant.APITypeGemini:
		version := common.AssignOrDefault(meta.APIVersion, "v1")
		action := "generateContent"
		if textRequest.Stream {
			action = "streamGenerateContent"
		}
		fullRequestURL = fmt.Sprintf("%s/%s/models/%s:%s", meta.BaseURL, version, textRequest.Model, action)
	case constant.APITypeZhipu:
		method := "invoke"
		if textRequest.Stream {
			method = "sse-invoke"
		}
		fullRequestURL = fmt.Sprintf("https://open.bigmodel.cn/api/paas/v3/model-api/%s/%s", textRequest.Model, method)
	case constant.APITypeAli:
		fullRequestURL = "https://dashscope.aliyuncs.com/api/v1/services/aigc/text-generation/generation"
		if relayMode == constant.RelayModeEmbeddings {
			fullRequestURL = "https://dashscope.aliyuncs.com/api/v1/services/embeddings/text-embedding/text-embedding"
		}
	case constant.APITypeTencent:
		fullRequestURL = "https://hunyuan.cloud.tencent.com/hyllm/v1/chat/completions"
	case constant.APITypeAIProxyLibrary:
		fullRequestURL = fmt.Sprintf("%s/api/library/ask", meta.BaseURL)
	}
	return fullRequestURL, nil
}

func GetRequestBody(c *gin.Context, textRequest openai.GeneralOpenAIRequest, isModelMapped bool, apiType int, relayMode int) (io.Reader, error) {
	var requestBody io.Reader
	if isModelMapped {
		jsonStr, err := json.Marshal(textRequest)
		if err != nil {
			return nil, err
		}
		requestBody = bytes.NewBuffer(jsonStr)
	} else {
		requestBody = c.Request.Body
	}
	switch apiType {
	case constant.APITypeClaude:
		claudeRequest := anthropic.ConvertRequest(textRequest)
		jsonStr, err := json.Marshal(claudeRequest)
		if err != nil {
			return nil, err
		}
		requestBody = bytes.NewBuffer(jsonStr)
	case constant.APITypeBaidu:
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
			return nil, err
		}
		requestBody = bytes.NewBuffer(jsonData)
	case constant.APITypePaLM:
		palmRequest := google.ConvertPaLMRequest(textRequest)
		jsonStr, err := json.Marshal(palmRequest)
		if err != nil {
			return nil, err
		}
		requestBody = bytes.NewBuffer(jsonStr)
	case constant.APITypeGemini:
		geminiChatRequest := google.ConvertGeminiRequest(textRequest)
		jsonStr, err := json.Marshal(geminiChatRequest)
		if err != nil {
			return nil, err
		}
		requestBody = bytes.NewBuffer(jsonStr)
	case constant.APITypeZhipu:
		zhipuRequest := zhipu.ConvertRequest(textRequest)
		jsonStr, err := json.Marshal(zhipuRequest)
		if err != nil {
			return nil, err
		}
		requestBody = bytes.NewBuffer(jsonStr)
	case constant.APITypeAli:
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
			return nil, err
		}
		requestBody = bytes.NewBuffer(jsonStr)
	case constant.APITypeTencent:
		apiKey := c.Request.Header.Get("Authorization")
		apiKey = strings.TrimPrefix(apiKey, "Bearer ")
		appId, secretId, secretKey, err := tencent.ParseConfig(apiKey)
		if err != nil {
			return nil, err
		}
		tencentRequest := tencent.ConvertRequest(textRequest)
		tencentRequest.AppId = appId
		tencentRequest.SecretId = secretId
		jsonStr, err := json.Marshal(tencentRequest)
		if err != nil {
			return nil, err
		}
		sign := tencent.GetSign(*tencentRequest, secretKey)
		c.Request.Header.Set("Authorization", sign)
		requestBody = bytes.NewBuffer(jsonStr)
	case constant.APITypeAIProxyLibrary:
		aiProxyLibraryRequest := aiproxy.ConvertRequest(textRequest)
		aiProxyLibraryRequest.LibraryId = c.GetString("library_id")
		jsonStr, err := json.Marshal(aiProxyLibraryRequest)
		if err != nil {
			return nil, err
		}
		requestBody = bytes.NewBuffer(jsonStr)
	}
	return requestBody, nil
}

func SetupRequestHeaders(c *gin.Context, req *http.Request, apiType int, meta *util.RelayMeta, isStream bool) {
	SetupAuthHeaders(c, req, apiType, meta, isStream)
	req.Header.Set("Content-Type", c.Request.Header.Get("Content-Type"))
	req.Header.Set("Accept", c.Request.Header.Get("Accept"))
	if isStream && c.Request.Header.Get("Accept") == "" {
		req.Header.Set("Accept", "text/event-stream")
	}
}

func SetupAuthHeaders(c *gin.Context, req *http.Request, apiType int, meta *util.RelayMeta, isStream bool) {
	apiKey := meta.APIKey
	switch apiType {
	case constant.APITypeOpenAI:
		if meta.ChannelType == common.ChannelTypeAzure {
			req.Header.Set("api-key", apiKey)
		} else {
			req.Header.Set("Authorization", c.Request.Header.Get("Authorization"))
			if meta.ChannelType == common.ChannelTypeOpenRouter {
				req.Header.Set("HTTP-Referer", "https://github.com/songquanpeng/one-api")
				req.Header.Set("X-Title", "One API")
			}
		}
	case constant.APITypeClaude:
		req.Header.Set("x-api-key", apiKey)
		anthropicVersion := c.Request.Header.Get("anthropic-version")
		if anthropicVersion == "" {
			anthropicVersion = "2023-06-01"
		}
		req.Header.Set("anthropic-version", anthropicVersion)
	case constant.APITypeZhipu:
		token := zhipu.GetToken(apiKey)
		req.Header.Set("Authorization", token)
	case constant.APITypeAli:
		req.Header.Set("Authorization", "Bearer "+apiKey)
		if isStream {
			req.Header.Set("X-DashScope-SSE", "enable")
		}
		if c.GetString("plugin") != "" {
			req.Header.Set("X-DashScope-Plugin", c.GetString("plugin"))
		}
	case constant.APITypeTencent:
		req.Header.Set("Authorization", apiKey)
	case constant.APITypePaLM:
		req.Header.Set("x-goog-api-key", apiKey)
	case constant.APITypeGemini:
		req.Header.Set("x-goog-api-key", apiKey)
	default:
		req.Header.Set("Authorization", "Bearer "+apiKey)
	}
}

func DoResponse(c *gin.Context, textRequest *openai.GeneralOpenAIRequest, resp *http.Response, relayMode int, apiType int, isStream bool, promptTokens int) (usage *openai.Usage, err *openai.ErrorWithStatusCode) {
	var responseText string
	switch apiType {
	case constant.APITypeOpenAI:
		if isStream {
			err, responseText = openai.StreamHandler(c, resp, relayMode)
		} else {
			err, usage = openai.Handler(c, resp, promptTokens, textRequest.Model)
		}
	case constant.APITypeClaude:
		if isStream {
			err, responseText = anthropic.StreamHandler(c, resp)
		} else {
			err, usage = anthropic.Handler(c, resp, promptTokens, textRequest.Model)
		}
	case constant.APITypeBaidu:
		if isStream {
			err, usage = baidu.StreamHandler(c, resp)
		} else {
			switch relayMode {
			case constant.RelayModeEmbeddings:
				err, usage = baidu.EmbeddingHandler(c, resp)
			default:
				err, usage = baidu.Handler(c, resp)
			}
		}
	case constant.APITypePaLM:
		if isStream { // PaLM2 API does not support stream
			err, responseText = google.PaLMStreamHandler(c, resp)
		} else {
			err, usage = google.PaLMHandler(c, resp, promptTokens, textRequest.Model)
		}
	case constant.APITypeGemini:
		if isStream {
			err, responseText = google.StreamHandler(c, resp)
		} else {
			err, usage = google.GeminiHandler(c, resp, promptTokens, textRequest.Model)
		}
	case constant.APITypeZhipu:
		if isStream {
			err, usage = zhipu.StreamHandler(c, resp)
		} else {
			err, usage = zhipu.Handler(c, resp)
		}
	case constant.APITypeAli:
		if isStream {
			err, usage = ali.StreamHandler(c, resp)
		} else {
			switch relayMode {
			case constant.RelayModeEmbeddings:
				err, usage = ali.EmbeddingHandler(c, resp)
			default:
				err, usage = ali.Handler(c, resp)
			}
		}
	case constant.APITypeXunfei:
		auth := c.Request.Header.Get("Authorization")
		auth = strings.TrimPrefix(auth, "Bearer ")
		splits := strings.Split(auth, "|")
		if len(splits) != 3 {
			return nil, openai.ErrorWrapper(errors.New("invalid auth"), "invalid_auth", http.StatusBadRequest)
		}
		if isStream {
			err, usage = xunfei.StreamHandler(c, *textRequest, splits[0], splits[1], splits[2])
		} else {
			err, usage = xunfei.Handler(c, *textRequest, splits[0], splits[1], splits[2])
		}
	case constant.APITypeAIProxyLibrary:
		if isStream {
			err, usage = aiproxy.StreamHandler(c, resp)
		} else {
			err, usage = aiproxy.Handler(c, resp)
		}
	case constant.APITypeTencent:
		if isStream {
			err, responseText = tencent.StreamHandler(c, resp)
		} else {
			err, usage = tencent.Handler(c, resp)
		}
	default:
		return nil, openai.ErrorWrapper(errors.New("unknown api type"), "unknown_api_type", http.StatusInternalServerError)
	}
	if err != nil {
		return nil, err
	}
	if usage == nil && responseText != "" {
		usage = &openai.Usage{}
		usage.PromptTokens = promptTokens
		usage.CompletionTokens = openai.CountTokenText(responseText, textRequest.Model)
		usage.TotalTokens = usage.PromptTokens + usage.CompletionTokens
	}
	return usage, nil
}
