package openai

import (
	"encoding/json"
	"fmt"
	"net/http"
	"one-api/common"
	"one-api/common/config"
	"one-api/common/requester"
	"one-api/model"
	"one-api/types"
	"strings"

	"one-api/providers/base"
)

type OpenAIProviderFactory struct{}

type OpenAIProvider struct {
	base.BaseProvider
	IsAzure              bool
	BalanceAction        bool
	SupportStreamOptions bool
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
	openaiConfig := getOpenAIConfig(baseURL)

	OpenAIProvider := &OpenAIProvider{
		BaseProvider: base.BaseProvider{
			Config:    openaiConfig,
			Channel:   channel,
			Requester: requester.NewHTTPRequester(*channel.Proxy, RequestErrorHandle),
		},
		IsAzure:       false,
		BalanceAction: true,
	}

	if channel.Type == config.ChannelTypeOpenAI {
		OpenAIProvider.SupportStreamOptions = true
	}

	return OpenAIProvider
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
		ModelList:           "/v1/models",
	}
}

// 请求错误处理
func RequestErrorHandle(resp *http.Response) *types.OpenAIError {
	errorResponse := &types.OpenAIErrorResponse{}
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
		if modelName != "" {
			// 检测模型是是否包含 . 如果有则直接去掉
			modelName = strings.Replace(modelName, ".", "", -1)

			if modelName == "dall-e-2" {
				// 因为dall-e-3需要api-version=2023-12-01-preview，但是该版本
				// 已经没有dall-e-2了，所以暂时写死
				requestURL = fmt.Sprintf("/openai/%s:submit?api-version=2023-09-01-preview", requestURL)
			} else {
				requestURL = fmt.Sprintf("/openai/deployments/%s%s?api-version=%s", modelName, requestURL, apiVersion)
			}
		} else {
			requestURL = strings.TrimPrefix(requestURL, "/v1")
			requestURL = fmt.Sprintf("/openai%s?api-version=%s", requestURL, apiVersion)
		}

	} else if p.Channel.Type == config.ChannelTypeCustom && p.Channel.Other != "" {
		replaceValue := p.Channel.Other
		if replaceValue == "disable" {
			replaceValue = ""
		} else {
			replaceValue = "/" + replaceValue
		}

		requestURL = strings.Replace(requestURL, "/v1", replaceValue, 1)
	}

	if strings.HasPrefix(baseURL, "https://gateway.ai.cloudflare.com") {
		if p.IsAzure {
			requestURL = strings.TrimPrefix(requestURL, "/openai")
			requestURL = strings.TrimPrefix(requestURL, "/deployments")
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
