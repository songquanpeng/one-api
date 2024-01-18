package ali

import (
	"encoding/json"
	"net/http"
	"one-api/common"
	"one-api/common/requester"
	"one-api/types"
	"strings"
)

type aliStreamHandler struct {
	Usage              *types.Usage
	Request            *types.ChatCompletionRequest
	lastStreamResponse string
}

const AliEnableSearchModelSuffix = "-internet"

func (p *AliProvider) CreateChatCompletion(request *types.ChatCompletionRequest) (*types.ChatCompletionResponse, *types.OpenAIErrorWithStatusCode) {
	req, errWithCode := p.getAliChatRequest(request)
	if errWithCode != nil {
		return nil, errWithCode
	}
	defer req.Body.Close()

	aliResponse := &AliChatResponse{}
	// 发送请求
	_, errWithCode = p.Requester.SendRequest(req, aliResponse, false)
	if errWithCode != nil {
		return nil, errWithCode
	}

	return p.convertToChatOpenai(aliResponse, request)
}

func (p *AliProvider) CreateChatCompletionStream(request *types.ChatCompletionRequest) (requester.StreamReaderInterface[types.ChatCompletionStreamResponse], *types.OpenAIErrorWithStatusCode) {
	req, errWithCode := p.getAliChatRequest(request)
	if errWithCode != nil {
		return nil, errWithCode
	}
	defer req.Body.Close()

	// 发送请求
	resp, errWithCode := p.Requester.SendRequestRaw(req)
	if errWithCode != nil {
		return nil, errWithCode
	}

	chatHandler := &aliStreamHandler{
		Usage:   p.Usage,
		Request: request,
	}

	return requester.RequestStream[types.ChatCompletionStreamResponse](p.Requester, resp, chatHandler.handlerStream)
}

func (p *AliProvider) getAliChatRequest(request *types.ChatCompletionRequest) (*http.Request, *types.OpenAIErrorWithStatusCode) {
	url, errWithCode := p.GetSupportedAPIUri(common.RelayModeChatCompletions)
	if errWithCode != nil {
		return nil, errWithCode
	}
	// 获取请求地址
	fullRequestURL := p.GetFullRequestURL(url, request.Model)

	// 获取请求头
	headers := p.GetRequestHeaders()
	if request.Stream {
		headers["Accept"] = "text/event-stream"
		headers["X-DashScope-SSE"] = "enable"
	}

	aliRequest := convertFromChatOpenai(request)
	// 创建请求
	req, err := p.Requester.NewRequest(http.MethodPost, fullRequestURL, p.Requester.WithBody(aliRequest), p.Requester.WithHeader(headers))
	if err != nil {
		return nil, common.ErrorWrapper(err, "new_request_failed", http.StatusInternalServerError)
	}

	return req, nil
}

// 转换为OpenAI聊天请求体
func (p *AliProvider) convertToChatOpenai(response *AliChatResponse, request *types.ChatCompletionRequest) (openaiResponse *types.ChatCompletionResponse, errWithCode *types.OpenAIErrorWithStatusCode) {
	error := errorHandle(&response.AliError)
	if error != nil {
		errWithCode = &types.OpenAIErrorWithStatusCode{
			OpenAIError: *error,
			StatusCode:  http.StatusBadRequest,
		}
		return
	}

	openaiResponse = &types.ChatCompletionResponse{
		ID:      response.RequestId,
		Object:  "chat.completion",
		Created: common.GetTimestamp(),
		Model:   request.Model,
		Choices: response.Output.ToChatCompletionChoices(),
		Usage: &types.Usage{
			PromptTokens:     response.Usage.InputTokens,
			CompletionTokens: response.Usage.OutputTokens,
			TotalTokens:      response.Usage.InputTokens + response.Usage.OutputTokens,
		},
	}

	*p.Usage = *openaiResponse.Usage

	return
}

