package openai

type OpenAIProviderStreamResponseHandler interface {
	// 请求流处理函数
	responseStreamHandler() (responseText string)
}
