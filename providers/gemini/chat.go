package gemini

import (
	"encoding/json"
	"fmt"
	"net/http"
	"one-api/common"
	"one-api/common/image"
	"one-api/common/requester"
	"one-api/types"
	"strings"
)

const (
	GeminiVisionMaxImageNum = 16
)

type geminiStreamHandler struct {
	Usage   *types.Usage
	Request *types.ChatCompletionRequest
}

func (p *GeminiProvider) CreateChatCompletion(request *types.ChatCompletionRequest) (*types.ChatCompletionResponse, *types.OpenAIErrorWithStatusCode) {
	req, errWithCode := p.getChatRequest(request)
	if errWithCode != nil {
		return nil, errWithCode
	}
	defer req.Body.Close()

	geminiChatResponse := &GeminiChatResponse{}
	// 发送请求
	_, errWithCode = p.Requester.SendRequest(req, geminiChatResponse, false)
	if errWithCode != nil {
		return nil, errWithCode
	}

	return p.convertToChatOpenai(geminiChatResponse, request)
}

func (p *GeminiProvider) CreateChatCompletionStream(request *types.ChatCompletionRequest) (requester.StreamReaderInterface[types.ChatCompletionStreamResponse], *types.OpenAIErrorWithStatusCode) {
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

	chatHandler := &geminiStreamHandler{
		Usage:   p.Usage,
		Request: request,
	}

	return requester.RequestStream[types.ChatCompletionStreamResponse](p.Requester, resp, chatHandler.handlerStream)
}

func (p *GeminiProvider) getChatRequest(request *types.ChatCompletionRequest) (*http.Request, *types.OpenAIErrorWithStatusCode) {
	url := "generateContent"
	if request.Stream {
		url = "streamGenerateContent?alt=sse"
	}
	// 获取请求地址
	fullRequestURL := p.GetFullRequestURL(url, request.Model)

	// 获取请求头
	headers := p.GetRequestHeaders()
	if request.Stream {
		headers["Accept"] = "text/event-stream"
	}

	geminiRequest, errWithCode := convertFromChatOpenai(request)
	if errWithCode != nil {
		return nil, errWithCode
	}

	// 创建请求
	req, err := p.Requester.NewRequest(http.MethodPost, fullRequestURL, p.Requester.WithBody(geminiRequest), p.Requester.WithHeader(headers))
	if err != nil {
		return nil, common.ErrorWrapper(err, "new_request_failed", http.StatusInternalServerError)
	}

	return req, nil
}

func convertFromChatOpenai(request *types.ChatCompletionRequest) (*GeminiChatRequest, *types.OpenAIErrorWithStatusCode) {
	geminiRequest := GeminiChatRequest{
		Contents: make([]GeminiChatContent, 0, len(request.Messages)),
		SafetySettings: []GeminiChatSafetySettings{
			{
				Category:  "HARM_CATEGORY_HARASSMENT",
				Threshold: "BLOCK_NONE",
			},
			{
				Category:  "HARM_CATEGORY_HATE_SPEECH",
				Threshold: "BLOCK_NONE",
			},
			{
				Category:  "HARM_CATEGORY_SEXUALLY_EXPLICIT",
				Threshold: "BLOCK_NONE",
			},
			{
				Category:  "HARM_CATEGORY_DANGEROUS_CONTENT",
				Threshold: "BLOCK_NONE",
			},
		},
		GenerationConfig: GeminiChatGenerationConfig{
			Temperature:     request.Temperature,
			TopP:            request.TopP,
			MaxOutputTokens: request.MaxTokens,
		},
	}
	if request.Functions != nil {
		geminiRequest.Tools = []GeminiChatTools{
			{
				FunctionDeclarations: request.Functions,
			},
		}
	}
	shouldAddDummyModelMessage := false
	for _, message := range request.Messages {
		content := GeminiChatContent{
			Role: message.Role,
			Parts: []GeminiPart{
				{
					Text: message.StringContent(),
				},
			},
		}

		openaiContent := message.ParseContent()
		var parts []GeminiPart
		imageNum := 0
		for _, part := range openaiContent {
			if part.Type == types.ContentTypeText {
				parts = append(parts, GeminiPart{
					Text: part.Text,
				})
			} else if part.Type == types.ContentTypeImageURL {
				imageNum += 1
				if imageNum > GeminiVisionMaxImageNum {
					continue
				}
				mimeType, data, err := image.GetImageFromUrl(part.ImageURL.URL)
				if err != nil {
					return nil, common.ErrorWrapper(err, "image_url_invalid", http.StatusBadRequest)
				}
				parts = append(parts, GeminiPart{
					InlineData: &GeminiInlineData{
						MimeType: mimeType,
						Data:     data,
					},
				})
			}
		}
		content.Parts = parts

		// there's no assistant role in gemini and API shall vomit if Role is not user or model
		if content.Role == "assistant" {
			content.Role = "model"
		}
		// Converting system prompt to prompt from user for the same reason
		if content.Role == "system" {
			content.Role = "user"
			shouldAddDummyModelMessage = true
		}
		geminiRequest.Contents = append(geminiRequest.Contents, content)

		// If a system message is the last message, we need to add a dummy model message to make gemini happy
		if shouldAddDummyModelMessage {
			geminiRequest.Contents = append(geminiRequest.Contents, GeminiChatContent{
				Role: "model",
				Parts: []GeminiPart{
					{
						Text: "Okay",
					},
				},
			})
			shouldAddDummyModelMessage = false
		}
	}

	return &geminiRequest, nil
}

