package controller

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/pkoukk/tiktoken-go"
	"io"
	"net/http"
	"one-api/common"
	"regexp"
	"strconv"
)

var stopFinishReason = "stop"

var tokenEncoderMap = map[string]*tiktoken.Tiktoken{}

func InitTokenEncoders() {
	common.SysLog("initializing token encoders")
	fallbackTokenEncoder, err := tiktoken.EncodingForModel("gpt-3.5-turbo")
	if err != nil {
		common.FatalLog(fmt.Sprintf("failed to get fallback token encoder: %s", err.Error()))
	}
	for model, _ := range common.ModelRatio {
		tokenEncoder, err := tiktoken.EncodingForModel(model)
		if err != nil {
			common.SysError(fmt.Sprintf("using fallback encoder for model %s", model))
			tokenEncoderMap[model] = fallbackTokenEncoder
			continue
		}
		tokenEncoderMap[model] = tokenEncoder
	}
	common.SysLog("token encoders initialized")
}

func getTokenEncoder(model string) *tiktoken.Tiktoken {
	if tokenEncoder, ok := tokenEncoderMap[model]; ok {
		return tokenEncoder
	}
	tokenEncoder, err := tiktoken.EncodingForModel(model)
	if err != nil {
		common.SysError(fmt.Sprintf("failed to get token encoder for model %s: %s, using encoder for gpt-3.5-turbo", model, err.Error()))
		tokenEncoder, err = tiktoken.EncodingForModel("gpt-3.5-turbo")
		if err != nil {
			common.FatalLog(fmt.Sprintf("failed to get token encoder for model gpt-3.5-turbo: %s", err.Error()))
		}
	}
	tokenEncoderMap[model] = tokenEncoder
	return tokenEncoder
}

func getTokenNum(tokenEncoder *tiktoken.Tiktoken, text string) int {
	if common.ApproximateTokenEnabled {
		return int(float64(len(text)) * 0.38)
	}
	return len(tokenEncoder.Encode(text, nil, nil))
}

func countTokenMessages(messages []Message, model string) int {
	tokenEncoder := getTokenEncoder(model)
	// Reference:
	// https://github.com/openai/openai-cookbook/blob/main/examples/How_to_count_tokens_with_tiktoken.ipynb
	// https://github.com/pkoukk/tiktoken-go/issues/6
	//
	// Every message follows <|start|>{role/name}\n{content}<|end|>\n
	var tokensPerMessage int
	var tokensPerName int
	if model == "gpt-3.5-turbo-0301" {
		tokensPerMessage = 4
		tokensPerName = -1 // If there's a name, the role is omitted
	} else {
		tokensPerMessage = 3
		tokensPerName = 1
	}
	tokenNum := 0
	for _, message := range messages {
		tokenNum += tokensPerMessage
		tokenNum += getTokenNum(tokenEncoder, message.Content)
		tokenNum += getTokenNum(tokenEncoder, message.Role)
		if message.Name != nil {
			tokenNum += tokensPerName
			tokenNum += getTokenNum(tokenEncoder, *message.Name)
		}
	}
	tokenNum += 3 // Every reply is primed with <|start|>assistant<|message|>
	return tokenNum
}

func countTokenInput(input any, model string) int {
	switch input.(type) {
	case string:
		return countTokenText(input.(string), model)
	case []string:
		text := ""
		for _, s := range input.([]string) {
			text += s
		}
		return countTokenText(text, model)
	}
	return 0
}

func countTokenText(text string, model string) int {
	tokenEncoder := getTokenEncoder(model)
	return getTokenNum(tokenEncoder, text)
}

func errorWrapper(err error, code string, statusCode int) *OpenAIErrorWithStatusCode {
	text := err.Error()
	// 定义一个正则表达式匹配URL
	urlPattern := `http[s]?://(?:[a-zA-Z]|[0-9]|[$-_@.&+]|[!*\\(\\),]|(?:%[0-9a-fA-F][0-9a-fA-F]))+`

	urlRegexp, reErr := regexp.Compile(urlPattern)
	if reErr == nil {
		text = urlRegexp.ReplaceAllString(text, "https://api.openai.com")
	}
	//避免暴露内部错误

	openAIError := OpenAIError{
		Message: text,
		Type:    "one_api_error",
		Code:    code,
	}
	return &OpenAIErrorWithStatusCode{
		OpenAIError: openAIError,
		StatusCode:  statusCode,
	}
}

func shouldDisableChannel(err *OpenAIError, statusCode int) bool {
	if !common.AutomaticDisableChannelEnabled {
		return false
	}
	if err == nil {
		return false
	}
	if statusCode == http.StatusUnauthorized || statusCode == http.StatusTooManyRequests {
		return true
	}
	if err.Type == "insufficient_quota" || err.Code == "invalid_api_key" || err.Code == "account_deactivated" {
		return true
	}
	return false
}

func setEventStreamHeaders(c *gin.Context) {
	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Header().Set("Transfer-Encoding", "chunked")
	c.Writer.Header().Set("X-Accel-Buffering", "no")
}

func relayErrorHandler(resp *http.Response) (openAIErrorWithStatusCode *OpenAIErrorWithStatusCode) {
	openAIErrorWithStatusCode = &OpenAIErrorWithStatusCode{
		StatusCode: resp.StatusCode,
		OpenAIError: OpenAIError{
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
	var textResponse TextResponse
	err = json.Unmarshal(responseBody, &textResponse)
	if err != nil {
		return
	}
	openAIErrorWithStatusCode.OpenAIError = textResponse.Error
	return
}
