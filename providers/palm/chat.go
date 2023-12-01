package palm

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"one-api/common"
	"one-api/providers/base"
	"one-api/types"
)

func (palmResponse *PaLMChatResponse) ResponseHandler(resp *http.Response) (OpenAIResponse any, errWithCode *types.OpenAIErrorWithStatusCode) {
	if palmResponse.Error.Code != 0 || len(palmResponse.Candidates) == 0 {
		return nil, &types.OpenAIErrorWithStatusCode{
			OpenAIError: types.OpenAIError{
				Message: palmResponse.Error.Message,
				Type:    palmResponse.Error.Status,
				Param:   "",
				Code:    palmResponse.Error.Code,
			},
			StatusCode: resp.StatusCode,
		}
	}

	fullTextResponse := types.ChatCompletionResponse{
		Choices: make([]types.ChatCompletionChoice, 0, len(palmResponse.Candidates)),
	}
	for i, candidate := range palmResponse.Candidates {
		choice := types.ChatCompletionChoice{
			Index: i,
			Message: types.ChatCompletionMessage{
				Role:    "assistant",
				Content: candidate.Content,
			},
			FinishReason: "stop",
		}
		fullTextResponse.Choices = append(fullTextResponse.Choices, choice)
	}

	completionTokens := common.CountTokenText(palmResponse.Candidates[0].Content, palmResponse.Model)
	palmResponse.Usage.CompletionTokens = completionTokens
	palmResponse.Usage.TotalTokens = palmResponse.Usage.PromptTokens + completionTokens

	fullTextResponse.Usage = palmResponse.Usage

	return fullTextResponse, nil
}

func (p *PalmProvider) getChatRequestBody(request *types.ChatCompletionRequest) *PaLMChatRequest {
	palmRequest := PaLMChatRequest{
		Prompt: PaLMPrompt{
			Messages: make([]PaLMChatMessage, 0, len(request.Messages)),
		},
		Temperature:    request.Temperature,
		CandidateCount: request.N,
		TopP:           request.TopP,
		TopK:           request.MaxTokens,
	}
	for _, message := range request.Messages {
		palmMessage := PaLMChatMessage{
			Content: message.StringContent(),
		}
		if message.Role == "user" {
			palmMessage.Author = "0"
		} else {
			palmMessage.Author = "1"
		}
		palmRequest.Prompt.Messages = append(palmRequest.Prompt.Messages, palmMessage)
	}
	return &palmRequest
}

func (p *PalmProvider) ChatAction(request *types.ChatCompletionRequest, isModelMapped bool, promptTokens int) (usage *types.Usage, errWithCode *types.OpenAIErrorWithStatusCode) {
	requestBody := p.getChatRequestBody(request)
	fullRequestURL := p.GetFullRequestURL(p.ChatCompletions, request.Model)
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
		var palmChatResponse = &PaLMChatResponse{
			Model: request.Model,
			Usage: &types.Usage{
				PromptTokens: promptTokens,
			},
		}
		errWithCode = p.SendRequest(req, palmChatResponse, false)
		if errWithCode != nil {
			return
		}

		usage = palmChatResponse.Usage
	}
	return

}

func (p *PalmProvider) streamResponsePaLM2OpenAI(palmResponse *PaLMChatResponse) *types.ChatCompletionStreamResponse {
	var choice types.ChatCompletionStreamChoice
	if len(palmResponse.Candidates) > 0 {
		choice.Delta.Content = palmResponse.Candidates[0].Content
	}
	choice.FinishReason = &base.StopFinishReason
	var response types.ChatCompletionStreamResponse
	response.Object = "chat.completion.chunk"
	response.Model = "palm2"
	response.Choices = []types.ChatCompletionStreamChoice{choice}
	return &response
}

func (p *PalmProvider) sendStreamRequest(req *http.Request) (*types.OpenAIErrorWithStatusCode, string) {
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
	responseId := fmt.Sprintf("chatcmpl-%s", common.GetUUID())
	createdTime := common.GetTimestamp()
	dataChan := make(chan string)
	stopChan := make(chan bool)
	go func() {
		responseBody, err := io.ReadAll(resp.Body)
		if err != nil {
			common.SysError("error reading stream response: " + err.Error())
			stopChan <- true
			return
		}
		err = resp.Body.Close()
		if err != nil {
			common.SysError("error closing stream response: " + err.Error())
			stopChan <- true
			return
		}
		var palmResponse PaLMChatResponse
		err = json.Unmarshal(responseBody, &palmResponse)
		if err != nil {
			common.SysError("error unmarshalling stream response: " + err.Error())
			stopChan <- true
			return
		}
		fullTextResponse := p.streamResponsePaLM2OpenAI(&palmResponse)
		fullTextResponse.ID = responseId
		fullTextResponse.Created = createdTime
		if len(palmResponse.Candidates) > 0 {
			responseText = palmResponse.Candidates[0].Content
		}
		jsonResponse, err := json.Marshal(fullTextResponse)
		if err != nil {
			common.SysError("error marshalling stream response: " + err.Error())
			stopChan <- true
			return
		}
		dataChan <- string(jsonResponse)
		stopChan <- true
	}()
	common.SetEventStreamHeaders(p.Context)
	p.Context.Stream(func(w io.Writer) bool {
		select {
		case data := <-dataChan:
			p.Context.Render(-1, common.CustomEvent{Data: "data: " + data})
			return true
		case <-stopChan:
			p.Context.Render(-1, common.CustomEvent{Data: "data: [DONE]"})
			return false
		}
	})

	return nil, responseText
}
