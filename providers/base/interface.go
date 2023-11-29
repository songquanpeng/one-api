package base

import (
	"net/http"
	"one-api/model"
	"one-api/types"
)

// 基础接口
type ProviderInterface interface {
	GetBaseURL() string
	GetFullRequestURL(requestURL string, modelName string) string
	GetRequestHeaders() (headers map[string]string)
}

// 完成接口
type CompletionInterface interface {
	ProviderInterface
	CompleteAction(request *types.CompletionRequest, isModelMapped bool, promptTokens int) (usage *types.Usage, errWithCode *types.OpenAIErrorWithStatusCode)
}

// 聊天接口
type ChatInterface interface {
	ProviderInterface
	ChatAction(request *types.ChatCompletionRequest, isModelMapped bool, promptTokens int) (usage *types.Usage, errWithCode *types.OpenAIErrorWithStatusCode)
}

// 嵌入接口
type EmbeddingsInterface interface {
	ProviderInterface
	EmbeddingsAction(request *types.EmbeddingRequest, isModelMapped bool, promptTokens int) (usage *types.Usage, errWithCode *types.OpenAIErrorWithStatusCode)
}

// 余额接口
type BalanceInterface interface {
	BalanceAction(channel *model.Channel) (float64, error)
}

type ProviderResponseHandler interface {
	// 响应处理函数
	ResponseHandler(resp *http.Response) (OpenAIResponse any, errWithCode *types.OpenAIErrorWithStatusCode)
}
