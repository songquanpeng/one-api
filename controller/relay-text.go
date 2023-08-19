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
	"one-api/model"
	"strings"
	"time"
)

const (
	APITypeOpenAI = iota
	APITypeClaude
	APITypePaLM
	APITypeBaidu
	APITypeZhipu
	APITypeAli
	APITypeXunfei
)

var httpClient *http.Client
var impatientHTTPClient *http.Client

func init() {
	httpClient = &http.Client{}
	impatientHTTPClient = &http.Client{
		Timeout: 5 * time.Second,
	}
}

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
	if modelMapping != "" && modelMapping != "{}" {
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
	case APITypeBaidu:
		switch textRequest.Model {
		case "ERNIE-Bot":
			fullRequestURL = "https://aip.baidubce.com/rpc/2.0/ai_custom/v1/wenxinworkshop/chat/completions"
		case "ERNIE-Bot-turbo":
			fullRequestURL = "https://aip.baidubce.com/rpc/2.0/ai_custom/v1/wenxinworkshop/chat/eb-instant"
		case "BLOOMZ-7B":
			fullRequestURL = "https://aip.baidubce.com/rpc/2.0/ai_custom/v1/wenxinworkshop/chat/bloomz_7b1"
		case "Embedding-V1":
			fullRequestURL = "https://aip.baidubce.com/rpc/2.0/ai_custom/v1/wenxinworkshop/embeddings/embedding-v1"
		}
		apiKey := c.Request.Header.Get("Authorization")
		apiKey = strings.TrimPrefix(apiKey, "Bearer ")
		var err error
		if apiKey, err = getBaiduAccessToken(apiKey); err != nil {
			return errorWrapper(err, "invalid_baidu_config", http.StatusInternalServerError)
		}
		fullRequestURL += "?access_token=" + apiKey
	case APITypePaLM:
		fullRequestURL = "https://generativelanguage.googleapis.com/v1beta2/models/chat-bison-001:generateMessage"
		if baseURL != "" {
			fullRequestURL = fmt.Sprintf("%s/v1beta2/models/chat-bison-001:generateMessage", baseURL)
		}
		apiKey := c.Request.Header.Get("Authorization")
		apiKey = strings.TrimPrefix(apiKey, "Bearer ")
		fullRequestURL += "?key=" + apiKey
	case APITypeZhipu:
		method := "invoke"
		if textRequest.Stream {
			method = "sse-invoke"
		}
		fullRequestURL = fmt.Sprintf("https://open.bigmodel.cn/api/paas/v3/model-api/%s/%s", textRequest.Model, method)
	case APITypeAli:
		fullRequestURL = "https://dashscope.aliyuncs.com/api/v1/services/aigc/text-generation/generation"
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
	err = model.CacheDecreaseUserQuota(userId, preConsumedQuota)
	if err != nil {
		return errorWrapper(err, "decrease_user_quota_failed", http.StatusInternalServerError)
	}
	if userQuota > 100*preConsumedQuota {
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
	case APITypeBaidu:
		var jsonData []byte
		var err error
		switch relayMode {
		case RelayModeEmbeddings:
			baiduEmbeddingRequest := embeddingRequestOpenAI2Baidu(textRequest)
			jsonData, err = json.Marshal(baiduEmbeddingRequest)
		default:
			baiduRequest := requestOpenAI2Baidu(textRequest)
			jsonData, err = json.Marshal(baiduRequest)
		}
		if err != nil {
			return errorWrapper(err, "marshal_text_request_failed", http.StatusInternalServerError)
		}
		requestBody = bytes.NewBuffer(jsonData)
	case APITypePaLM:
		palmRequest := requestOpenAI2PaLM(textRequest)
		jsonStr, err := json.Marshal(palmRequest)
		if err != nil {
			return errorWrapper(err, "marshal_text_request_failed", http.StatusInternalServerError)
		}
		requestBody = bytes.NewBuffer(jsonStr)
	case APITypeZhipu:
		zhipuRequest := requestOpenAI2Zhipu(textRequest)
		jsonStr, err := json.Marshal(zhipuRequest)
		if err != nil {
			return errorWrapper(err, "marshal_text_request_failed", http.StatusInternalServerError)
		}
		requestBody = bytes.NewBuffer(jsonStr)
	case APITypeAli:
		aliRequest := requestOpenAI2Ali(textRequest)
		jsonStr, err := json.Marshal(aliRequest)
		if err != nil {
			return errorWrapper(err, "marshal_text_request_failed", http.StatusInternalServerError)
		}
		requestBody = bytes.NewBuffer(jsonStr)
	}

	var req *http.Request
	var resp *http.Response
	isStream := textRequest.Stream

	if apiType != APITypeXunfei { // cause xunfei use websocket
		req, err = http.NewRequest(c.Request.Method, fullRequestURL, requestBody)
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
		case APITypeZhipu:
			token := getZhipuToken(apiKey)
			req.Header.Set("Authorization", token)
		case APITypeAli:
			req.Header.Set("Authorization", "Bearer "+apiKey)
			if textRequest.Stream {
				req.Header.Set("X-DashScope-SSE", "enable")
			}
		}
		req.Header.Set("Content-Type", c.Request.Header.Get("Content-Type"))
		req.Header.Set("Accept", c.Request.Header.Get("Accept"))
		//req.Header.Set("Connection", c.Request.Header.Get("Connection"))
		resp, err = httpClient.Do(req)
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
		isStream = isStream || strings.HasPrefix(resp.Header.Get("Content-Type"), "text/event-stream")
	}

	var textResponse TextResponse
	tokenName := c.GetString("token_name")
	channelId := c.GetInt("channel_id")

	defer func() {
		// c.Writer.Flush()
		go func() {
			if consumeQuota {
				quota := 0
				completionRatio := common.GetCompletionRatio(textRequest.Model)
				promptTokens = textResponse.Usage.PromptTokens
				completionTokens = textResponse.Usage.CompletionTokens

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
					logContent := fmt.Sprintf("模型倍率 %.2f，分组倍率 %.2f", modelRatio, groupRatio)
					model.RecordConsumeLog(userId, promptTokens, completionTokens, textRequest.Model, tokenName, quota, logContent)
					model.UpdateUserUsedQuotaAndRequestCount(userId, quota)

					model.UpdateChannelUsedQuota(channelId, quota)
				}
			}
		}()
	}()
	switch apiType {
	case APITypeOpenAI:
		if isStream {
			err, responseText := openaiStreamHandler(c, resp, relayMode)
			if err != nil {
				return err
			}
			textResponse.Usage.PromptTokens = promptTokens
			textResponse.Usage.CompletionTokens = countTokenText(responseText, textRequest.Model)
			return nil
		} else {
			err, usage := openaiHandler(c, resp, consumeQuota, promptTokens, textRequest.Model)
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
			err, responseText := claudeStreamHandler(c, resp)
			if err != nil {
				return err
			}
			textResponse.Usage.PromptTokens = promptTokens
			textResponse.Usage.CompletionTokens = countTokenText(responseText, textRequest.Model)
			return nil
		} else {
			err, usage := claudeHandler(c, resp, promptTokens, textRequest.Model)
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
			err, usage := baiduStreamHandler(c, resp)
			if err != nil {
				return err
			}
			if usage != nil {
				textResponse.Usage = *usage
			}
			return nil
		} else {
			var err *OpenAIErrorWithStatusCode
			var usage *Usage
			switch relayMode {
			case RelayModeEmbeddings:
				err, usage = baiduEmbeddingHandler(c, resp)
			default:
				err, usage = baiduHandler(c, resp)
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
			err, responseText := palmStreamHandler(c, resp)
			if err != nil {
				return err
			}
			textResponse.Usage.PromptTokens = promptTokens
			textResponse.Usage.CompletionTokens = countTokenText(responseText, textRequest.Model)
			return nil
		} else {
			err, usage := palmHandler(c, resp, promptTokens, textRequest.Model)
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
			err, usage := zhipuStreamHandler(c, resp)
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
			err, usage := zhipuHandler(c, resp)
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
			err, usage := aliStreamHandler(c, resp)
			if err != nil {
				return err
			}
			if usage != nil {
				textResponse.Usage = *usage
			}
			return nil
		} else {
			err, usage := aliHandler(c, resp)
			if err != nil {
				return err
			}
			if usage != nil {
				textResponse.Usage = *usage
			}
			return nil
		}
	case APITypeXunfei:
		if isStream {
			auth := c.Request.Header.Get("Authorization")
			auth = strings.TrimPrefix(auth, "Bearer ")
			splits := strings.Split(auth, "|")
			if len(splits) != 3 {
				return errorWrapper(errors.New("invalid auth"), "invalid_auth", http.StatusBadRequest)
			}
			err, usage := xunfeiStreamHandler(c, textRequest, splits[0], splits[1], splits[2])
			if err != nil {
				return err
			}
			if usage != nil {
				textResponse.Usage = *usage
			}
			return nil
		} else {
			return errorWrapper(errors.New("xunfei api does not support non-stream mode"), "invalid_api_type", http.StatusBadRequest)
		}
	default:
		return errorWrapper(errors.New("unknown api type"), "unknown_api_type", http.StatusInternalServerError)
	}
}
