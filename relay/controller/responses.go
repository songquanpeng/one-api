package controller

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/songquanpeng/one-api/common"
	"github.com/songquanpeng/one-api/common/logger"
	"github.com/songquanpeng/one-api/common/random"
	"github.com/songquanpeng/one-api/common/render"
	"github.com/songquanpeng/one-api/relay/adaptor/openai"
	"github.com/songquanpeng/one-api/relay/constant/role"
	"github.com/songquanpeng/one-api/relay/meta"
	relaymodel "github.com/songquanpeng/one-api/relay/model"
)

const (
	responsesDataPrefix = "data: "
	responsesDone       = "[DONE]"
)

// ========== 请求转换：Responses API -> Chat Completions ==========

// getAndValidateResponsesRequest 解析 Responses API 请求体并转换为内部 Chat Completions 请求
func getAndValidateResponsesRequest(c *gin.Context) (*relaymodel.GeneralOpenAIRequest, error) {
	responsesRequest := &relaymodel.ResponsesRequest{}
	if err := common.UnmarshalBodyReusable(c, responsesRequest); err != nil {
		return nil, err
	}
	return responsesRequestToChatRequest(responsesRequest)
}

func responsesRequestToChatRequest(req *relaymodel.ResponsesRequest) (*relaymodel.GeneralOpenAIRequest, error) {
	if req.Model == "" {
		return nil, errors.New("model is required")
	}
	chatRequest := &relaymodel.GeneralOpenAIRequest{
		Model:            req.Model,
		Stream:           req.Stream,
		StreamOptions:    req.StreamOptions,
		Temperature:      req.Temperature,
		TopP:             req.TopP,
		Tools:            req.Tools,
		ToolChoice:       req.ToolChoice,
		ParallelTooCalls: req.ParallelToolCalls,
		Metadata:         req.Metadata,
		User:             req.User,
	}
	if req.MaxOutputTokens != nil {
		chatRequest.MaxTokens = *req.MaxOutputTokens
	}
	if req.Instructions != "" {
		chatRequest.Messages = append(chatRequest.Messages, relaymodel.Message{
			Role:    role.System,
			Content: req.Instructions,
		})
	}
	messages, err := responsesInputToMessages(req.Input)
	if err != nil {
		return nil, err
	}
	chatRequest.Messages = append(chatRequest.Messages, messages...)
	if len(chatRequest.Messages) == 0 {
		return nil, errors.New("input or instructions is required")
	}
	return chatRequest, nil
}

func responsesInputToMessages(input any) ([]relaymodel.Message, error) {
	if input == nil {
		return nil, nil
	}
	switch v := input.(type) {
	case string:
		return []relaymodel.Message{{Role: "user", Content: v}}, nil
	case []any:
		messages := make([]relaymodel.Message, 0, len(v))
		for _, item := range v {
			itemMap, ok := item.(map[string]any)
			if !ok {
				return nil, fmt.Errorf("invalid input item: %v", item)
			}
			message, err := responsesInputItemToMessage(itemMap)
			if err != nil {
				return nil, err
			}
			if message != nil {
				messages = append(messages, *message)
			}
		}
		return messages, nil
	case map[string]any:
		message, err := responsesInputItemToMessage(v)
		if err != nil {
			return nil, err
		}
		if message == nil {
			return nil, nil
		}
		return []relaymodel.Message{*message}, nil
	default:
		return nil, fmt.Errorf("unsupported input type: %T", input)
	}
}

func responsesInputItemToMessage(item map[string]any) (*relaymodel.Message, error) {
	switch item["type"] {
	case nil, "message":
		roleName, _ := item["role"].(string)
		if roleName == "" {
			roleName = "user"
		}
		content, err := responsesContentToChatContent(item["content"])
		if err != nil {
			return nil, err
		}
		return &relaymodel.Message{Role: roleName, Content: content}, nil
	case "function_call":
		callID, _ := item["call_id"].(string)
		name, _ := item["name"].(string)
		arguments, _ := item["arguments"].(string)
		return &relaymodel.Message{
			Role: role.Assistant,
			ToolCalls: []relaymodel.Tool{{
				Id:       callID,
				Type:     "function",
				Function: relaymodel.Function{Name: name, Arguments: arguments},
			}},
		}, nil
	case "function_call_output":
		callID, _ := item["call_id"].(string)
		output, _ := item["output"].(string)
		return &relaymodel.Message{
			Role:       "tool",
			ToolCallId: callID,
			Content:    output,
		}, nil
	default:
		// item_reference 等类型无法映射为消息，忽略
		return nil, nil
	}
}

