package controller

import (
	"bufio"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"io"
	"net/http"
	"one-api/common"
	"reflect"
	"strings"
)

// https://api.minimax.chat/document/guides/chat?id=6433f37294878d408fc82953

type MinimaxError struct {
	StatusCode int    `json:"status_code"`
	StatusMsg  string `json:"status_msg"`
}

type MinimaxChatMessage struct {
	SenderType string `json:"sender_type,omitempty"` //USER or BOT
	Text       string `json:"text,omitempty"`
}

type MinimaxChatRequest struct {
	Model       string               `json:"model,omitempty"`
	Stream      bool                 `json:"stream,omitempty"`
	Prompt      string               `json:"prompt,omitempty"`
	Messages    []MinimaxChatMessage `json:"messages,omitempty"`
	Temperature float64              `json:"temperature,omitempty"`
	TopP        float64              `json:"top_p,omitempty"`
}

type MinimaxChoice struct {
	Text         string `json:"text"`
	Index        int    `json:"index"`
	FinishReason string `json:"finish_reason"`
}

type MinimaxStreamChoice struct {
	Delta        string `json:"delta"`
	FinishReason string `json:"finish_reason"`
}

type MinimaxChatResponse struct {
	Id       string          `json:"id"`
	Created  int64           `json:"created"`
	Choices  []MinimaxChoice `json:"choices"`
	Usage    `json:"usage"`
	BaseResp MinimaxError `json:"base_resp"`
}

type MinimaxChatStreamResponse struct {
	Id       string                `json:"id"`
	Created  int64                 `json:"created"`
	Choices  []MinimaxStreamChoice `json:"choices"`
	Usage    `json:"usage"`
	BaseResp MinimaxError `json:"base_resp"`
}

type MinimaxEmbeddingRequest struct {
	Model string   `json:"model,omitempty"`
	Texts []string `json:"texts,omitempty"` //upper bound: 4096 tokens
	Type  string   `json:"type,omitempty"`  //
	// must choose one of the cases: {"db", "query"};
	// because of the default meaning of embedding request is "Creates an embedding vector representing the input text"
	// so we default use the "db" input to generate texts' embedding vector
	// for the "query" input, we will support later
	// Refer: https://api.minimax.chat/document/guides/embeddings?id=6464722084cdc277dfaa966a#%E6%8E%A5%E5%8F%A3%E5%8F%82%E6%95%B0%E8%AF%B4%E6%98%8E
}

type MinimaxEmbeddingResponse struct {
	Vectors  [][]float64  `json:"vectors"`
	BaseResp MinimaxError `json:"base_resp"`
}

func openAIMsgRoleToMinimaxMsgRole(input string) string {
	if input == "user" {
		return "USER"
	} else {
		return "BOT"
	}
}

func requestOpenAI2Minimax(request GeneralOpenAIRequest) *MinimaxChatRequest {
	messages := make([]MinimaxChatMessage, 0, len(request.Messages))
	prompt := ""
	for _, message := range request.Messages {
		if message.Role == "system" {
			prompt += message.Content
		} else {
			messages = append(messages, MinimaxChatMessage{
				SenderType: openAIMsgRoleToMinimaxMsgRole(message.Role),
				Text:       message.Content,
			})
		}
	}
	return &MinimaxChatRequest{
		Model:       request.Model,
		Stream:      request.Stream,
		Messages:    messages,
		Prompt:      prompt,
		Temperature: request.Temperature,
		TopP:        request.TopP,
	}
}

func responseMinimaxChat2OpenAI(response *MinimaxChatResponse) *OpenAITextResponse {
	ans := OpenAITextResponse{
		Id:      response.Id,
		Object:  "",
		Created: response.Created,
		Choices: make([]OpenAITextResponseChoice, 0, len(response.Choices)),
		Usage:   response.Usage,
	}
	for _, choice := range response.Choices {
		ans.Choices = append(ans.Choices, OpenAITextResponseChoice{
			Index: choice.Index,
			Message: Message{
				Role:    "assistant",
				Content: choice.Text,
			},
			FinishReason: choice.FinishReason,
		})
	}
	return &ans
}

func streamResponseMinimaxChat2OpenAI(response *MinimaxChatStreamResponse) *ChatCompletionsStreamResponse {
	ans := ChatCompletionsStreamResponse{
		Id:      response.Id,
		Object:  "chat.completion.chunk",
		Created: response.Created,
		Model:   "abab", //"abab5.5-chat", "abab5-chat"
		Choices: make([]ChatCompletionsStreamResponseChoice, 0, len(response.Choices)),
	}
	for i := range response.Choices {
		choice := response.Choices[i]
		ans.Choices = append(ans.Choices, ChatCompletionsStreamResponseChoice{
			Delta: struct {
				Content string `json:"content"`
			}{
				Content: choice.Delta,
			},
			FinishReason: &choice.FinishReason,
		})
	}
	return &ans
}

func embeddingRequestOpenAI2Minimax(request GeneralOpenAIRequest) *MinimaxEmbeddingRequest {
	texts := make([]string, 0, 100)
	v := reflect.ValueOf(request.Input)
	switch v.Kind() {
	case reflect.String:
		texts = []string{v.Interface().(string)}
	case reflect.Array, reflect.Slice:
		for i := 0; i < v.Len(); i++ {
			texts = append(texts, v.Index(i).Interface().(string))
		}
	}
	ans := MinimaxEmbeddingRequest{
		Model: request.Model,
		Texts: texts,
		Type:  "db",
	}
	return &ans
}