// 阿里云聊天请求体
func convertFromChatOpenai(request *types.ChatCompletionRequest) *AliChatRequest {
	messages := make([]AliMessage, 0, len(request.Messages))
	for i := 0; i < len(request.Messages); i++ {
		message := request.Messages[i]
		if request.Model != "qwen-vl-plus" {
			messages = append(messages, AliMessage{
				Content: message.StringContent(),
				Role:    strings.ToLower(message.Role),
			})
		} else {
			openaiContent := message.ParseContent()
			var parts []AliMessagePart
			for _, part := range openaiContent {
				if part.Type == types.ContentTypeText {
					parts = append(parts, AliMessagePart{
						Text: part.Text,
					})
				} else if part.Type == types.ContentTypeImageURL {
					parts = append(parts, AliMessagePart{
						Image: part.ImageURL.URL,
					})
				}
			}
			messages = append(messages, AliMessage{
				Content: parts,
				Role:    strings.ToLower(message.Role),
			})
		}

	}

	enableSearch := false
	aliModel := request.Model
	if strings.HasSuffix(aliModel, AliEnableSearchModelSuffix) {
		enableSearch = true
		aliModel = strings.TrimSuffix(aliModel, AliEnableSearchModelSuffix)
	}

	return &AliChatRequest{
		Model: aliModel,
		Input: AliInput{
			Messages: messages,
		},
		Parameters: AliParameters{
			ResultFormat:      "message",
			EnableSearch:      enableSearch,
			IncrementalOutput: request.Stream,
		},
	}
}

// 转换为OpenAI聊天流式请求体
func (h *aliStreamHandler) handlerStream(rawLine *[]byte, isFinished *bool, response *[]types.ChatCompletionStreamResponse) error {
	// 如果rawLine 前缀不为data:，则直接返回
	if !strings.HasPrefix(string(*rawLine), "data:") {
		*rawLine = nil
		return nil
	}

	// 去除前缀
	*rawLine = (*rawLine)[5:]

	var aliResponse AliChatResponse
	err := json.Unmarshal(*rawLine, &aliResponse)
	if err != nil {
		return common.ErrorToOpenAIError(err)
	}

	error := errorHandle(&aliResponse.AliError)
	if error != nil {
		return error
	}

	return h.convertToOpenaiStream(&aliResponse, response)

}

func (h *aliStreamHandler) convertToOpenaiStream(aliResponse *AliChatResponse, response *[]types.ChatCompletionStreamResponse) error {
	content := aliResponse.Output.Choices[0].Message.StringContent()

	var choice types.ChatCompletionStreamChoice
	choice.Index = aliResponse.Output.Choices[0].Index
	choice.Delta.Content = strings.TrimPrefix(content, h.lastStreamResponse)
	if aliResponse.Output.Choices[0].FinishReason != "" {
		if aliResponse.Output.Choices[0].FinishReason != "null" {
			finishReason := aliResponse.Output.Choices[0].FinishReason
			choice.FinishReason = &finishReason
		}
	}

	if aliResponse.Output.FinishReason != "" {
		if aliResponse.Output.FinishReason != "null" {
			finishReason := aliResponse.Output.FinishReason
			choice.FinishReason = &finishReason
		}
	}

	h.lastStreamResponse = content
	streamResponse := types.ChatCompletionStreamResponse{
		ID:      aliResponse.RequestId,
		Object:  "chat.completion.chunk",
		Created: common.GetTimestamp(),
		Model:   h.Request.Model,
		Choices: []types.ChatCompletionStreamChoice{choice},
	}

	if aliResponse.Usage.OutputTokens != 0 {
		h.Usage.PromptTokens = aliResponse.Usage.InputTokens
		h.Usage.CompletionTokens = aliResponse.Usage.OutputTokens
		h.Usage.TotalTokens = aliResponse.Usage.InputTokens + aliResponse.Usage.OutputTokens
	}

	*response = append(*response, streamResponse)

	return nil
}
