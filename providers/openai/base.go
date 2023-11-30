package openai

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"one-api/common"
	"one-api/types"
	"strings"

	"one-api/providers/base"

	"github.com/gin-gonic/gin"
)

type OpenAIProvider struct {
	base.BaseProvider
	IsAzure bool
}

// 创建 OpenAIProvider
// https://platform.openai.com/docs/api-reference/introduction
func CreateOpenAIProvider(c *gin.Context, baseURL string) *OpenAIProvider {
	if baseURL == "" {
		baseURL = "https://api.openai.com"
	}

	return &OpenAIProvider{
		BaseProvider: base.BaseProvider{
			BaseURL:             baseURL,
			Completions:         "/v1/completions",
			ChatCompletions:     "/v1/chat/completions",
			Embeddings:          "/v1/embeddings",
			Moderation:          "/v1/moderations",
			AudioSpeech:         "/v1/audio/speech",
			AudioTranscriptions: "/v1/audio/transcriptions",
			AudioTranslations:   "/v1/audio/translations",
			Context:             c,
		},
		IsAzure: false,
	}
}

// 获取完整请求 URL
func (p *OpenAIProvider) GetFullRequestURL(requestURL string, modelName string) string {
	baseURL := strings.TrimSuffix(p.GetBaseURL(), "/")

	if p.IsAzure {
		apiVersion := p.Context.GetString("api_version")
		requestURL = fmt.Sprintf("/openai/deployments/%s%s?api-version=%s", modelName, requestURL, apiVersion)
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
		headers["api-key"] = p.Context.GetString("api_key")
	} else {
		headers["Authorization"] = fmt.Sprintf("Bearer %s", p.Context.GetString("api_key"))
	}

	return headers
}

// 获取请求体
func (p *OpenAIProvider) getRequestBody(request any, isModelMapped bool) (requestBody io.Reader, err error) {
	if isModelMapped {
		jsonStr, err := json.Marshal(request)
		if err != nil {
			return nil, err
		}
		requestBody = bytes.NewBuffer(jsonStr)
	} else {
		requestBody = p.Context.Request.Body
	}
	return
}

// 发送流式请求
func (p *OpenAIProvider) sendStreamRequest(req *http.Request, response OpenAIProviderStreamResponseHandler) (openAIErrorWithStatusCode *types.OpenAIErrorWithStatusCode, responseText string) {

	resp, err := common.HttpClient.Do(req)
	if err != nil {
		return types.ErrorWrapper(err, "http_request_failed", http.StatusInternalServerError), ""
	}

	if common.IsFailureStatusCode(resp) {
		return common.HandleErrorResp(resp), ""
	}

	defer resp.Body.Close()

	scanner := bufio.NewScanner(resp.Body)
	scanner.Split(func(data []byte, atEOF bool) (advance int, token []byte, err error) {
		if atEOF && len(data) == 0 {
			return 0, nil, nil
		}
		if i := strings.Index(string(data), "\n"); i >= 0 {
			return i + 1, data[0:i], nil
		}
		if atEOF {
			return len(data), data, nil
		}
		return 0, nil, nil
	})
	dataChan := make(chan string)
	stopChan := make(chan bool)
	go func() {
		for scanner.Scan() {
			data := scanner.Text()
			if len(data) < 6 { // ignore blank line or wrong format
				continue
			}
			if data[:6] != "data: " && data[:6] != "[DONE]" {
				continue
			}
			dataChan <- data
			data = data[6:]
			if !strings.HasPrefix(data, "[DONE]") {
				err := json.Unmarshal([]byte(data), response)
				if err != nil {
					common.SysError("error unmarshalling stream response: " + err.Error())
					continue // just ignore the error
				}
				responseText += response.responseStreamHandler()
			}
		}
		stopChan <- true
	}()
	common.SetEventStreamHeaders(p.Context)
	p.Context.Stream(func(w io.Writer) bool {
		select {
		case data := <-dataChan:
			if strings.HasPrefix(data, "data: [DONE]") {
				data = data[:12]
			}
			// some implementations may add \r at the end of data
			data = strings.TrimSuffix(data, "\r")
			p.Context.Render(-1, common.CustomEvent{Data: data})
			return true
		case <-stopChan:
			return false
		}
	})

	return nil, responseText
}
