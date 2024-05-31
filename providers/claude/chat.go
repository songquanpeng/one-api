package claude

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"one-api/common"
	"one-api/common/config"
	"one-api/common/image"
	"one-api/common/requester"
	"one-api/common/utils"
	"one-api/providers/base"
	"one-api/types"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws/protocol/eventstream"
)

type ClaudeStreamHandler struct {
	Usage   *types.Usage
	Request *types.ChatCompletionRequest
	Prefix  string
}

func (p *ClaudeProvider) CreateChatCompletion(request *types.ChatCompletionRequest) (*types.ChatCompletionResponse, *types.OpenAIErrorWithStatusCode) {
	req, errWithCode := p.getChatRequest(request)
	if errWithCode != nil {
		return nil, errWithCode
	}
	defer req.Body.Close()

	claudeResponse := &ClaudeResponse{}
	// 发送请求
	_, errWithCode = p.Requester.SendRequest(req, claudeResponse, false)
	if errWithCode != nil {
		return nil, errWithCode
	}

	return ConvertToChatOpenai(p, claudeResponse, request)
}

func (p *ClaudeProvider) CreateChatCompletionStream(request *types.ChatCompletionRequest) (requester.StreamReaderInterface[string], *types.OpenAIErrorWithStatusCode) {
	req, errWithCode := p.getChatRequest(request)
	if errWithCode != nil {
		return nil, errWithCode
	}
	defer req.Body.Close()

	// 发送请求
	resp, errWithCode := p.Requester.SendRequestRaw(req)
	if errWithCode != nil {
		return nil, errWithCode
	}

	chatHandler := &ClaudeStreamHandler{
		Usage:   p.Usage,
		Request: request,
		Prefix:  `data: {"type"`,
	}

	eventstream.NewDecoder()

	return requester.RequestStream(p.Requester, resp, chatHandler.HandlerStream)
}

func (p *ClaudeProvider) getChatRequest(request *types.ChatCompletionRequest) (*http.Request, *types.OpenAIErrorWithStatusCode) {
	url, errWithCode := p.GetSupportedAPIUri(config.RelayModeChatCompletions)
	if errWithCode != nil {
		return nil, errWithCode
	}

	// 获取请求地址
	fullRequestURL := p.GetFullRequestURL(url)
	if fullRequestURL == "" {
		return nil, common.ErrorWrapper(nil, "invalid_claude_config", http.StatusInternalServerError)
	}

	headers := p.GetRequestHeaders()
	if request.Stream {
		headers["Accept"] = "text/event-stream"
	}

	claudeRequest, errWithCode := ConvertFromChatOpenai(request)
	if errWithCode != nil {
		return nil, errWithCode
	}

	// 创建请求
	req, err := p.Requester.NewRequest(http.MethodPost, fullRequestURL, p.Requester.WithBody(claudeRequest), p.Requester.WithHeader(headers))
	if err != nil {
		return nil, common.ErrorWrapper(err, "new_request_failed", http.StatusInternalServerError)
	}

	return req, nil
}

func ConvertFromChatOpenai(request *types.ChatCompletionRequest) (*ClaudeRequest, *types.OpenAIErrorWithStatusCode) {
	request.ClearEmptyMessages()
	claudeRequest := ClaudeRequest{
		Model:         request.Model,
		Messages:      []Message{},
		System:        "",
		MaxTokens:     request.MaxTokens,
		StopSequences: nil,
		Temperature:   request.Temperature,
		TopP:          request.TopP,
		Stream:        request.Stream,
	}
	if claudeRequest.MaxTokens == 0 {
		claudeRequest.MaxTokens = 4096
	}

	for _, message := range request.Messages {
		if message.Role == "system" {
			claudeRequest.System = message.Content.(string)
			continue
		}
		content := Message{
			Role:    convertRole(message.Role),
			Content: []MessageContent{},
		}

		openaiContent := message.ParseContent()
		for _, part := range openaiContent {
			if part.Type == types.ContentTypeText {
				content.Content = append(content.Content, MessageContent{
					Type: "text",
					Text: part.Text,
				})
				continue
			}

			if part.Type == types.ContentTypeImageURL {
				mimeType, data, err := image.GetImageFromUrl(part.ImageURL.URL)
				if err != nil {
					return nil, common.ErrorWrapper(err, "image_url_invalid", http.StatusBadRequest)
				}
				content.Content = append(content.Content, MessageContent{
					Type: "image",
					Source: &ContentSource{
						Type:      "base64",
						MediaType: mimeType,
						Data:      data,
					},
				})
			}
		}
		claudeRequest.Messages = append(claudeRequest.Messages, content)
	}

	return &claudeRequest, nil
}

