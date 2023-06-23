package controller

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"one-api/common"
	"strings"
)

type Message struct {
	Role    string  `json:"role"`
	Content string  `json:"content"`
	Name    *string `json:"name,omitempty"`
}

const (
	RelayModeUnknown = iota
	RelayModeChatCompletions
	RelayModeCompletions
	RelayModeEmbeddings
	RelayModeModeration
	RelayModeImagesGenerations
)

// https://platform.openai.com/docs/api-reference/chat

type GeneralOpenAIRequest struct {
	Model       string    `json:"model"`
	Messages    []Message `json:"messages"`
	Prompt      any       `json:"prompt"`
	Stream      bool      `json:"stream"`
	MaxTokens   int       `json:"max_tokens"`
	Temperature float64   `json:"temperature"`
	TopP        float64   `json:"top_p"`
	N           int       `json:"n"`
	Input       any       `json:"input"`
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
	Code    any    `json:"code"`
}

type OpenAIErrorWithStatusCode struct {
	OpenAIError
	StatusCode int `json:"status_code"`
}

type TextResponse struct {
	Usage `json:"usage"`
	Error OpenAIError `json:"error"`
}

type ChatCompletionsStreamResponse struct {
	Choices []struct {
		Delta struct {
			Content string `json:"content"`
		} `json:"delta"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
}

type CompletionsStreamResponse struct {
	Choices []struct {
		Text         string `json:"text"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
}

func Relay(c *gin.Context) {
	relayMode := RelayModeUnknown
	if strings.HasPrefix(c.Request.URL.Path, "/v1/chat/completions") {
		relayMode = RelayModeChatCompletions
	} else if strings.HasPrefix(c.Request.URL.Path, "/v1/completions") {
		relayMode = RelayModeCompletions
	} else if strings.HasPrefix(c.Request.URL.Path, "/v1/embeddings") {
		relayMode = RelayModeEmbeddings
	} else if strings.HasPrefix(c.Request.URL.Path, "/v1/moderations") {
		relayMode = RelayModeModeration
	} else if strings.HasPrefix(c.Request.URL.Path, "/v1/images/generations") {
		relayMode = RelayModeImagesGenerations
	}
	var err *OpenAIErrorWithStatusCode
	switch relayMode {
	case RelayModeImagesGenerations:
		err = relayImageHelper(c, relayMode)
	default:
		err = relayTextHelper(c, relayMode)
	}
	if err != nil {
		if err.StatusCode == http.StatusTooManyRequests {
			err.OpenAIError.Message = "当前分组负载已饱和，请稍后再试，或升级账户以提升服务质量。"
		}
		c.JSON(err.StatusCode, gin.H{
			"error": err.OpenAIError,
		})
		channelId := c.GetInt("channel_id")
		common.SysError(fmt.Sprintf("relay error (channel #%d): %s", channelId, err.Message))
		// https://platform.openai.com/docs/guides/error-codes/api-errors
		if common.AutomaticDisableChannelEnabled && (err.Type == "insufficient_quota" || err.Code == "invalid_api_key") {
			channelId := c.GetInt("channel_id")
			channelName := c.GetString("channel_name")
			disableChannel(channelId, channelName, err.Message)
		}
	}
}

func RelayNotImplemented(c *gin.Context) {
	err := OpenAIError{
		Message: "API not implemented",
		Type:    "one_api_error",
		Param:   "",
		Code:    "api_not_implemented",
	}
	c.JSON(http.StatusNotImplemented, gin.H{
		"error": err,
	})
}

func RelayNotFound(c *gin.Context) {
	err := OpenAIError{
		Message: fmt.Sprintf("API not found: %s:%s", c.Request.Method, c.Request.URL.Path),
		Type:    "one_api_error",
		Param:   "",
		Code:    "api_not_found",
	}
	c.JSON(http.StatusNotFound, gin.H{
		"error": err,
	})
}
