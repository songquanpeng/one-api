package controller

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"io"
	"net/http"
	"one-api/common"
	"one-api/model"
	"strings"
)

type Message struct {
	Role    string  `json:"role"`
	Content string  `json:"content"`
	Name    *string `json:"name,omitempty"`
}

type ChatRequest struct {
	Model     string    `json:"model"`
	Messages  []Message `json:"messages"`
	MaxTokens int       `json:"max_tokens"`
}

type TextRequest struct {
	Model     string    `json:"model"`
	Messages  []Message `json:"messages"`
	Prompt    string    `json:"prompt"`
	MaxTokens int       `json:"max_tokens"`
	//Stream   bool      `json:"stream"`
}

type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

type OpenAIError struct {
	Message string `json:"message"`
	Type    string `json:"type"`
	Param   string `json:"param"`
	Code    string `json:"code"`
}

type OpenAIErrorWithStatusCode struct {
	OpenAIError
	StatusCode int `json:"status_code"`
}

type TextResponse struct {
	Usage `json:"usage"`
	Error OpenAIError `json:"error"`
}

type StreamResponse struct {
	Choices []struct {
		Delta struct {
			Content string `json:"content"`
		} `json:"delta"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
}

func Relay(c *gin.Context) {
	err := relayHelper(c)
	if err != nil {
		if err.StatusCode == http.StatusTooManyRequests {
			err.OpenAIError.Message = "负载已满，请稍后再试，或升级账户以提升服务质量。"
		}
		c.JSON(err.StatusCode, gin.H{
			"error": err.OpenAIError,
		})
		channelId := c.GetInt("channel_id")
		common.SysError(fmt.Sprintf("Relay error (channel #%d): %s", channelId, err.Message))
		if err.Type != "invalid_request_error" && err.StatusCode != http.StatusTooManyRequests &&
			common.AutomaticDisableChannelEnabled {
			channelId := c.GetInt("channel_id")
			channelName := c.GetString("channel_name")
			disableChannel(channelId, channelName, err.Message)
		}
	}
}

func errorWrapper(err error, code string, statusCode int) *OpenAIErrorWithStatusCode {
	openAIError := OpenAIError{
		Message: err.Error(),
		Type:    "one_api_error",
		Code:    code,
	}
	return &OpenAIErrorWithStatusCode{
		OpenAIError: openAIError,
		StatusCode:  statusCode,
	}
}

func relayHelper(c *gin.Context) *OpenAIErrorWithStatusCode {
	channelType := c.GetInt("channel")
	tokenId := c.GetInt("token_id")
	consumeQuota := c.GetBool("consume_quota")
	var textRequest TextRequest
	if consumeQuota || channelType == common.ChannelTypeAzure {
		requestBody, err := io.ReadAll(c.Request.Body)
		if err != nil {
			return errorWrapper(err, "read_request_body_failed", http.StatusBadRequest)
		}
		err = c.Request.Body.Close()
		if err != nil {
			return errorWrapper(err, "close_request_body_failed", http.StatusBadRequest)
		}
		err = json.Unmarshal(requestBody, &textRequest)
		if err != nil {
			return errorWrapper(err, "unmarshal_request_body_failed", http.StatusBadRequest)
		}
		// Reset request body
		c.Request.Body = io.NopCloser(bytes.NewBuffer(requestBody))
	}
	baseURL := common.ChannelBaseURLs[channelType]
	requestURL := c.Request.URL.String()
	if channelType == common.ChannelTypeCustom {
		baseURL = c.GetString("base_url")
	}
	fullRequestURL := fmt.Sprintf("%s%s", baseURL, requestURL)
	if channelType == common.ChannelTypeAzure {
		// https://learn.microsoft.com/en-us/azure/cognitive-services/openai/chatgpt-quickstart?pivots=rest-api&tabs=command-line#rest-api
		query := c.Request.URL.Query()
		apiVersion := query.Get("api-version")
		if apiVersion == "" {
			apiVersion = c.GetString("api_version")
		}
		requestURL := strings.Split(requestURL, "?")[0]
		requestURL = fmt.Sprintf("%s?api-version=%s", requestURL, apiVersion)
		baseURL = c.GetString("base_url")
		task := strings.TrimPrefix(requestURL, "/v1/")
		model_ := textRequest.Model
		model_ = strings.Replace(model_, ".", "", -1)
		// https://github.com/songquanpeng/one-api/issues/67
		model_ = strings.TrimSuffix(model_, "-0301")
		model_ = strings.TrimSuffix(model_, "-0314")
		fullRequestURL = fmt.Sprintf("%s/openai/deployments/%s/%s", baseURL, model_, task)
	}

	promptTokens := countTokenMessages(textRequest.Messages, textRequest.Model)
	preConsumedTokens := common.PreConsumedQuota
	if textRequest.MaxTokens != 0 {
		preConsumedTokens = promptTokens + textRequest.MaxTokens
	}
	ratio := common.GetModelRatio(textRequest.Model)
	preConsumedQuota := int(float64(preConsumedTokens) * ratio)
	if consumeQuota {
		err := model.PreConsumeTokenQuota(tokenId, preConsumedQuota)
		if err != nil {
			return errorWrapper(err, "pre_consume_token_quota_failed", http.StatusOK)
		}
	}
	req, err := http.NewRequest(c.Request.Method, fullRequestURL, c.Request.Body)
	if err != nil {
		return errorWrapper(err, "new_request_failed", http.StatusOK)
	}
	if channelType == common.ChannelTypeAzure {
		key := c.Request.Header.Get("Authorization")
		key = strings.TrimPrefix(key, "Bearer ")
		req.Header.Set("api-key", key)
	} else {
		req.Header.Set("Authorization", c.Request.Header.Get("Authorization"))
	}
	req.Header.Set("Content-Type", c.Request.Header.Get("Content-Type"))
	req.Header.Set("Accept", c.Request.Header.Get("Accept"))
	req.Header.Set("Connection", c.Request.Header.Get("Connection"))
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return errorWrapper(err, "do_request_failed", http.StatusOK)
	}
	err = req.Body.Close()
	if err != nil {
		return errorWrapper(err, "close_request_body_failed", http.StatusOK)
	}
	err = c.Request.Body.Close()
	if err != nil {
		return errorWrapper(err, "close_request_body_failed", http.StatusOK)
	}
	var textResponse TextResponse
	isStream := strings.HasPrefix(resp.Header.Get("Content-Type"), "text/event-stream")
	var streamResponseText string

	defer func() {
		if consumeQuota {
			quota := 0
			usingGPT4 := strings.HasPrefix(textRequest.Model, "gpt-4")
			completionRatio := 1
			if usingGPT4 {
				completionRatio = 2
			}
			if isStream {
				responseTokens := countTokenText(streamResponseText, textRequest.Model)
				quota = promptTokens + responseTokens*completionRatio
			} else {
				quota = textResponse.Usage.PromptTokens + textResponse.Usage.CompletionTokens*completionRatio
			}
			quota = int(float64(quota) * ratio)
			quotaDelta := quota - preConsumedQuota
			err := model.PostConsumeTokenQuota(tokenId, quotaDelta)
			if err != nil {
				common.SysError("Error consuming token remain quota: " + err.Error())
			}
		}
	}()

	if isStream {
		scanner := bufio.NewScanner(resp.Body)
		scanner.Split(func(data []byte, atEOF bool) (advance int, token []byte, err error) {
			if atEOF && len(data) == 0 {
				return 0, nil, nil
			}

			if i := strings.Index(string(data), "\n\n"); i >= 0 {
				return i + 2, data[0:i], nil
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
				if len(data) < 6 { // must be something wrong!
					common.SysError("Invalid stream response: " + data)
					continue
				}
				dataChan <- data
				data = data[6:]
				if !strings.HasPrefix(data, "[DONE]") {
					var streamResponse StreamResponse
					err = json.Unmarshal([]byte(data), &streamResponse)
					if err != nil {
						common.SysError("Error unmarshalling stream response: " + err.Error())
						return
					}
					for _, choice := range streamResponse.Choices {
						streamResponseText += choice.Delta.Content
					}
				}
			}
			stopChan <- true
		}()
		c.Writer.Header().Set("Content-Type", "text/event-stream")
		c.Writer.Header().Set("Cache-Control", "no-cache")
		c.Writer.Header().Set("Connection", "keep-alive")
		c.Writer.Header().Set("Transfer-Encoding", "chunked")
		c.Writer.Header().Set("X-Accel-Buffering", "no")
		c.Stream(func(w io.Writer) bool {
			select {
			case data := <-dataChan:
				if strings.HasPrefix(data, "data: [DONE]") {
					data = data[:12]
				}
				c.Render(-1, common.CustomEvent{Data: data})
				return true
			case <-stopChan:
				return false
			}
		})
		err = resp.Body.Close()
		if err != nil {
			return errorWrapper(err, "close_response_body_failed", http.StatusOK)
		}
		return nil
	} else {
		if consumeQuota {
			responseBody, err := io.ReadAll(resp.Body)
			if err != nil {
				return errorWrapper(err, "read_response_body_failed", http.StatusOK)
			}
			err = resp.Body.Close()
			if err != nil {
				return errorWrapper(err, "close_response_body_failed", http.StatusOK)
			}
			err = json.Unmarshal(responseBody, &textResponse)
			if err != nil {
				return errorWrapper(err, "unmarshal_response_body_failed", http.StatusOK)
			}
			if textResponse.Error.Type != "" {
				return &OpenAIErrorWithStatusCode{
					OpenAIError: textResponse.Error,
					StatusCode:  resp.StatusCode,
				}
			}
			// Reset response body
			resp.Body = io.NopCloser(bytes.NewBuffer(responseBody))
		}
		// We shouldn't set the header before we parse the response body, because the parse part may fail.
		// And then we will have to send an error response, but in this case, the header has already been set.
		// So the client will be confused by the response.
		// For example, Postman will report error, and we cannot check the response at all.
		for k, v := range resp.Header {
			c.Writer.Header().Set(k, v[0])
		}
		c.Writer.WriteHeader(resp.StatusCode)
		_, err = io.Copy(c.Writer, resp.Body)
		if err != nil {
			return errorWrapper(err, "copy_response_body_failed", http.StatusOK)
		}
		err = resp.Body.Close()
		if err != nil {
			return errorWrapper(err, "close_response_body_failed", http.StatusOK)
		}
		return nil
	}
}

func RelayNotImplemented(c *gin.Context) {
	err := OpenAIError{
		Message: "API not implemented",
		Type:    "one_api_error",
		Param:   "",
		Code:    "api_not_implemented",
	}
	c.JSON(http.StatusOK, gin.H{
		"error": err,
	})
}