func ConvertToChatOpenai(provider base.ProviderInterface, response *ClaudeResponse, request *types.ChatCompletionRequest) (openaiResponse *types.ChatCompletionResponse, errWithCode *types.OpenAIErrorWithStatusCode) {
	aiError := errorHandle(&response.Error)
	if aiError != nil {
		errWithCode = &types.OpenAIErrorWithStatusCode{
			OpenAIError: *aiError,
			StatusCode:  http.StatusBadRequest,
		}
		return
	}

	choice := types.ChatCompletionChoice{
		Index: 0,
		Message: types.ChatCompletionMessage{
			Role:    response.Role,
			Content: strings.TrimPrefix(response.Content[0].Text, " "),
			Name:    nil,
		},
		FinishReason: stopReasonClaude2OpenAI(response.StopReason),
	}
	openaiResponse = &types.ChatCompletionResponse{
		ID:      response.Id,
		Object:  "chat.completion",
		Created: utils.GetTimestamp(),
		Choices: []types.ChatCompletionChoice{choice},
		Model:   request.Model,
		Usage: &types.Usage{
			CompletionTokens: 0,
			PromptTokens:     0,
			TotalTokens:      0,
		},
	}

	completionTokens := response.Usage.OutputTokens

	promptTokens := response.Usage.InputTokens

	openaiResponse.Usage.PromptTokens = promptTokens
	openaiResponse.Usage.CompletionTokens = completionTokens
	openaiResponse.Usage.TotalTokens = promptTokens + completionTokens

	usage := provider.GetUsage()
	*usage = *openaiResponse.Usage

	return openaiResponse, nil
}

// 转换为OpenAI聊天流式请求体
func (h *ClaudeStreamHandler) HandlerStream(rawLine *[]byte, dataChan chan string, errChan chan error) {
	// 如果rawLine 前缀不为data:，则直接返回
	if !strings.HasPrefix(string(*rawLine), h.Prefix) {
		*rawLine = nil
		return
	}

	if strings.HasPrefix(string(*rawLine), "data: ") {
		// 去除前缀
		*rawLine = (*rawLine)[6:]
	}

	var claudeResponse ClaudeStreamResponse
	err := json.Unmarshal(*rawLine, &claudeResponse)
	if err != nil {
		errChan <- common.ErrorToOpenAIError(err)
		return
	}

	aiError := errorHandle(&claudeResponse.Error)
	if aiError != nil {
		errChan <- aiError
		return
	}

	if claudeResponse.Type == "message_stop" {
		errChan <- io.EOF
		*rawLine = requester.StreamClosed
		return
	}

	switch claudeResponse.Type {
	case "message_start":
		h.convertToOpenaiStream(&claudeResponse, dataChan)
		h.Usage.PromptTokens = claudeResponse.Message.Usage.InputTokens

	case "message_delta":
		h.convertToOpenaiStream(&claudeResponse, dataChan)
		h.Usage.CompletionTokens = claudeResponse.Usage.OutputTokens
		h.Usage.TotalTokens = h.Usage.PromptTokens + h.Usage.CompletionTokens

	case "content_block_delta":
		h.convertToOpenaiStream(&claudeResponse, dataChan)
		h.Usage.CompletionTokens += common.CountTokenText(claudeResponse.Delta.Text, h.Request.Model)
		h.Usage.TotalTokens = h.Usage.PromptTokens + h.Usage.CompletionTokens

	default:
		return
	}
}

func (h *ClaudeStreamHandler) convertToOpenaiStream(claudeResponse *ClaudeStreamResponse, dataChan chan string) {
	choice := types.ChatCompletionStreamChoice{
		Index: claudeResponse.Index,
	}

	if claudeResponse.Message.Role != "" {
		choice.Delta.Role = claudeResponse.Message.Role
	}

	if claudeResponse.Delta.Text != "" {
		choice.Delta.Content = claudeResponse.Delta.Text
	}

	finishReason := stopReasonClaude2OpenAI(claudeResponse.Delta.StopReason)
	if finishReason != "" {
		choice.FinishReason = &finishReason
	}
	chatCompletion := types.ChatCompletionStreamResponse{
		ID:      fmt.Sprintf("chatcmpl-%s", utils.GetUUID()),
		Object:  "chat.completion.chunk",
		Created: utils.GetTimestamp(),
		Model:   h.Request.Model,
		Choices: []types.ChatCompletionStreamChoice{choice},
	}

	responseBody, _ := json.Marshal(chatCompletion)
	dataChan <- string(responseBody)
}
