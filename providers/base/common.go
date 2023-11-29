package base

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"one-api/common"
	"one-api/types"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

var StopFinishReason = "stop"

type BaseProvider struct {
	BaseURL             string
	Completions         string
	ChatCompletions     string
	Embeddings          string
	AudioSpeech         string
	Moderation          string
	AudioTranscriptions string
	AudioTranslations   string
	Proxy               string
	Context             *gin.Context
}

// 获取基础URL
func (p *BaseProvider) GetBaseURL() string {
	if p.Context.GetString("base_url") != "" {
		return p.Context.GetString("base_url")
	}

	return p.BaseURL
}

// 获取完整请求URL
func (p *BaseProvider) GetFullRequestURL(requestURL string, modelName string) string {
	baseURL := strings.TrimSuffix(p.GetBaseURL(), "/")

	return fmt.Sprintf("%s%s", baseURL, requestURL)
}

// 获取请求头
func (p *BaseProvider) CommonRequestHeaders(headers map[string]string) {
	headers["Content-Type"] = p.Context.Request.Header.Get("Content-Type")
	headers["Accept"] = p.Context.Request.Header.Get("Accept")
	if headers["Content-Type"] == "" {
		headers["Content-Type"] = "application/json"
	}
}

// 发送请求
func (p *BaseProvider) SendRequest(req *http.Request, response ProviderResponseHandler) (openAIErrorWithStatusCode *types.OpenAIErrorWithStatusCode) {

	// 发送请求
	resp, err := common.HttpClient.Do(req)
	if err != nil {
		return types.ErrorWrapper(err, "http_request_failed", http.StatusInternalServerError)
	}

	defer resp.Body.Close()

	// 处理响应
	if common.IsFailureStatusCode(resp) {
		return p.HandleErrorResp(resp)
	}

	// 解析响应
	err = common.DecodeResponse(resp.Body, response)
	if err != nil {
		return types.ErrorWrapper(err, "decode_response_failed", http.StatusInternalServerError)
	}

	openAIResponse, openAIErrorWithStatusCode := response.ResponseHandler(resp)
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

	if err != nil {
		return types.ErrorWrapper(err, "write_response_body_failed", http.StatusInternalServerError)
	}

	return nil
}

// 处理错误响应
func (p *BaseProvider) HandleErrorResp(resp *http.Response) (openAIErrorWithStatusCode *types.OpenAIErrorWithStatusCode) {
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

func (p *BaseProvider) SupportAPI(relayMode int) bool {
	switch relayMode {
	case common.RelayModeChatCompletions:
		return p.ChatCompletions != ""
	case common.RelayModeCompletions:
		return p.Completions != ""
	case common.RelayModeEmbeddings:
		return p.Embeddings != ""
	case common.RelayModeAudioSpeech:
		return p.AudioSpeech != ""
	case common.RelayModeAudioTranscription:
		return p.AudioTranscriptions != ""
	case common.RelayModeAudioTranslation:
		return p.AudioTranslations != ""
	case common.RelayModeModerations:
		return p.Moderation != ""
	default:
		return false
	}
}
