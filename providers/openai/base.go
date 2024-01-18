package openai

import (
	"encoding/json"
	"fmt"
	"net/http"
	"one-api/common"
	"one-api/common/requester"
	"one-api/model"
	"one-api/types"
	"strings"

	"one-api/providers/base"
)

type OpenAIProviderFactory struct{}

type OpenAIProvider struct {
	base.BaseProvider
	IsAzure       bool
	BalanceAction bool
}

// 创建 OpenAIProvider
func (f OpenAIProviderFactory) Create(channel *model.Channel) base.ProviderInterface {
	openAIProvider := CreateOpenAIProvider(channel, "https://api.openai.com")
	openAIProvider.BalanceAction = true
	return openAIProvider
}

// 创建 OpenAIProvider
// https://platform.openai.com/docs/api-reference/introduction
func CreateOpenAIProvider(channel *model.Channel, baseURL string) *OpenAIProvider {
	config := getOpenAIConfig(baseURL)

	return &OpenAIProvider{
		BaseProvider: base.BaseProvider{
			Config:    config,
			Channel:   channel,
			Requester: requester.NewHTTPRequester(channel.Proxy, RequestErrorHandle),
		},
		IsAzure:       false,
		BalanceAction: true,
	}
}

func getOpenAIConfig(baseURL string) base.ProviderConfig {
	return base.ProviderConfig{
		BaseURL:             baseURL,
		Completions:         "/v1/completions",
		ChatCompletions:     "/v1/chat/completions",
		Embeddings:          "/v1/embeddings",
		Moderation:          "/v1/moderations",
		AudioSpeech:         "/v1/audio/speech",
		AudioTranscriptions: "/v1/audio/transcriptions",
		AudioTranslations:   "/v1/audio/translations",
		ImagesGenerations:   "/v1/images/generations",
		ImagesEdit:          "/v1/images/edits",
		ImagesVariations:    "/v1/images/variations",
	}
}

// 请求错误处理
func RequestErrorHandle(resp *http.Response) *types.OpenAIError {
	var errorResponse *types.OpenAIErrorResponse
	err := json.NewDecoder(resp.Body).Decode(errorResponse)
	if err != nil {
		return nil
	}

	return ErrorHandle(errorResponse)
}

// 错误处理
func ErrorHandle(openaiError *types.OpenAIErrorResponse) *types.OpenAIError {
	if openaiError.Error.Message == "" {
		return nil
	}
	return &openaiError.Error
}

// 获取完整请求 URL
func (p *OpenAIProvider) GetFullRequestURL(requestURL string, modelName string) string {
	baseURL := strings.TrimSuffix(p.GetBaseURL(), "/")

	if p.IsAzure {
		apiVersion := p.Channel.Other
		// 以-分割，检测modelName 最后一个元素是否为4位数字,必须是数字，如果是则删除modelName最后一个元素
		modelNameSlice := strings.Split(modelName, "-")
		lastModelNameSlice := modelNameSlice[len(modelNameSlice)-1]
		modelNum := common.String2Int(lastModelNameSlice)
		if modelNum > 999 && modelNum < 10000 {
			modelName = strings.TrimSuffix(modelName, "-"+lastModelNameSlice)
		}
		// 检测模型是是否包含 . 如果有则直接去掉
		modelName = strings.Replace(modelName, ".", "", -1)

		if modelName == "dall-e-2" {
			// 因为dall-e-3需要api-version=2023-12-01-preview，但是该版本
			// 已经没有dall-e-2了，所以暂时写死
			requestURL = fmt.Sprintf("/openai/%s:submit?api-version=2023-09-01-preview", requestURL)
		} else {
			requestURL = fmt.Sprintf("/openai/deployments/%s%s?api-version=%s", modelName, requestURL, apiVersion)
		}

	}

	if strings.HasPrefix(baseURL, "https://gateway.ai.cloudflare.com") {
		if p.IsAzure {
			requestURL = strings.TrimPrefix(requestURL, "/openai/deployments")
		} else {
			requestURL = strings.TrimPrefix(requestURL, "/v1")
		}
	}

	return fmt.Sprintf("%s%s", baseURL, requestURL)
}

// 获取请求头
func (p *OpenAIProvider) GetRequestHeaders() (headers map[string]string) {
	headers = make(map[string]string)
	p.CommonRequestHeaders(headers)
	if p.IsAzure {
		headers["api-key"] = p.Channel.Key
	} else {
		headers["Authorization"] = fmt.Sprintf("Bearer %s", p.Channel.Key)
	}

	return headers
}

func (p *OpenAIProvider) GetRequestTextBody(relayMode int, ModelName string, request any) (*http.Request, *types.OpenAIErrorWithStatusCode) {
	url, errWithCode := p.GetSupportedAPIUri(relayMode)
	if errWithCode != nil {
		return nil, errWithCode
	}
	// 获取请求地址
	fullRequestURL := p.GetFullRequestURL(url, ModelName)

	// 获取请求头
	headers := p.GetRequestHeaders()
	// 创建请求
	req, err := p.Requester.NewRequest(http.MethodPost, fullRequestURL, p.Requester.WithBody(request), p.Requester.WithHeader(headers))
	if err != nil {
		return nil, common.ErrorWrapper(err, "new_request_failed", http.StatusInternalServerError)
	}

	return req, nil
}
