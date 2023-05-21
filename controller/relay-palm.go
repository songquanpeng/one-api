package controller

import (
	"fmt"
	"github.com/gin-gonic/gin"
)

type PaLMChatMessage struct {
	Author  string `json:"author"`
	Content string `json:"content"`
}

type PaLMFilter struct {
	Reason  string `json:"reason"`
	Message string `json:"message"`
}

// https://developers.generativeai.google/api/rest/generativelanguage/models/generateMessage#request-body
type PaLMChatRequest struct {
	Prompt         []Message `json:"prompt"`
	Temperature    float64   `json:"temperature"`
	CandidateCount int       `json:"candidateCount"`
	TopP           float64   `json:"topP"`
	TopK           int       `json:"topK"`
}

// https://developers.generativeai.google/api/rest/generativelanguage/models/generateMessage#response-body
type PaLMChatResponse struct {
	Candidates []Message    `json:"candidates"`
	Messages   []Message    `json:"messages"`
	Filters    []PaLMFilter `json:"filters"`
}

func relayPaLM(openAIRequest GeneralOpenAIRequest, c *gin.Context) *OpenAIErrorWithStatusCode {
	// https://developers.generativeai.google/api/rest/generativelanguage/models/generateMessage
	messages := make([]PaLMChatMessage, 0, len(openAIRequest.Messages))
	for _, message := range openAIRequest.Messages {
		var author string
		if message.Role == "user" {
			author = "0"
		} else {
			author = "1"
		}
		messages = append(messages, PaLMChatMessage{
			Author:  author,
			Content: message.Content,
		})
	}
	request := PaLMChatRequest{
		Prompt:         nil,
		Temperature:    openAIRequest.Temperature,
		CandidateCount: openAIRequest.N,
		TopP:           openAIRequest.TopP,
		TopK:           openAIRequest.MaxTokens,
	}
	// TODO: forward request to PaLM & convert response
	fmt.Print(request)
	return nil
}
