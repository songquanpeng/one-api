package zhipu_v4

import (
	"bufio"
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"one-api/common"
	"one-api/common/logger"
	"one-api/relay/channel/openai"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
)

// https://open.bigmodel.cn/dev/api

var zhipuTokens sync.Map
var expSeconds int64 = 24 * 3600

func GetToken(apikey string) string {
	data, ok := zhipuTokens.Load(apikey)
	if ok {
		tokenData := data.(tokenData)
		if time.Now().Before(tokenData.ExpiryTime) {
			return tokenData.Token
		}
	}

	split := strings.Split(apikey, ".")
	if len(split) != 2 {
		logger.SysError("invalid zhipu key: " + apikey)
		return ""
	}

	id := split[0]
	secret := split[1]

	expMillis := time.Now().Add(time.Duration(expSeconds)*time.Second).UnixNano() / 1e6
	expiryTime := time.Now().Add(time.Duration(expSeconds) * time.Second)

	timestamp := time.Now().UnixNano() / 1e6

	payload := jwt.MapClaims{
		"api_key":   id,
		"exp":       expMillis,
		"timestamp": timestamp,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, payload)

	token.Header["alg"] = "HS256"
	token.Header["sign_type"] = "SIGN"

	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		return ""
	}

	zhipuTokens.Store(apikey, tokenData{
		Token:      tokenString,
		ExpiryTime: expiryTime,
	})

	return tokenString
}

func ConvertRequest(request openai.GeneralOpenAIRequest) *Request {
	messages := make([]Message, 0, len(request.Messages))
	for _, message := range request.Messages {
		messages = append(messages, Message{
			Role:       message.Role,
			Content:    message.StringContent(),
			ToolCalls:  message.ToolCalls,
			ToolCallId: message.ToolCallId,
		})
	}
	str, ok := request.Stop.(string)
	var Stop []string
	if ok {
		Stop = []string{str}
	} else {
		Stop, _ = request.Stop.([]string)
	}
	return &Request{
		Model:       request.Model,
		Stream:      request.Stream,
		Messages:    messages,
		Temperature: request.Temperature,
		TopP:        request.TopP,
		MaxTokens:   request.MaxTokens,
		Stop:        Stop,
		Tools:       request.Tools,
		ToolChoice:  request.ToolChoice,
	}
}

func StreamResponseZhipuV42OpenAI(zhipuResponse *StreamResponse) (*openai.ChatCompletionsStreamResponse) {
	var choice openai.ChatCompletionsStreamResponseChoice
	choice.Delta.Content = zhipuResponse.Choices[0].Delta.Content
	choice.Delta.Role = zhipuResponse.Choices[0].Delta.Role
	choice.Delta.ToolCalls = zhipuResponse.Choices[0].Delta.ToolCalls
	choice.Index = zhipuResponse.Choices[0].Index
	choice.FinishReason = zhipuResponse.Choices[0].FinishReason
	response := openai.ChatCompletionsStreamResponse{
		Id:      zhipuResponse.Id,
		Object:  "chat.completion.chunk",
		Created: zhipuResponse.Created,
		Model:   "glm-4",
		Choices: []openai.ChatCompletionsStreamResponseChoice{choice},
	}
	return &response
}

func LastStreamResponseZhipuV42OpenAI(zhipuResponse *StreamResponse) (*openai.ChatCompletionsStreamResponse, *openai.Usage) {
	response := StreamResponseZhipuV42OpenAI(zhipuResponse)
	return response, &zhipuResponse.Usage
}

func StreamHandler(c *gin.Context, resp *http.Response) (*openai.ErrorWithStatusCode, *openai.Usage) {
	var usage *openai.Usage
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
		}
		stopChan <- true
	}()
	common.SetEventStreamHeaders(c)
	c.Stream(func(w io.Writer) bool {
		select {
		case data := <-dataChan:
			if strings.HasPrefix(data, "data: [DONE]") {
				data = data[:12]
			}
			// some implementations may add \r at the end of data
			data = strings.TrimSuffix(data, "\r")

			var streamResponse StreamResponse
			err := json.Unmarshal([]byte(data), &streamResponse)
			if err != nil {
				logger.SysError("error unmarshalling stream response: " + err.Error())
			}
			var response *openai.ChatCompletionsStreamResponse
			if strings.Contains(data, "prompt_tokens") {
				response, usage = LastStreamResponseZhipuV42OpenAI(&streamResponse)
			} else {
				response = StreamResponseZhipuV42OpenAI(&streamResponse)
			}
			jsonResponse, err := json.Marshal(response)
			if err != nil {
				logger.SysError("error marshalling stream response: " + err.Error())
				return true
			}
			c.Render(-1, common.CustomEvent{Data: "data: " + string(jsonResponse)})
			return true
		case <-stopChan:
			return false
		}
	})
	err := resp.Body.Close()
	if err != nil {
		return openai.ErrorWrapper(err, "close_response_body_failed", http.StatusInternalServerError), nil
	}
	return nil, usage
}

func Handler(c *gin.Context, resp *http.Response) (*openai.ErrorWithStatusCode, *openai.Usage) {
	var textResponse Response
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return openai.ErrorWrapper(err, "read_response_body_failed", http.StatusInternalServerError), nil
	}
	err = resp.Body.Close()
	if err != nil {
		return openai.ErrorWrapper(err, "close_response_body_failed", http.StatusInternalServerError), nil
	}
	err = json.Unmarshal(responseBody, &textResponse)
	if err != nil {
		return openai.ErrorWrapper(err, "unmarshal_response_body_failed", http.StatusInternalServerError), nil
	}
	if textResponse.Error.Type != "" {
		return &openai.ErrorWithStatusCode{
			Error:      textResponse.Error,
			StatusCode: resp.StatusCode,
		}, nil
	}
	// Reset response body
	resp.Body = io.NopCloser(bytes.NewBuffer(responseBody))

	// We shouldn't set the header before we parse the response body, because the parse part may fail.
	// And then we will have to send an error response, but in this case, the header has already been set.
	// So the HTTPClient will be confused by the response.
	// For example, Postman will report error, and we cannot check the response at all.
	for k, v := range resp.Header {
		c.Writer.Header().Set(k, v[0])
	}
	c.Writer.WriteHeader(resp.StatusCode)
	_, err = io.Copy(c.Writer, resp.Body)
	if err != nil {
		return openai.ErrorWrapper(err, "copy_response_body_failed", http.StatusInternalServerError), nil
	}
	err = resp.Body.Close()
	if err != nil {
		return openai.ErrorWrapper(err, "close_response_body_failed", http.StatusInternalServerError), nil
	}

	return nil, &textResponse.Usage
}
