package gemini

import (
	"encoding/json"
	"fmt"
	"net/http"
	"one-api/common"
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

func (p *GeminiProvider) CreateChatCompletionStream(request *types.ChatCompletionRequest) (requester.StreamReaderInterface[string], *types.OpenAIErrorWithStatusCode) {
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

	return requester.RequestStream[string](p.Requester, resp, chatHandler.handlerStream)
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
	if request.Tools != nil {
		var geminiChatTools GeminiChatTools
		for _, tool := range request.Tools {
			geminiChatTools.FunctionDeclarations = append(geminiChatTools.FunctionDeclarations, tool.Function)
		}
		geminiRequest.Tools = append(geminiRequest.Tools, geminiChatTools)
	}

	geminiContent, err := OpenAIToGeminiChatContent(request.Messages)
	if err != nil {
		return nil, err
	}

	geminiRequest.Contents = geminiContent

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
				Role: "assistant",
				// Content: "",
			},
			FinishReason: types.FinishReasonStop,
		}
		if len(candidate.Content.Parts) == 0 {
			choice.Message.Content = ""
			openaiResponse.Choices = append(openaiResponse.Choices, choice)
			continue
			// choice.Message.Content = candidate.Content.Parts[0].Text
		}
		// 开始判断
		geminiParts := candidate.Content.Parts[0]

		if geminiParts.FunctionCall != nil {
			choice.Message.ToolCalls = geminiParts.FunctionCall.ToOpenAITool()
		} else {
			choice.Message.Content = geminiParts.Text
		}
		openaiResponse.Choices = append(openaiResponse.Choices, choice)
	}

	*p.Usage = convertOpenAIUsage(request.Model, response.UsageMetadata)
	openaiResponse.Usage = p.Usage

	return
}

// 转换为OpenAI聊天流式请求体
func (h *geminiStreamHandler) handlerStream(rawLine *[]byte, dataChan chan string, errChan chan error) {
	// 如果rawLine 前缀不为data:，则直接返回
	if !strings.HasPrefix(string(*rawLine), "data: ") {
		*rawLine = nil
		return
	}

	// 去除前缀
	*rawLine = (*rawLine)[6:]

	var geminiResponse GeminiChatResponse
	err := json.Unmarshal(*rawLine, &geminiResponse)
	if err != nil {
		errChan <- common.ErrorToOpenAIError(err)
		return
	}

	error := errorHandle(&geminiResponse.GeminiErrorResponse)
	if error != nil {
		errChan <- error
		return
	}

	h.convertToOpenaiStream(&geminiResponse, dataChan)

}

func (h *geminiStreamHandler) convertToOpenaiStream(geminiResponse *GeminiChatResponse, dataChan chan string) {
	streamResponse := types.ChatCompletionStreamResponse{
		ID:      fmt.Sprintf("chatcmpl-%s", common.GetUUID()),
		Object:  "chat.completion.chunk",
		Created: common.GetTimestamp(),
		Model:   h.Request.Model,
		// Choices: choices,
	}

	choices := make([]types.ChatCompletionStreamChoice, 0, len(geminiResponse.Candidates))

	for i, candidate := range geminiResponse.Candidates {
		parts := candidate.Content.Parts[0]

		choice := types.ChatCompletionStreamChoice{
			Index: i,
			Delta: types.ChatCompletionStreamChoiceDelta{
				Role: types.ChatMessageRoleAssistant,
			},
			FinishReason: types.FinishReasonStop,
		}

		if parts.FunctionCall != nil {
			if parts.FunctionCall.Args == nil {
				parts.FunctionCall.Args = map[string]interface{}{}
			}
			args, _ := json.Marshal(parts.FunctionCall.Args)

			choice.Delta.ToolCalls = []*types.ChatCompletionToolCalls{
				{
					Id:    "call_" + common.GetRandomString(24),
					Type:  types.ChatMessageRoleFunction,
					Index: 0,
					Function: &types.ChatCompletionToolCallsFunction{
						Name:      parts.FunctionCall.Name,
						Arguments: string(args),
					},
				},
			}
		} else {
			choice.Delta.Content = parts.Text
		}

		choices = append(choices, choice)
	}

	if len(choices) > 0 && choices[0].Delta.ToolCalls != nil {
		choices := choices[0].ConvertOpenaiStream()
		for _, choice := range choices {
			chatCompletionCopy := streamResponse
			chatCompletionCopy.Choices = []types.ChatCompletionStreamChoice{choice}
			responseBody, _ := json.Marshal(chatCompletionCopy)
			dataChan <- string(responseBody)
		}
	} else {
		streamResponse.Choices = choices
		responseBody, _ := json.Marshal(streamResponse)
		dataChan <- string(responseBody)
	}

	if geminiResponse.UsageMetadata != nil {
		*h.Usage = convertOpenAIUsage(h.Request.Model, geminiResponse.UsageMetadata)

	}
}

const tokenThreshold = 1000000

var modelAdjustRatios = map[string]int{
	"gemini-1.5-pro":   2,
	"gemini-1.5-flash": 2,
}

func adjustTokenCounts(modelName string, usage *GeminiUsageMetadata) {
	if usage.PromptTokenCount <= tokenThreshold && usage.CandidatesTokenCount <= tokenThreshold {
		return
	}

	currentRatio := 1
	for model, r := range modelAdjustRatios {
		if strings.HasPrefix(modelName, model) {
			currentRatio = r
			break
		}
	}

	if currentRatio == 1 {
		return
	}

	adjustTokenCount := func(count int) int {
		if count > tokenThreshold {
			return tokenThreshold + (count-tokenThreshold)*currentRatio
		}
		return count
	}

	if usage.PromptTokenCount > tokenThreshold {
		usage.PromptTokenCount = adjustTokenCount(usage.PromptTokenCount)
	}

	if usage.CandidatesTokenCount > tokenThreshold {
		usage.CandidatesTokenCount = adjustTokenCount(usage.CandidatesTokenCount)
	}

	usage.TotalTokenCount = usage.PromptTokenCount + usage.CandidatesTokenCount
}

func convertOpenAIUsage(modelName string, geminiUsage *GeminiUsageMetadata) types.Usage {
	adjustTokenCounts(modelName, geminiUsage)

	return types.Usage{
		PromptTokens:     geminiUsage.PromptTokenCount,
		CompletionTokens: geminiUsage.CandidatesTokenCount,
		TotalTokens:      geminiUsage.TotalTokenCount,
	}
}
