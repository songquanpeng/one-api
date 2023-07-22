package controller

import (
	"fmt"
	"one-api/common"
	"strings"
)

type ClaudeMetadata struct {
	UserId string `json:"user_id"`
}

type ClaudeRequest struct {
	Model             string   `json:"model"`
	Prompt            string   `json:"prompt"`
	MaxTokensToSample int      `json:"max_tokens_to_sample"`
	StopSequences     []string `json:"stop_sequences,omitempty"`
	Temperature       float64  `json:"temperature,omitempty"`
	TopP              float64  `json:"top_p,omitempty"`
	TopK              int      `json:"top_k,omitempty"`
	//ClaudeMetadata    `json:"metadata,omitempty"`
	Stream bool `json:"stream,omitempty"`
}

type ClaudeError struct {
	Type    string `json:"type"`
	Message string `json:"message"`
}

type ClaudeResponse struct {
	Completion string      `json:"completion"`
	StopReason string      `json:"stop_reason"`
	Model      string      `json:"model"`
	Error      ClaudeError `json:"error"`
}

func stopReasonClaude2OpenAI(reason string) string {
	switch reason {
	case "stop_sequence":
		return "stop"
	case "max_tokens":
		return "length"
	default:
		return reason
	}
}

func requestOpenAI2Claude(textRequest GeneralOpenAIRequest) *ClaudeRequest {
	claudeRequest := ClaudeRequest{
		Model:             textRequest.Model,
		Prompt:            "",
		MaxTokensToSample: textRequest.MaxTokens,
		StopSequences:     nil,
		Temperature:       textRequest.Temperature,
		TopP:              textRequest.TopP,
		Stream:            textRequest.Stream,
	}
	if claudeRequest.MaxTokensToSample == 0 {
		claudeRequest.MaxTokensToSample = 1000000
	}
	prompt := ""
	for _, message := range textRequest.Messages {
		if message.Role == "user" {
			prompt += fmt.Sprintf("\n\nHuman: %s", message.Content)
		} else if message.Role == "assistant" {
			prompt += fmt.Sprintf("\n\nAssistant: %s", message.Content)
		} else {
			// ignore other roles
		}
		prompt += "\n\nAssistant:"
	}
	claudeRequest.Prompt = prompt
	return &claudeRequest
}

func streamResponseClaude2OpenAI(claudeResponse *ClaudeResponse) *ChatCompletionsStreamResponse {
	var choice ChatCompletionsStreamResponseChoice
	choice.Delta.Content = claudeResponse.Completion
	choice.FinishReason = stopReasonClaude2OpenAI(claudeResponse.StopReason)
	var response ChatCompletionsStreamResponse
	response.Object = "chat.completion.chunk"
	response.Model = claudeResponse.Model
	response.Choices = []ChatCompletionsStreamResponseChoice{choice}
	return &response
}

func responseClaude2OpenAI(claudeResponse *ClaudeResponse) *OpenAITextResponse {
	choice := OpenAITextResponseChoice{
		Index: 0,
		Message: Message{
			Role:    "assistant",
			Content: strings.TrimPrefix(claudeResponse.Completion, " "),
			Name:    nil,
		},
		FinishReason: stopReasonClaude2OpenAI(claudeResponse.StopReason),
	}
	fullTextResponse := OpenAITextResponse{
		Id:      fmt.Sprintf("chatcmpl-%s", common.GetUUID()),
		Object:  "chat.completion",
		Created: common.GetTimestamp(),
		Choices: []OpenAITextResponseChoice{choice},
	}
	return &fullTextResponse
}
