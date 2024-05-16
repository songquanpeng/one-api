package mistral

import (
	"encoding/json"
	"fmt"
	"net/http"
	"one-api/common/requester"
	"one-api/model"
	"one-api/types"

	"one-api/providers/base"
)

type MistralProviderFactory struct{}

type MistralProvider struct {
	base.BaseProvider
}

// 创建 MistralProvider
func (f MistralProviderFactory) Create(channel *model.Channel) base.ProviderInterface {
	MistralProvider := CreateMistralProvider(channel, "https://api.mistral.ai")
	return MistralProvider
}

// 创建 MistralProvider
func CreateMistralProvider(channel *model.Channel, baseURL string) *MistralProvider {
	config := getMistralConfig(baseURL)

	return &MistralProvider{
		BaseProvider: base.BaseProvider{
			Config:    config,
			Channel:   channel,
			Requester: requester.NewHTTPRequester(*channel.Proxy, RequestErrorHandle),
		},
	}
}

func getMistralConfig(baseURL string) base.ProviderConfig {
	return base.ProviderConfig{
		BaseURL:         baseURL,
		ChatCompletions: "/v1/chat/completions",
		Embeddings:      "/v1/embeddings",
		ModelList:       "/v1/models",
	}
}

// 请求错误处理
func RequestErrorHandle(resp *http.Response) *types.OpenAIError {
	errorResponse := &MistralError{}
	err := json.NewDecoder(resp.Body).Decode(errorResponse)
	if err != nil {
		return nil
	}

	return errorHandle(errorResponse)
}

// 错误处理
func errorHandle(MistralError *MistralError) *types.OpenAIError {
	if MistralError.Object != "error" {
		return nil
	}
	return &types.OpenAIError{
		Message: MistralError.Message.Detail[0].errorMsg(),
		Type:    MistralError.Type,
	}
}

// 获取请求头
func (p *MistralProvider) GetRequestHeaders() (headers map[string]string) {
	headers = make(map[string]string)
	p.CommonRequestHeaders(headers)

	headers["Authorization"] = fmt.Sprintf("Bearer %s", p.Channel.Key)

	return headers
}
