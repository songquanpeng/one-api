package anthropic

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/songquanpeng/one-api/common"
	"github.com/songquanpeng/one-api/common/helper"
	"github.com/songquanpeng/one-api/common/image"
	"github.com/songquanpeng/one-api/common/logger"
	"github.com/songquanpeng/one-api/relay/adaptor/openai"
	"github.com/songquanpeng/one-api/relay/model"
)

func stopReasonClaude2OpenAI(reason *string) string {
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
	claudeRequest := Request{
		Model:       textRequest.Model,
		MaxTokens:   textRequest.MaxTokens,
		Temperature: textRequest.Temperature,
		TopP:        textRequest.TopP,
		TopK:        textRequest.TopK,
		Stream:      textRequest.Stream,
	}
	if claudeRequest.MaxTokens == 0 {
		claudeRequest.MaxTokens = 4096
	}
	// legacy model name mapping
	if claudeRequest.Model == "claude-instant-1" {
		claudeRequest.Model = "claude-instant-1.1"
	} else if claudeRequest.Model == "claude-2" {
		claudeRequest.Model = "claude-2.1"
	}
	for _, message := range textRequest.Messages {
		if message.Role == "system" && claudeRequest.System == "" {
			claudeRequest.System = message.StringContent()
			continue
		}
		claudeMessage := Message{
			Role: message.Role,
		}
		var content Content
		if message.IsStringContent() {
			content.Type = "text"
			content.Text = message.StringContent()
			claudeMessage.Content = append(claudeMessage.Content, content)
			claudeRequest.Messages = append(claudeRequest.Messages, claudeMessage)
			continue
		}
		var contents []Content
		openaiContent := message.ParseContent()
		for _, part := range openaiContent {
			var content Content
			if part.Type == model.ContentTypeText {
				content.Type = "text"
				content.Text = part.Text
			} else if part.Type == model.ContentTypeImageURL {
				content.Type = "image"
				content.Source = &ImageSource{
					Type: "base64",
				}
				mimeType, data, _ := image.GetImageFromUrl(part.ImageURL.Url)
				content.Source.MediaType = mimeType
				content.Source.Data = data
			}
			contents = append(contents, content)
		}
		claudeMessage.Content = contents
		claudeRequest.Messages = append(claudeRequest.Messages, claudeMessage)
	}
	return &claudeRequest
}

// https://docs.anthropic.com/claude/reference/messages-streaming
func StreamResponseClaude2OpenAI(claudeResponse *StreamResponse) (*openai.ChatCompletionsStreamResponse, *Response) {
	var response *Response
	var responseText string
	var stopReason string
	switch claudeResponse.Type {
	case "message_start":
		return nil, claudeResponse.Message
	case "content_block_start":
		if claudeResponse.ContentBlock != nil {
			responseText = claudeResponse.ContentBlock.Text
		}
	case "content_block_delta":
		if claudeResponse.Delta != nil {
			responseText = claudeResponse.Delta.Text
		}
	case "message_delta":
		if claudeResponse.Usage != nil {
			response = &Response{
				Usage: *claudeResponse.Usage,
			}
		}
		if claudeResponse.Delta != nil && claudeResponse.Delta.StopReason != nil {
			stopReason = *claudeResponse.Delta.StopReason
		}
	}
	var choice openai.ChatCompletionsStreamResponseChoice
	choice.Delta.Content = responseText
	choice.Delta.Role = "assistant"
	finishReason := stopReasonClaude2OpenAI(&stopReason)
	if finishReason != "null" {
		choice.FinishReason = &finishReason
	}
	var openaiResponse openai.ChatCompletionsStreamResponse
	openaiResponse.Object = "chat.completion.chunk"
	openaiResponse.Choices = []openai.ChatCompletionsStreamResponseChoice{choice}
	return &openaiResponse, response
}

func ResponseClaude2OpenAI(claudeResponse *Response) *openai.TextResponse {
	var responseText string
	if len(claudeResponse.Content) > 0 {
		responseText = claudeResponse.Content[0].Text
	}
	choice := openai.TextResponseChoice{
		Index: 0,
		Message: model.Message{
			Role:    "assistant",
			Content: responseText,
			Name:    nil,
		},
		FinishReason: stopReasonClaude2OpenAI(claudeResponse.StopReason),
	}
	fullTextResponse := openai.TextResponse{
		Id:      fmt.Sprintf("chatcmpl-%s", claudeResponse.Id),
		Model:   claudeResponse.Model,
		Object:  "chat.completion",
		Created: helper.GetTimestamp(),
		Choices: []openai.TextResponseChoice{choice},
	}
	return &fullTextResponse
}

func StreamHandler(c *gin.Context, resp *http.Response) (*model.ErrorWithStatusCode, *model.Usage) {
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
			if len(data) < 6 {
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
	var usage model.Usage
	var modelName string
	var id string
	c.Stream(func(w io.Writer) bool {
		select {
		case data := <-dataChan:
			// some implementations may add \r at the end of data
			data = strings.TrimSpace(data)
			var claudeResponse StreamResponse
			err := json.Unmarshal([]byte(data), &claudeResponse)
			if err != nil {
				logger.SysError("error unmarshalling stream response: " + err.Error())
				return true
			}
			response, meta := StreamResponseClaude2OpenAI(&claudeResponse)
			if meta != nil {
				usage.PromptTokens += meta.Usage.InputTokens
				usage.CompletionTokens += meta.Usage.OutputTokens
				modelName = meta.Model
				id = fmt.Sprintf("chatcmpl-%s", meta.Id)
				return true
			}
			if response == nil {
				return true
			}
			response.Id = id
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
	return nil, &usage
}

func Handler(c *gin.Context, resp *http.Response, promptTokens int, modelName string) (*model.ErrorWithStatusCode, *model.Usage) {
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return openai.ErrorWrapper(err, "read_response_body_failed", http.StatusInternalServerError), nil
	}
	err = resp.Body.Close()
	if err != nil {
		return openai.ErrorWrapper(err, "close_response_body_failed", http.StatusInternalServerError), nil
	}
	var claudeResponse Response
	err = json.Unmarshal(responseBody, &claudeResponse)
	if err != nil {
		return openai.ErrorWrapper(err, "unmarshal_response_body_failed", http.StatusInternalServerError), nil
	}
	if claudeResponse.Error.Type != "" {
		return &model.ErrorWithStatusCode{
			Error: model.Error{
				Message: claudeResponse.Error.Message,
				Type:    claudeResponse.Error.Type,
				Param:   "",
				Code:    claudeResponse.Error.Type,
			},
			StatusCode: resp.StatusCode,
		}, nil
	}
	fullTextResponse := ResponseClaude2OpenAI(&claudeResponse)
	fullTextResponse.Model = modelName
	usage := model.Usage{
		PromptTokens:     claudeResponse.Usage.InputTokens,
		CompletionTokens: claudeResponse.Usage.OutputTokens,
		TotalTokens:      claudeResponse.Usage.InputTokens + claudeResponse.Usage.OutputTokens,
	}
	fullTextResponse.Usage = usage
	jsonResponse, err := json.Marshal(fullTextResponse)
	if err != nil {
		return openai.ErrorWrapper(err, "marshal_response_body_failed", http.StatusInternalServerError), nil
	}
	c.Writer.Header().Set("Content-Type", "application/json")
	c.Writer.WriteHeader(resp.StatusCode)
	_, err = c.Writer.Write(jsonResponse)
	return nil, &usage
}
