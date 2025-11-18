package anthropic

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/songquanpeng/one-api/common/render"

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
	case "tool_use":
		return "tool_calls"
	default:
		return *reason
	}
}

func ConvertRequest(textRequest model.GeneralOpenAIRequest) *Request {
	claudeTools := make([]Tool, 0, len(textRequest.Tools))

	for _, tool := range textRequest.Tools {
		if params, ok := tool.Function.Parameters.(map[string]any); ok {
			claudeTools = append(claudeTools, Tool{
				Name:        tool.Function.Name,
				Description: tool.Function.Description,
				InputSchema: InputSchema{
					Type:       params["type"].(string),
					Properties: params["properties"],
					Required:   params["required"],
				},
			})
		}
	}

	claudeRequest := Request{
		Model:       textRequest.Model,
		MaxTokens:   textRequest.MaxTokens,
		Temperature: textRequest.Temperature,
		TopP:        textRequest.TopP,
		TopK:        textRequest.TopK,
		Stream:      textRequest.Stream,
		Tools:       claudeTools,
	}
	if len(claudeTools) > 0 {
		claudeToolChoice := struct {
			Type string `json:"type"`
			Name string `json:"name,omitempty"`
		}{Type: "auto"} // default value https://docs.anthropic.com/en/docs/build-with-claude/tool-use#controlling-claudes-output
		if choice, ok := textRequest.ToolChoice.(map[string]any); ok {
			if function, ok := choice["function"].(map[string]any); ok {
				claudeToolChoice.Type = "tool"
				claudeToolChoice.Name = function["name"].(string)
			}
		} else if toolChoiceType, ok := textRequest.ToolChoice.(string); ok {
			if toolChoiceType == "any" {
				claudeToolChoice.Type = toolChoiceType
			}
		}
		claudeRequest.ToolChoice = claudeToolChoice
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
		if message.Role == "system" && claudeRequest.System.IsEmpty() {
			// Create a SystemPrompt from the string content
			systemPrompt := SystemPrompt{}
			systemData, err := json.Marshal(message.StringContent()) // Safely escape string for JSON
			if err != nil {
				logger.SysError(fmt.Sprintf("Failed to marshal system prompt: %v", err))
			} else {
				if err := systemPrompt.UnmarshalJSON(systemData); err != nil {
					logger.SysError(fmt.Sprintf("Failed to unmarshal system prompt: %v", err))
				}
				claudeRequest.System = systemPrompt
			}
			continue
		}
		claudeMessage := Message{
			Role: message.Role,
		}
		var contents []Content
		if message.IsStringContent() {
			content := Content{
				Type: "text",
				Text: message.StringContent(),
			}
			if message.Role == "tool" {
				claudeMessage.Role = "user"
				content.Type = "tool_result"
				content.Content = content.Text
				content.Text = ""
				content.ToolUseId = message.ToolCallId
			}
			contents = append(contents, content)
			for i := range message.ToolCalls {
				inputParam := make(map[string]any)
				_ = json.Unmarshal([]byte(message.ToolCalls[i].Function.Arguments.(string)), &inputParam)
				contents = append(contents, Content{
					Type:  "tool_use",
					Id:    message.ToolCalls[i].Id,
					Name:  message.ToolCalls[i].Function.Name,
					Input: inputParam,
				})
			}
			claudeMessage.Content = MessageContent{value: contents}
			claudeRequest.Messages = append(claudeRequest.Messages, claudeMessage)
			continue
		}
		contents = []Content{} // Reset the slice for reuse
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
		claudeMessage.Content = MessageContent{value: contents}
		claudeRequest.Messages = append(claudeRequest.Messages, claudeMessage)
	}
	return &claudeRequest
}