func (p *GeminiProvider) convertToChatOpenai(response *GeminiChatResponse, request *types.ChatCompletionRequest) (openaiResponse *types.ChatCompletionResponse, errWithCode *types.OpenAIErrorWithStatusCode) {
	error := errorHandle(&response.GeminiErrorResponse)
	if error != nil {
		errWithCode = &types.OpenAIErrorWithStatusCode{
			OpenAIError: *error,
			StatusCode:  http.StatusBadRequest,
		}
		return
	}

	openaiResponse = &types.ChatCompletionResponse{
		ID:      fmt.Sprintf("chatcmpl-%s", common.GetUUID()),
		Object:  "chat.completion",
		Created: common.GetTimestamp(),
		Model:   request.Model,
		Choices: make([]types.ChatCompletionChoice, 0, len(response.Candidates)),
	}
	for i, candidate := range response.Candidates {
		choice := types.ChatCompletionChoice{
			Index: i,
			Message: types.ChatCompletionMessage{
				Role:    "assistant",
				Content: "",
			},
			FinishReason: types.FinishReasonStop,
		}
		if len(candidate.Content.Parts) > 0 {
			choice.Message.Content = candidate.Content.Parts[0].Text
		}
		openaiResponse.Choices = append(openaiResponse.Choices, choice)
	}

	completionTokens := common.CountTokenText(response.GetResponseText(), response.Model)

	p.Usage.CompletionTokens = completionTokens
	p.Usage.TotalTokens = p.Usage.PromptTokens + completionTokens
	openaiResponse.Usage = p.Usage

	return
}

// 转换为OpenAI聊天流式请求体
func (h *geminiStreamHandler) handlerStream(rawLine *[]byte, isFinished *bool, response *[]types.ChatCompletionStreamResponse) error {
	// 如果rawLine 前缀不为data:，则直接返回
	if !strings.HasPrefix(string(*rawLine), "data: ") {
		*rawLine = nil
		return nil
	}

	// 去除前缀
	*rawLine = (*rawLine)[6:]

	var geminiResponse GeminiChatResponse
	err := json.Unmarshal(*rawLine, &geminiResponse)
	if err != nil {
		return common.ErrorToOpenAIError(err)
	}

	error := errorHandle(&geminiResponse.GeminiErrorResponse)
	if error != nil {
		return error
	}

	return h.convertToOpenaiStream(&geminiResponse, response)

}

func (h *geminiStreamHandler) convertToOpenaiStream(geminiResponse *GeminiChatResponse, response *[]types.ChatCompletionStreamResponse) error {
	choices := make([]types.ChatCompletionStreamChoice, 0, len(geminiResponse.Candidates))

	for i, candidate := range geminiResponse.Candidates {
		choice := types.ChatCompletionStreamChoice{
			Index: i,
			Delta: types.ChatCompletionStreamChoiceDelta{
				Content: candidate.Content.Parts[0].Text,
			},
			FinishReason: types.FinishReasonStop,
		}
		choices = append(choices, choice)
	}

	streamResponse := types.ChatCompletionStreamResponse{
		ID:      fmt.Sprintf("chatcmpl-%s", common.GetUUID()),
		Object:  "chat.completion.chunk",
		Created: common.GetTimestamp(),
		Model:   h.Request.Model,
		Choices: choices,
	}

	*response = append(*response, streamResponse)

	h.Usage.CompletionTokens += common.CountTokenText(geminiResponse.GetResponseText(), h.Request.Model)
	h.Usage.TotalTokens = h.Usage.PromptTokens + h.Usage.CompletionTokens

	return nil
}