// responsesContentToChatContent 将 Responses 的 content 转换为 Chat Completions 的 content
func responsesContentToChatContent(content any) (any, error) {
	if content == nil {
		return "", nil
	}
	switch v := content.(type) {
	case string:
		return v, nil
	case []any:
		parts := make([]any, 0, len(v))
		for _, part := range v {
			partMap, ok := part.(map[string]any)
			if !ok {
				continue
			}
			switch partMap["type"] {
			case "input_text", "output_text", "text":
				text, _ := partMap["text"].(string)
				parts = append(parts, map[string]any{"type": "text", "text": text})
			case "input_image", "image_url":
				if imageURL, ok := partMap["image_url"].(map[string]any); ok {
					parts = append(parts, map[string]any{"type": "image_url", "image_url": imageURL})
				}
			}
		}
		if len(parts) == 1 {
			if textPart, ok := parts[0].(map[string]any); ok && textPart["type"] == "text" {
				return textPart["text"], nil
			}
		}
		return parts, nil
	default:
		return nil, fmt.Errorf("unsupported content type: %T", content)
	}
}

// ========== 响应转换：Chat Completions -> Responses API ==========

// responsesDoResponse Responses 模式的响应处理入口
func responsesDoResponse(c *gin.Context, resp *http.Response, meta *meta.Meta) (*relaymodel.Usage, *relaymodel.ErrorWithStatusCode) {
	if meta.IsStream {
		return responsesStreamHandler(c, resp, meta)
	}
	return responsesHandler(c, resp, meta)
}

func responsesHandler(c *gin.Context, resp *http.Response, meta *meta.Meta) (*relaymodel.Usage, *relaymodel.ErrorWithStatusCode) {
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, openai.ErrorWrapper(err, "read_response_body_failed", http.StatusInternalServerError)
	}
	if err = resp.Body.Close(); err != nil {
		return nil, openai.ErrorWrapper(err, "close_response_body_failed", http.StatusInternalServerError)
	}
	var textResponse openai.SlimTextResponse
	if err = json.Unmarshal(responseBody, &textResponse); err != nil {
		return nil, openai.ErrorWrapper(err, "unmarshal_response_body_failed", http.StatusInternalServerError)
	}
	if textResponse.Error.Message != "" {
		return nil, openai.ErrorWrapper(errors.New(textResponse.Error.Message), "convert_response_failed", http.StatusInternalServerError)
	}
	responsesResponse := chatResponseToResponsesResponse(&textResponse, meta)
	c.JSON(http.StatusOK, responsesResponse)
	return &textResponse.Usage, nil
}

func chatResponseToResponsesResponse(textResponse *openai.SlimTextResponse, meta *meta.Meta) *relaymodel.ResponsesResponse {
	response := &relaymodel.ResponsesResponse{
		Id:        "resp_" + random.GetUUID(),
		Object:    "response",
		CreatedAt: time.Now().Unix(),
		Status:    "completed",
		Model:     meta.ActualModelName,
		Output:    []relaymodel.ResponsesOutputItem{},
	}
	if len(textResponse.Choices) > 0 {
		message := textResponse.Choices[0].Message
		content := message.StringContent()
		if content != "" || len(message.ToolCalls) == 0 {
			response.Output = append(response.Output, relaymodel.ResponsesOutputItem{
				Type:    "message",
				Id:      "msg_" + random.GetUUID(),
				Role:    role.Assistant,
				Status:  "completed",
				Content: []relaymodel.ResponsesOutputTextContent{{Type: "output_text", Text: content}},
			})
		}
		for _, toolCall := range message.ToolCalls {
			arguments, _ := toolCall.Function.Arguments.(string)
			response.Output = append(response.Output, relaymodel.ResponsesOutputItem{
				Type:      "function_call",
				Id:        "fc_" + random.GetUUID(),
				Status:    "completed",
				Name:      toolCall.Function.Name,
				Arguments: arguments,
				CallID:    toolCall.Id,
			})
		}
	}
	response.Usage = &relaymodel.ResponsesUsage{
		InputTokens:  textResponse.Usage.PromptTokens,
		OutputTokens: textResponse.Usage.CompletionTokens,
		TotalTokens:  textResponse.Usage.TotalTokens,
	}
	return response
}

