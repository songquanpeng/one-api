package providers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"one-api/common"
	"one-api/model"
	"one-api/types"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

var stopFinishReason = "stop"

type ProviderConfig struct {
	BaseURL             string
	Completions         string
	ChatCompletions     string
	Embeddings          string
	AudioSpeech         string
	AudioTranscriptions string
	AudioTranslations   string
	Proxy               string
	Context             *gin.Context
}

type BaseProviderAction interface {
	GetBaseURL() string
	GetFullRequestURL(requestURL string, modelName string) string
	GetRequestHeaders() (headers map[string]string)
}

type CompletionProviderAction interface {
	BaseProviderAction
	CompleteResponse(request *types.CompletionRequest, isModelMapped bool, promptTokens int) (usage *types.Usage, openAIErrorWithStatusCode *types.OpenAIErrorWithStatusCode)
}

type ChatProviderAction interface {
	BaseProviderAction
	ChatCompleteResponse(request *types.ChatCompletionRequest, isModelMapped bool, promptTokens int) (usage *types.Usage, openAIErrorWithStatusCode *types.OpenAIErrorWithStatusCode)
}

type EmbeddingsProviderAction interface {
	BaseProviderAction
	EmbeddingsResponse(request *types.EmbeddingRequest, isModelMapped bool, promptTokens int) (usage *types.Usage, openAIErrorWithStatusCode *types.OpenAIErrorWithStatusCode)
}

type BalanceProviderAction interface {
	Balance(channel *model.Channel) (float64, error)
}

func (p *ProviderConfig) GetBaseURL() string {
	if p.Context.GetString("base_url") != "" {
		return p.Context.GetString("base_url")
	}

	return p.BaseURL
}

func (p *ProviderConfig) GetFullRequestURL(requestURL string, modelName string) string {
	baseURL := strings.TrimSuffix(p.GetBaseURL(), "/")

	return fmt.Sprintf("%s%s", baseURL, requestURL)
}

func setEventStreamHeaders(c *gin.Context) {
	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Header().Set("Transfer-Encoding", "chunked")
	c.Writer.Header().Set("X-Accel-Buffering", "no")
}

func (p *ProviderConfig) handleErrorResp(resp *http.Response) (openAIErrorWithStatusCode *types.OpenAIErrorWithStatusCode) {
	openAIErrorWithStatusCode = &types.OpenAIErrorWithStatusCode{
		StatusCode: resp.StatusCode,
		OpenAIError: types.OpenAIError{
			Message: fmt.Sprintf("bad response status code %d", resp.StatusCode),
			Type:    "upstream_error",
			Code:    "bad_response_status_code",
			Param:   strconv.Itoa(resp.StatusCode),
		},
	}
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return
	}
	err = resp.Body.Close()
	if err != nil {
		return
	}
	var errorResponse types.OpenAIErrorResponse
	err = json.Unmarshal(responseBody, &errorResponse)
	if err != nil {
		return
	}
	if errorResponse.Error.Type != "" {
		openAIErrorWithStatusCode.OpenAIError = errorResponse.Error
	} else {
		openAIErrorWithStatusCode.OpenAIError.Message = string(responseBody)
	}
	return
}

// 供应商响应处理函数
type ProviderResponseHandler interface {
	// 请求处理函数
	requestHandler(resp *http.Response) (OpenAIResponse any, openAIErrorWithStatusCode *types.OpenAIErrorWithStatusCode)
}

// 发送请求
func (p *ProviderConfig) sendRequest(req *http.Request, response ProviderResponseHandler) (openAIErrorWithStatusCode *types.OpenAIErrorWithStatusCode) {

	// 发送请求
	resp, err := common.HttpClient.Do(req)
	if err != nil {
		return types.ErrorWrapper(err, "http_request_failed", http.StatusInternalServerError)
	}

	defer resp.Body.Close()

	// 处理响应
	if common.IsFailureStatusCode(resp) {
		return p.handleErrorResp(resp)
	}

	// 解析响应
	err = common.DecodeResponse(resp.Body, response)
	if err != nil {
		return types.ErrorWrapper(err, "decode_response_failed", http.StatusInternalServerError)
	}

	openAIResponse, openAIErrorWithStatusCode := response.requestHandler(resp)
	if openAIErrorWithStatusCode != nil {
		return
	}

	jsonResponse, err := json.Marshal(openAIResponse)
	if err != nil {
		return types.ErrorWrapper(err, "marshal_response_body_failed", http.StatusInternalServerError)
	}
	p.Context.Writer.Header().Set("Content-Type", "application/json")
	p.Context.Writer.WriteHeader(resp.StatusCode)
	_, err = p.Context.Writer.Write(jsonResponse)
	return nil
}
