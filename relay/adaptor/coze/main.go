package coze

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/songquanpeng/one-api/common"
	"github.com/songquanpeng/one-api/common/conv"
	"github.com/songquanpeng/one-api/common/helper"
	"github.com/songquanpeng/one-api/common/logger"
	"github.com/songquanpeng/one-api/relay/adaptor/coze/constant/messagetype"
	"github.com/songquanpeng/one-api/relay/adaptor/openai"
	"github.com/songquanpeng/one-api/relay/model"
	"io"
	"net/http"
	"strings"
)

// https://www.coze.com/open

func stopReasonCoze2OpenAI(reason *string) string {
	if reason == nil {
		return ""
	}
	switch *reason {
	case "end_turn":
		return "stop"
	case "stop_sequence":
		return "stop"
	case "max_tokens":
		return "length"
	default:
		return *reason
	}
}

func ConvertRequest(textRequest model.GeneralOpenAIRequest) *Request {
	cozeRequest := Request{
		Stream: textRequest.Stream,
		User:   textRequest.User,
		BotId:  strings.TrimPrefix(textRequest.Model, "bot-"),
	}
	for i, message := range textRequest.Messages {
		if i == len(textRequest.Messages)-1 {
			cozeRequest.Query = message.StringContent()
			continue
		}
		cozeMessage := Message{
			Role:    message.Role,
			Content: message.StringContent(),
		}
		cozeRequest.ChatHistory = append(cozeRequest.ChatHistory, cozeMessage)
	}
	return &cozeRequest
}

func StreamResponseCoze2OpenAI(cozeResponse *StreamResponse) (*openai.ChatCompletionsStreamResponse, *Response) {
	var response *Response
	var stopReason string
	var choice openai.ChatCompletionsStreamResponseChoice

	if cozeResponse.Message != nil {
		if cozeResponse.Message.Type != messagetype.Answer {
			return nil, nil
		}
		choice.Delta.Content = cozeResponse.Message.Content
	}
	choice.Delta.Role = "assistant"
	finishReason := stopReasonCoze2OpenAI(&stopReason)
	if finishReason != "null" {
		choice.FinishReason = &finishReason
	}
	var openaiResponse openai.ChatCompletionsStreamResponse
	openaiResponse.Object = "chat.completion.chunk"
	openaiResponse.Choices = []openai.ChatCompletionsStreamResponseChoice{choice}
	openaiResponse.Id = cozeResponse.ConversationId
	return &openaiResponse, response
}

func ResponseCoze2OpenAI(cozeResponse *Response) *openai.TextResponse {
	var responseText string
	for _, message := range cozeResponse.Messages {
		if message.Type == messagetype.Answer {
			responseText = message.Content
			break
		}
	}
	choice := openai.TextResponseChoice{
		Index: 0,
		Message: model.Message{
			Role:    "assistant",
			Content: responseText,
			Name:    nil,
		},
		FinishReason: "stop",
	}
	fullTextResponse := openai.TextResponse{
		Id:      fmt.Sprintf("chatcmpl-%s", cozeResponse.ConversationId),
		Model:   "coze-bot",
		Object:  "chat.completion",
		Created: helper.GetTimestamp(),
		Choices: []openai.TextResponseChoice{choice},
	}
	return &fullTextResponse
}

func StreamHandler(c *gin.Context, resp *http.Response) (*model.ErrorWithStatusCode, *string) {
	var responseText string
	createdTime := helper.GetTimestamp()
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
			if len(data) < 5 {
				continue
			}
			if !strings.HasPrefix(data, "data:") {
				continue
			}
			data = strings.TrimPrefix(data, "data:")
			dataChan <- data
		}
		stopChan <- true
	}()
	common.SetEventStreamHeaders(c)
	var modelName string
	c.Stream(func(w io.Writer) bool {
		select {
		case data := <-dataChan:
			// some implementations may add \r at the end of data
			data = strings.TrimSuffix(data, "\r")
			var cozeResponse StreamResponse
			err := json.Unmarshal([]byte(data), &cozeResponse)
			if err != nil {
				logger.SysError("error unmarshalling stream response: " + err.Error())
				return true
			}
			response, _ := StreamResponseCoze2OpenAI(&cozeResponse)
			if response == nil {
				return true
			}
			for _, choice := range response.Choices {
				responseText += conv.AsString(choice.Delta.Content)
			}
			response.Model = modelName
			response.Created = createdTime
			jsonStr, err := json.Marshal(response)
			if err != nil {
				logger.SysError("error marshalling stream response: " + err.Error())
				return true
			}
			c.Render(-1, common.CustomEvent{Data: "data: " + string(jsonStr)})
			return true
		case <-stopChan:
			c.Render(-1, common.CustomEvent{Data: "data: [DONE]"})
			return false
		}
	})
	_ = resp.Body.Close()
	return nil, &responseText
}

func Handler(c *gin.Context, resp *http.Response, promptTokens int, modelName string) (*model.ErrorWithStatusCode, *string) {
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return openai.ErrorWrapper(err, "read_response_body_failed", http.StatusInternalServerError), nil
	}
	err = resp.Body.Close()
	if err != nil {
		return openai.ErrorWrapper(err, "close_response_body_failed", http.StatusInternalServerError), nil
	}
	var cozeResponse Response
	err = json.Unmarshal(responseBody, &cozeResponse)
	if err != nil {
		return openai.ErrorWrapper(err, "unmarshal_response_body_failed", http.StatusInternalServerError), nil
	}
	if cozeResponse.Code != 0 {
		return &model.ErrorWithStatusCode{
			Error: model.Error{
				Message: cozeResponse.Msg,
				Code:    cozeResponse.Code,
			},
			StatusCode: resp.StatusCode,
		}, nil
	}
	fullTextResponse := ResponseCoze2OpenAI(&cozeResponse)
	fullTextResponse.Model = modelName
	jsonResponse, err := json.Marshal(fullTextResponse)
	if err != nil {
		return openai.ErrorWrapper(err, "marshal_response_body_failed", http.StatusInternalServerError), nil
	}
	c.Writer.Header().Set("Content-Type", "application/json")
	c.Writer.WriteHeader(resp.StatusCode)
	_, err = c.Writer.Write(jsonResponse)
	var responseText string
	if len(fullTextResponse.Choices) > 0 {
		responseText = fullTextResponse.Choices[0].Message.StringContent()
	}
	return nil, &responseText
}