func embeddingResponseMinimax2OpenAI(response *MinimaxEmbeddingResponse) *OpenAIEmbeddingResponse {
	ans := OpenAIEmbeddingResponse{
		Object: "list",
		Data:   make([]OpenAIEmbeddingResponseItem, 0, len(response.Vectors)),
		Model:  "minimax-embedding",
	}
	for i, vector := range response.Vectors {
		ans.Data = append(ans.Data, OpenAIEmbeddingResponseItem{
			Object:    "embedding",
			Index:     i,
			Embedding: vector,
		})
	}
	return &ans
}

func minimaxHandler(c *gin.Context, resp *http.Response) (*OpenAIErrorWithStatusCode, *Usage) {
	var minimaxChatRsp MinimaxChatResponse
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return errorWrapper(err, "read_response_body_failed", http.StatusInternalServerError), nil
	}
	err = resp.Body.Close()
	if err != nil {
		return errorWrapper(err, "close_response_body_failed", http.StatusInternalServerError), nil
	}
	err = json.Unmarshal(responseBody, &minimaxChatRsp)
	if err != nil {
		return errorWrapper(err, "unmarshal_response_body_failed", http.StatusInternalServerError), nil
	}
	if minimaxChatRsp.BaseResp.StatusMsg != "success" {
		return &OpenAIErrorWithStatusCode{
			OpenAIError: OpenAIError{
				Message: minimaxChatRsp.BaseResp.StatusMsg,
				Type:    "minimax_error",
				Param:   "",
				Code:    minimaxChatRsp.BaseResp.StatusCode,
			},
			StatusCode: resp.StatusCode,
		}, nil
	}
	fullTextResponse := responseMinimaxChat2OpenAI(&minimaxChatRsp)
	jsonResponse, err := json.Marshal(fullTextResponse)
	if err != nil {
		return errorWrapper(err, "marshal_response_body_failed", http.StatusInternalServerError), nil
	}
	c.Writer.Header().Set("Content-Type", "application/json")
	c.Writer.WriteHeader(resp.StatusCode)
	_, err = c.Writer.Write(jsonResponse)
	return nil, &fullTextResponse.Usage
}

func minimaxStreamHandler(c *gin.Context, resp *http.Response) (*OpenAIErrorWithStatusCode, *Usage) {
	var usage Usage
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
	dataChan := make(chan string, 100)
	go func() {
		for scanner.Scan() {
			data := scanner.Text()
			if len(data) < 6 { // ignore blank line or wrong format
				continue
			}
			if data[:6] != "data: " {
				continue
			}
			data = data[6:]
			dataChan <- data
		}
		close(dataChan)
	}()

	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Header().Set("Transfer-Encoding", "chunked")
	c.Writer.Header().Set("X-Accel-Buffering", "no")
	c.Stream(func(w io.Writer) bool {
		if data, ok := <-dataChan; ok {
			var minimaxChatStreamRsp MinimaxChatStreamResponse
			err := json.Unmarshal([]byte(data), &minimaxChatStreamRsp)
			if err != nil {
				common.SysError("error unmarshalling stream response: " + err.Error())
				return true
			}
			usage.TotalTokens += minimaxChatStreamRsp.TotalTokens
			response := streamResponseMinimaxChat2OpenAI(&minimaxChatStreamRsp)
			jsonResponse, err := json.Marshal(response)
			if err != nil {
				common.SysError("error marshalling stream response: " + err.Error())
				return true
			}
			c.Render(-1, common.CustomEvent{Data: "data: " + string(jsonResponse)})
			return true
		}
		return false
	})
	err := resp.Body.Close()
	if err != nil {
		return errorWrapper(err, "close_response_body_failed", http.StatusInternalServerError), nil
	}
	return nil, &usage
}

func minimaxEmbeddingHandler(c *gin.Context, resp *http.Response, promptTokens int, model string) (*OpenAIErrorWithStatusCode, *Usage) {
	var minimaxEmbeddingRsp MinimaxEmbeddingResponse
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return errorWrapper(err, "read_response_body_failed", http.StatusInternalServerError), nil
	}
	err = resp.Body.Close()
	if err != nil {
		return errorWrapper(err, "close_response_body_failed", http.StatusInternalServerError), nil
	}
	err = json.Unmarshal(responseBody, &minimaxEmbeddingRsp)
	if err != nil {
		return errorWrapper(err, "unmarshal_response_body_failed", http.StatusInternalServerError), nil
	}

	if minimaxEmbeddingRsp.BaseResp.StatusMsg != "success" {
		return &OpenAIErrorWithStatusCode{
			OpenAIError: OpenAIError{
				Message: minimaxEmbeddingRsp.BaseResp.StatusMsg,
				Type:    "minimax_error",
				Param:   "",
				Code:    minimaxEmbeddingRsp.BaseResp.StatusCode,
			},
			StatusCode: resp.StatusCode,
		}, nil
	}
	fullTextResponse := embeddingResponseMinimax2OpenAI(&minimaxEmbeddingRsp)
	fullTextResponse.Usage = Usage{
		PromptTokens:     promptTokens,
		CompletionTokens: 0,
		TotalTokens:      promptTokens,
	}
	jsonResponse, err := json.Marshal(fullTextResponse)
	if err != nil {
		return errorWrapper(err, "marshal_response_body_failed", http.StatusInternalServerError), nil
	}
	c.Writer.Header().Set("Content-Type", "application/json")
	c.Writer.WriteHeader(resp.StatusCode)
	_, err = c.Writer.Write(jsonResponse)
	return nil, &fullTextResponse.Usage
}
