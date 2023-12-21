package gemini

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"one-api/common"
	"one-api/providers/base"
	"one-api/types"
	"strings"
)

func (response *GeminiChatResponse) ResponseHandler(resp *http.Response) (OpenAIResponse any, errWithCode *types.OpenAIErrorWithStatusCode) {
	if len(response.Candidates) == 0 {
		return nil, &types.OpenAIErrorWithStatusCode{
			OpenAIError: types.OpenAIError{
				Message: "No candidates returned",
				Type:    "server_error",
				Param:   "",
				Code:    500,
			},
			StatusCode: resp.StatusCode,
		}
	}

	fullTextResponse := &types.ChatCompletionResponse{
		ID:      fmt.Sprintf("chatcmpl-%s", common.GetUUID()),
		Object:  "chat.completion",
		Created: common.GetTimestamp(),
		Choices: make([]types.ChatCompletionChoice, 0, len(response.Candidates)),
	}
	for i, candidate := range response.Candidates {
		choice := types.ChatCompletionChoice{
			Index: i,
			Message: types.ChatCompletionMessage{
				Role:    "assistant",
				Content: "",
			},
			FinishReason: base.StopFinishReason,
		}
		if len(candidate.Content.Parts) > 0 {
			choice.Message.Content = candidate.Content.Parts[0].Text
		}
		fullTextResponse.Choices = append(fullTextResponse.Choices, choice)
	}

	completionTokens := common.CountTokenText(response.GetResponseText(), "gemini-pro")
	response.Usage.CompletionTokens = completionTokens
	response.Usage.TotalTokens = response.Usage.PromptTokens + completionTokens

	return fullTextResponse, nil
}

// Setting safety to the lowest possible values since Gemini is already powerless enough
func (p *GeminiProvider) getChatRequestBody(request *types.ChatCompletionRequest) (requestBody *GeminiChatRequest) {
	geminiRequest := GeminiChatRequest{
		Contents: make([]GeminiChatContent, 0, len(request.Messages)),
		//SafetySettings: []GeminiChatSafetySettings{
		//	{
		//		Category:  "HARM_CATEGORY_HARASSMENT",
		//		Threshold: "BLOCK_ONLY_HIGH",
		//	},
		//	{
		//		Category:  "HARM_CATEGORY_HATE_SPEECH",
		//		Threshold: "BLOCK_ONLY_HIGH",
		//	},
		//	{
		//		Category:  "HARM_CATEGORY_SEXUALLY_EXPLICIT",
		//		Threshold: "BLOCK_ONLY_HIGH",
		//	},
		//	{
		//		Category:  "HARM_CATEGORY_DANGEROUS_CONTENT",
		//		Threshold: "BLOCK_ONLY_HIGH",
		//	},
		//},
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

	return &geminiRequest
}

func (p *GeminiProvider) ChatAction(request *types.ChatCompletionRequest, isModelMapped bool, promptTokens int) (usage *types.Usage, errWithCode *types.OpenAIErrorWithStatusCode) {
	requestBody := p.getChatRequestBody(request)
	fullRequestURL := p.GetFullRequestURL("generateContent", request.Model)
	headers := p.GetRequestHeaders()
	if request.Stream {
		headers["Accept"] = "text/event-stream"
	}

	client := common.NewClient()
	req, err := client.NewRequest(p.Context.Request.Method, fullRequestURL, common.WithBody(requestBody), common.WithHeader(headers))
	if err != nil {
		return nil, common.ErrorWrapper(err, "new_request_failed", http.StatusInternalServerError)
	}

	if request.Stream {
		var responseText string
		errWithCode, responseText = p.sendStreamRequest(req)
		if errWithCode != nil {
			return
		}

		usage.PromptTokens = promptTokens
		usage.CompletionTokens = common.CountTokenText(responseText, request.Model)
		usage.TotalTokens = promptTokens + usage.CompletionTokens

	} else {
		var geminiResponse = &GeminiChatResponse{
			Usage: &types.Usage{
				PromptTokens: promptTokens,
			},
		}
		errWithCode = p.SendRequest(req, geminiResponse, false)
		if errWithCode != nil {
			return
		}

		usage = geminiResponse.Usage
	}
	return

}

func (p *GeminiProvider) streamResponseClaude2OpenAI(geminiResponse *GeminiChatResponse) *types.ChatCompletionStreamResponse {
	var choice types.ChatCompletionStreamChoice
	choice.Delta.Content = geminiResponse.GetResponseText()
	choice.FinishReason = &base.StopFinishReason
	var response types.ChatCompletionStreamResponse
	response.Object = "chat.completion.chunk"
	response.Model = "gemini"
	response.Choices = []types.ChatCompletionStreamChoice{choice}
	return &response
}

func (p *GeminiProvider) sendStreamRequest(req *http.Request) (*types.OpenAIErrorWithStatusCode, string) {
	defer req.Body.Close()

	// 发送请求
	resp, err := common.HttpClient.Do(req)
	if err != nil {
		return common.ErrorWrapper(err, "http_request_failed", http.StatusInternalServerError), ""
	}

	if common.IsFailureStatusCode(resp) {
		return common.HandleErrorResp(resp), ""
	}

	defer resp.Body.Close()

	responseText := ""
	dataChan := make(chan string)
	stopChan := make(chan bool)
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
	go func() {
		for scanner.Scan() {
			data := scanner.Text()
			data = strings.TrimSpace(data)
			if !strings.HasPrefix(data, "\"text\": \"") {
				continue
			}
			data = strings.TrimPrefix(data, "\"text\": \"")
			data = strings.TrimSuffix(data, "\"")
			dataChan <- data
		}
		stopChan <- true
	}()
	common.SetEventStreamHeaders(p.Context)
	p.Context.Stream(func(w io.Writer) bool {
		select {
		case data := <-dataChan:
			// this is used to prevent annoying \ related format bug
			data = fmt.Sprintf("{\"content\": \"%s\"}", data)
			type dummyStruct struct {
				Content string `json:"content"`
			}
			var dummy dummyStruct
			err := json.Unmarshal([]byte(data), &dummy)
			responseText += dummy.Content
			var choice types.ChatCompletionStreamChoice
			choice.Delta.Content = dummy.Content
			response := types.ChatCompletionStreamResponse{
				ID:      fmt.Sprintf("chatcmpl-%s", common.GetUUID()),
				Object:  "chat.completion.chunk",
				Created: common.GetTimestamp(),
				Model:   "gemini-pro",
				Choices: []types.ChatCompletionStreamChoice{choice},
			}
			jsonResponse, err := json.Marshal(response)
			if err != nil {
				common.SysError("error marshalling stream response: " + err.Error())
				return true
			}
			p.Context.Render(-1, common.CustomEvent{Data: "data: " + string(jsonResponse)})
			return true
		case <-stopChan:
			p.Context.Render(-1, common.CustomEvent{Data: "data: [DONE]"})
			return false
		}
	})

	return nil, responseText
}
