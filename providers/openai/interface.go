package openai

import (
	"net/http"
	"one-api/types"
)

type OpenAIProviderResponseHandler interface {
	// 请求处理函数
	responseHandler(resp *http.Response) (errWithCode *types.OpenAIErrorWithStatusCode)
}

type OpenAIProviderStreamResponseHandler interface {
	// 请求流处理函数
	responseStreamHandler() (responseText string)
}