// ========== 流式转换：Chat Completions SSE -> Responses SSE ==========

type responsesStreamToolCall struct {
	id        string
	name      string
	arguments strings.Builder
}

type responsesStreamChunk struct {
	Id      string `json:"id"`
	Model   string `json:"model"`
	Created int64  `json:"created"`
	Choices []struct {
		Index int `json:"index"`
		Delta struct {
			Role      string `json:"role,omitempty"`
			Content   any    `json:"content,omitempty"`
			ToolCalls []struct {
				Index    int    `json:"index"`
				Id       string `json:"id,omitempty"`
				Type     string `json:"type,omitempty"`
				Function struct {
					Name      string `json:"name,omitempty"`
					Arguments string `json:"arguments,omitempty"`
				} `json:"function"`
			} `json:"tool_calls,omitempty"`
		} `json:"delta"`
		FinishReason *string `json:"finish_reason,omitempty"`
	} `json:"choices"`
	Usage *relaymodel.Usage `json:"usage,omitempty"`
}

func responsesStreamHandler(c *gin.Context, resp *http.Response, meta *meta.Meta) (*relaymodel.Usage, *relaymodel.ErrorWithStatusCode) {
	common.SetEventStreamHeaders(c)

	responseId := "resp_" + random.GetUUID()
	createdAt := time.Now().Unix()
	sequenceNumber := 0

	emitEvent := func(event relaymodel.ResponsesStreamEvent) {
		event.SequenceNumber = sequenceNumber
		sequenceNumber++
		if err := render.ObjectData(c, event); err != nil {
			logger.SysError("error rendering responses stream event: " + err.Error())
		}
	}

	emptyResponse := func(status string) *relaymodel.ResponsesResponse {
		return &relaymodel.ResponsesResponse{
			Id:        responseId,
			Object:    "response",
			CreatedAt: createdAt,
			Status:    status,
			Model:     meta.ActualModelName,
			Output:    []relaymodel.ResponsesOutputItem{},
		}
	}

	emitEvent(relaymodel.ResponsesStreamEvent{Type: "response.created", Response: emptyResponse("in_progress")})
	emitEvent(relaymodel.ResponsesStreamEvent{Type: "response.in_progress", Response: emptyResponse("in_progress")})

	var usage *relaymodel.Usage
	var responseText strings.Builder
	var textStarted bool
	var toolCalls []*responsesStreamToolCall
	toolCallMap := make(map[int]*responsesStreamToolCall)

	scanner := bufio.NewScanner(resp.Body)
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		data := scanner.Text()
		if len(data) < len(responsesDataPrefix) || data[:len(responsesDataPrefix)] != responsesDataPrefix {
			continue // ignore blank line or wrong format
		}
		payload := strings.TrimSuffix(strings.TrimPrefix(data, responsesDataPrefix), "\r")
		if payload == responsesDone {
			continue
		}
		var chunk responsesStreamChunk
		if err := json.Unmarshal([]byte(payload), &chunk); err != nil {
			logger.SysError("error unmarshalling responses stream chunk: " + err.Error())
			continue
		}
		if chunk.Usage != nil {
			usage = chunk.Usage
		}
		for _, choice := range chunk.Choices {
			if content, ok := choice.Delta.Content.(string); ok && content != "" {
				if !textStarted {
					textStarted = true
					emitEvent(relaymodel.ResponsesStreamEvent{
						Type:        "response.output_item.added",
						OutputIndex: 0,
						Item: &relaymodel.ResponsesOutputItem{
							Type:    "message",
							Id:      "msg_" + random.GetUUID(),
							Role:    role.Assistant,
							Status:  "in_progress",
							Content: []relaymodel.ResponsesOutputTextContent{},
						},
					})
					emitEvent(relaymodel.ResponsesStreamEvent{
						Type:         "response.content_part.added",
						OutputIndex:  0,
						ContentIndex: 0,
						Part:         &relaymodel.ResponsesOutputTextContent{Type: "output_text", Text: ""},
					})
				}
				responseText.WriteString(content)
				emitEvent(relaymodel.ResponsesStreamEvent{
					Type:         "response.output_text.delta",
					OutputIndex:  0,
					ContentIndex: 0,
					Delta:        content,
				})
			}
			for _, toolCall := range choice.Delta.ToolCalls {
				acc, ok := toolCallMap[toolCall.Index]
				if !ok {
					acc = &responsesStreamToolCall{}
					toolCallMap[toolCall.Index] = acc
					toolCalls = append(toolCalls, acc)
				}
				if toolCall.Id != "" {
					acc.id = toolCall.Id
				}
				if toolCall.Function.Name != "" {
					acc.name = toolCall.Function.Name
				}
				acc.arguments.WriteString(toolCall.Function.Arguments)
			}
		}
	}
	if err := scanner.Err(); err != nil {
		logger.SysError("error reading responses stream: " + err.Error())
	}
	if err := resp.Body.Close(); err != nil {
		return nil, openai.ErrorWrapper(err, "close_response_body_failed", http.StatusInternalServerError)
	}

	finalText := responseText.String()
	if textStarted {
		emitEvent(relaymodel.ResponsesStreamEvent{Type: "response.output_text.done", OutputIndex: 0, ContentIndex: 0, Text: finalText})
		emitEvent(relaymodel.ResponsesStreamEvent{
			Type:         "response.content_part.done",
			OutputIndex:  0,
			ContentIndex: 0,
			Part:         &relaymodel.ResponsesOutputTextContent{Type: "output_text", Text: finalText},
		})
		emitEvent(relaymodel.ResponsesStreamEvent{
			Type:        "response.output_item.done",
			OutputIndex: 0,
			Item: &relaymodel.ResponsesOutputItem{
				Type:    "message",
				Id:      "msg_" + random.GetUUID(),
				Role:    role.Assistant,
				Status:  "completed",
				Content: []relaymodel.ResponsesOutputTextContent{{Type: "output_text", Text: finalText}},
			},
		})
	}
	for i, toolCall := range toolCalls {
		outputIndex := i + 1
		arguments := toolCall.arguments.String()
		itemId := "fc_" + random.GetUUID()
		emitEvent(relaymodel.ResponsesStreamEvent{
			Type:        "response.output_item.added",
			OutputIndex: outputIndex,
			Item: &relaymodel.ResponsesOutputItem{
				Type:   "function_call",
				Id:     itemId,
				Status: "in_progress",
				Name:   toolCall.name,
				CallID: toolCall.id,
			},
		})
		emitEvent(relaymodel.ResponsesStreamEvent{
			Type:        "response.function_call_arguments.delta",
			OutputIndex: outputIndex,
			Item: &relaymodel.ResponsesOutputItem{
				Type:   "function_call",
				Id:     itemId,
				Status: "in_progress",
				Name:   toolCall.name,
				CallID: toolCall.id,
			},
			Delta: arguments,
		})
		emitEvent(relaymodel.ResponsesStreamEvent{
			Type:        "response.output_item.done",
			OutputIndex: outputIndex,
			Item: &relaymodel.ResponsesOutputItem{
				Type:      "function_call",
				Id:        itemId,
				Status:    "completed",
				Name:      toolCall.name,
				Arguments: arguments,
				CallID:    toolCall.id,
			},
		})
	}

	finalResponse := emptyResponse("completed")
	if finalText != "" {
		finalResponse.Output = append(finalResponse.Output, relaymodel.ResponsesOutputItem{
			Type:    "message",
			Id:      "msg_" + random.GetUUID(),
			Role:    role.Assistant,
			Status:  "completed",
			Content: []relaymodel.ResponsesOutputTextContent{{Type: "output_text", Text: finalText}},
		})
	}
	for _, toolCall := range toolCalls {
		finalResponse.Output = append(finalResponse.Output, relaymodel.ResponsesOutputItem{
			Type:      "function_call",
			Id:        "fc_" + random.GetUUID(),
			Status:    "completed",
			Name:      toolCall.name,
			Arguments: toolCall.arguments.String(),
			CallID:    toolCall.id,
		})
	}
	if usage != nil {
		finalResponse.Usage = &relaymodel.ResponsesUsage{
			InputTokens:  usage.PromptTokens,
			OutputTokens: usage.CompletionTokens,
			TotalTokens:  usage.TotalTokens,
		}
	} else {
		finalResponse.Usage = &relaymodel.ResponsesUsage{
			InputTokens:  meta.PromptTokens,
			OutputTokens: 0,
			TotalTokens:  meta.PromptTokens,
		}
	}
	emitEvent(relaymodel.ResponsesStreamEvent{Type: "response.completed", Response: finalResponse})
	return usage, nil
}
