package base

import (
	"net/http"
	"one-api/common/requester"
	"one-api/model"
	"one-api/types"

	"github.com/gin-gonic/gin"
)

type Requestable interface {
	types.CompletionRequest | types.ChatCompletionRequest | types.EmbeddingRequest | types.ModerationRequest | types.SpeechAudioRequest | types.AudioRequest | types.ImageRequest | types.ImageEditRequest
}

// 基础接口
type ProviderInterface interface {
	// 获取基础URL
	// GetBaseURL() string
	// 获取完整请求URL
	// GetFullRequestURL(requestURL string, modelName string) string
	// 获取请求头
	// GetRequestHeaders() (headers map[string]string)
	// 获取用量
	GetUsage() *types.Usage
	// 设置用量
	SetUsage(usage *types.Usage)
	// 设置Context
	SetContext(c *gin.Context)
	// 设置原始模型
	SetOriginalModel(ModelName string)
	// 获取原始模型
	GetOriginalModel() string

	// SupportAPI(relayMode int) bool
	GetChannel() *model.Channel
	ModelMappingHandler(modelName string) (string, error)
}

// 完成接口
type CompletionInterface interface {
	ProviderInterface
	CreateCompletion(request *types.CompletionRequest) (*types.CompletionResponse, *types.OpenAIErrorWithStatusCode)
	CreateCompletionStream(request *types.CompletionRequest) (requester.StreamReaderInterface[types.CompletionResponse], *types.OpenAIErrorWithStatusCode)
}

// 聊天接口
type ChatInterface interface {
	ProviderInterface
	CreateChatCompletion(request *types.ChatCompletionRequest) (*types.ChatCompletionResponse, *types.OpenAIErrorWithStatusCode)
	CreateChatCompletionStream(request *types.ChatCompletionRequest) (requester.StreamReaderInterface[types.ChatCompletionStreamResponse], *types.OpenAIErrorWithStatusCode)
}

// 嵌入接口
type EmbeddingsInterface interface {
	ProviderInterface
	CreateEmbeddings(request *types.EmbeddingRequest) (*types.EmbeddingResponse, *types.OpenAIErrorWithStatusCode)
}

// 审查接口
type ModerationInterface interface {
	ProviderInterface
	CreateModeration(request *types.ModerationRequest) (*types.ModerationResponse, *types.OpenAIErrorWithStatusCode)
}

// 文字转语音接口
type SpeechInterface interface {
	ProviderInterface
	CreateSpeech(request *types.SpeechAudioRequest) (*http.Response, *types.OpenAIErrorWithStatusCode)
}

// 语音转文字接口
type TranscriptionsInterface interface {
	ProviderInterface
	CreateTranscriptions(request *types.AudioRequest) (*types.AudioResponseWrapper, *types.OpenAIErrorWithStatusCode)
}

// 语音翻译接口
type TranslationInterface interface {
	ProviderInterface
	CreateTranslation(request *types.AudioRequest) (*types.AudioResponseWrapper, *types.OpenAIErrorWithStatusCode)
}

// 图片生成接口
type ImageGenerationsInterface interface {
	ProviderInterface
	CreateImageGenerations(request *types.ImageRequest) (*types.ImageResponse, *types.OpenAIErrorWithStatusCode)
}

// 图片编辑接口
type ImageEditsInterface interface {
	ProviderInterface
	CreateImageEdits(request *types.ImageEditRequest) (*types.ImageResponse, *types.OpenAIErrorWithStatusCode)
}

type ImageVariationsInterface interface {
	ProviderInterface
	CreateImageVariations(request *types.ImageEditRequest) (*types.ImageResponse, *types.OpenAIErrorWithStatusCode)
}

// 余额接口
type BalanceInterface interface {
	Balance() (float64, error)
}

// type ProviderResponseHandler interface {
// 	// 响应处理函数
// 	ResponseHandler(resp *http.Response) (OpenAIResponse any, errWithCode *types.OpenAIErrorWithStatusCode)
// }