// https://docs.anthropic.com/claude/reference/messages-streaming
func StreamResponseClaude2OpenAI(claudeResponse *StreamResponse) (*openai.ChatCompletionsStreamResponse, *Response) {
	var response *Response
	var responseText string
	var stopReason string
	tools := make([]model.Tool, 0)

	switch claudeResponse.Type {
	case "message_start":
		return nil, claudeResponse.Message
	case "content_block_start":
		if claudeResponse.ContentBlock != nil {
			responseText = claudeResponse.ContentBlock.Text
			if claudeResponse.ContentBlock.Type == "tool_use" {
				tools = append(tools, model.Tool{
					Id:   claudeResponse.ContentBlock.Id,
					Type: "function",
					Function: model.Function{
						Name:      claudeResponse.ContentBlock.Name,
						Arguments: "",
					},
				})
			}
		}
	case "content_block_delta":
		if claudeResponse.Delta != nil {
			responseText = claudeResponse.Delta.Text
			if claudeResponse.Delta.Type == "input_json_delta" {
				tools = append(tools, model.Tool{
					Function: model.Function{
						Arguments: claudeResponse.Delta.PartialJson,
					},
				})
			}
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
	if len(tools) > 0 {
		choice.Delta.Content = nil // compatible with other OpenAI derivative applications, like LobeOpenAICompatibleFactory ...
		choice.Delta.ToolCalls = tools
	}
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
	contentArray := claudeResponse.Content.ToContentArray()
	if len(contentArray) > 0 {
		responseText = contentArray[0].Text
	}
	tools := make([]model.Tool, 0)
	for _, v := range contentArray {
		if v.Type == "tool_use" {
			args, _ := json.Marshal(v.Input)
			tools = append(tools, model.Tool{
				Id:   v.Id,
				Type: "function", // compatible with other OpenAI derivative applications
				Function: model.Function{
					Name:      v.Name,
					Arguments: string(args),
				},
			})
		}
	}
	choice := openai.TextResponseChoice{
		Index: 0,
		Message: model.Message{
			Role:      "assistant",
			Content:   responseText,
			Name:      nil,
			ToolCalls: tools,
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

	common.SetEventStreamHeaders(c)

	var usage model.Usage
	var modelName string
	var id string
	var lastToolCallChoice openai.ChatCompletionsStreamResponseChoice

	for scanner.Scan() {
		data := scanner.Text()
		if len(data) < 6 || !strings.HasPrefix(data, "data:") {
			continue
		}
		data = strings.TrimPrefix(data, "data:")
		data = strings.TrimSpace(data)

		var claudeResponse StreamResponse
		err := json.Unmarshal([]byte(data), &claudeResponse)
		if err != nil {
			logger.SysError("error unmarshalling stream response: " + err.Error())
			continue
		}

		response, meta := StreamResponseClaude2OpenAI(&claudeResponse)
		if meta != nil {
			usage.PromptTokens += meta.Usage.InputTokens
			usage.CompletionTokens += meta.Usage.OutputTokens
			if len(meta.Id) > 0 { // only message_start has an id, otherwise it's a finish_reason event.
				modelName = meta.Model
				id = fmt.Sprintf("chatcmpl-%s", meta.Id)
				continue
			} else { // finish_reason case
				if len(lastToolCallChoice.Delta.ToolCalls) > 0 {
					lastArgs := &lastToolCallChoice.Delta.ToolCalls[len(lastToolCallChoice.Delta.ToolCalls)-1].Function
					if len(lastArgs.Arguments.(string)) == 0 { // compatible with OpenAI sending an empty object `{}` when no arguments.
						lastArgs.Arguments = "{}"
						response.Choices[len(response.Choices)-1].Delta.Content = nil
						response.Choices[len(response.Choices)-1].Delta.ToolCalls = lastToolCallChoice.Delta.ToolCalls
					}
				}
			}
		}
		if response == nil {
			continue
		}

		response.Id = id
		response.Model = modelName
		response.Created = createdTime

		for _, choice := range response.Choices {
			if len(choice.Delta.ToolCalls) > 0 {
				lastToolCallChoice = choice
			}
		}
		err = render.ObjectData(c, response)
		if err != nil {
			logger.SysError(err.Error())
		}
	}

	if err := scanner.Err(); err != nil {
		logger.SysError("error reading stream: " + err.Error())
	}

	render.Done(c)

	err := resp.Body.Close()
	if err != nil {
		return openai.ErrorWrapper(err, "close_response_body_failed", http.StatusInternalServerError), nil
	}
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

// DirectHandler handles native Anthropic API responses without conversion to OpenAI format
func DirectHandler(c *gin.Context, resp *http.Response, promptTokens int, modelName string) (*model.ErrorWithStatusCode, *model.Usage) {
	ctx := c.Request.Context()
	logger.Debugf(ctx, "DirectHandler - Response status: %d", resp.StatusCode)

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Errorf(ctx, "Failed to read response body: %s", err.Error())
		return openai.ErrorWrapper(err, "read_response_body_failed", http.StatusInternalServerError), nil
	}
	err = resp.Body.Close()
	if err != nil {
		logger.Errorf(ctx, "Failed to close response body: %s", err.Error())
		return openai.ErrorWrapper(err, "close_response_body_failed", http.StatusInternalServerError), nil
	}

	logger.Debugf(ctx, "Raw response body: %s", string(responseBody))

	var claudeResponse Response
	err = json.Unmarshal(responseBody, &claudeResponse)
	if err != nil {
		logger.Errorf(ctx, "Failed to unmarshal response: %s", err.Error())
		// If we can't parse as Anthropic response, maybe it's an error response
		// Let's try to write it directly and see what happens
		c.Writer.Header().Set("Content-Type", "application/json")
		c.Writer.WriteHeader(resp.StatusCode)
		_, writeErr := c.Writer.Write(responseBody)
		if writeErr != nil {
			logger.Errorf(ctx, "Failed to write raw response: %s", writeErr.Error())
			return openai.ErrorWrapper(writeErr, "write_response_failed", http.StatusInternalServerError), nil
		}
		// Return a minimal usage for tracking
		usage := &model.Usage{PromptTokens: promptTokens, CompletionTokens: 0, TotalTokens: promptTokens}
		return nil, usage
	}

	logger.Debugf(ctx, "Parsed response - ID: %s, Model: %s, Usage: %+v",
		claudeResponse.Id, claudeResponse.Model, claudeResponse.Usage)

	if claudeResponse.Error.Type != "" {
		logger.Errorf(ctx, "Anthropic API error: %s - %s", claudeResponse.Error.Type, claudeResponse.Error.Message)
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

	// For direct mode, return the response as-is without conversion
	usage := model.Usage{
		PromptTokens:     claudeResponse.Usage.InputTokens,
		CompletionTokens: claudeResponse.Usage.OutputTokens,
		TotalTokens:      claudeResponse.Usage.InputTokens + claudeResponse.Usage.OutputTokens,
	}

	logger.Debugf(ctx, "Usage calculated: %+v", usage)

	// Write the original Anthropic response directly
	c.Writer.Header().Set("Content-Type", "application/json")
	c.Writer.WriteHeader(resp.StatusCode)
	_, err = c.Writer.Write(responseBody)
	if err != nil {
		logger.Errorf(ctx, "Failed to write response: %s", err.Error())
		return openai.ErrorWrapper(err, "write_response_failed", http.StatusInternalServerError), nil
	}

	logger.Debugf(ctx, "Response written successfully")
	return nil, &usage
}

// DirectStreamHandler handles native Anthropic API streaming responses without conversion
func DirectStreamHandler(c *gin.Context, resp *http.Response) (*model.ErrorWithStatusCode, *model.Usage) {
	defer resp.Body.Close()

	// Set headers for streaming
	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
	c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	c.Writer.WriteHeader(resp.StatusCode)

	// Stream the response directly without conversion
	var usage model.Usage
	scanner := bufio.NewScanner(resp.Body)

	// 优化: 一次性设置scanner缓冲区大小，减少扩容开销
	buf := make([]byte, StreamBufferInitialSize)
	scanner.Buffer(buf, StreamBufferMaxSize)

	// 优化: 使用write buffer累积数据，减少flush次数
	var writeBuffer strings.Builder

	for scanner.Scan() {
		data := scanner.Text()
		if len(data) < MinDataPrefixLength || !strings.HasPrefix(data, "data:") {
			continue
		}

		// 优化: 只在特定事件类型时解析JSON，减少80-90%的解析开销
		needFlush := false
		dataLen := len(data)

		// 快速判断是否为message_start事件（包含input_tokens）
		if dataLen > MinEventDataLength {
			// 安全地获取检查范围，避免slice越界
			checkRange := dataLen
			if dataLen > EventTypeCheckRange {
				checkRange = EventTypeCheckRange
			}
			if strings.Contains(data[:checkRange], "message_start") {
				jsonData := strings.TrimPrefix(data, "data:")
				jsonData = strings.TrimSpace(jsonData)

				var eventData map[string]interface{}
				if err := json.Unmarshal([]byte(jsonData), &eventData); err == nil {
					if msg, ok := eventData["message"].(map[string]interface{}); ok {
						if usageData, ok := msg["usage"].(map[string]interface{}); ok {
							if inputTokens, ok := usageData["input_tokens"].(float64); ok {
								usage.PromptTokens = int(inputTokens)
							}
						}
					}
				}
				needFlush = true
			} else if dataLen > MinEventDataLength && strings.Contains(data, "message_delta") {
				// message_delta事件可能包含output_tokens
				jsonData := strings.TrimPrefix(data, "data:")
				jsonData = strings.TrimSpace(jsonData)

				var eventData map[string]interface{}
				if err := json.Unmarshal([]byte(jsonData), &eventData); err == nil {
					if usageData, ok := eventData["usage"].(map[string]interface{}); ok {
						if outputTokens, ok := usageData["output_tokens"].(float64); ok {
							usage.CompletionTokens += int(outputTokens)
							usage.TotalTokens = usage.PromptTokens + usage.CompletionTokens
						}
					}
				}
				needFlush = true
			} else if strings.Contains(data, "content_block_delta") {
				// content_block_delta是内容更新事件，需要及时flush以保证实时性
				needFlush = true
			}
		}

		// 累积数据到write buffer
		writeBuffer.WriteString(data)
		writeBuffer.WriteByte('\n')

		// 优化: 智能flush策略 - 只在关键事件或缓冲区满时flush
		if needFlush || writeBuffer.Len() >= StreamFlushThreshold {
			c.Writer.WriteString(writeBuffer.String())
			c.Writer.Flush()
			writeBuffer.Reset()
		}
	}

	if err := scanner.Err(); err != nil {
		return openai.ErrorWrapper(err, "stream_read_failed", http.StatusInternalServerError), nil
	}

	// 优化: 刷新剩余数据
	if writeBuffer.Len() > 0 {
		c.Writer.WriteString(writeBuffer.String())
		c.Writer.Flush()
	}

	return nil, &usage
}
