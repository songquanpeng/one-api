package tencent

import (
	"bufio"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"one-api/common"
	"one-api/providers/base"
	"one-api/types"
	"strings"
)

func (TencentResponse *TencentChatResponse) ResponseHandler(resp *http.Response) (OpenAIResponse any, errWithCode *types.OpenAIErrorWithStatusCode) {
	if TencentResponse.Error.Code != 0 {
		return &types.OpenAIErrorWithStatusCode{
			OpenAIError: types.OpenAIError{
				Message: TencentResponse.Error.Message,
				Code:    TencentResponse.Error.Code,
			},
			StatusCode: resp.StatusCode,
		}, nil
	}

	fullTextResponse := types.ChatCompletionResponse{
		Object:  "chat.completion",
		Created: common.GetTimestamp(),
		Usage:   TencentResponse.Usage,
	}
	if len(TencentResponse.Choices) > 0 {
		choice := types.ChatCompletionChoice{
			Index: 0,
			Message: types.ChatCompletionMessage{
				Role:    "assistant",
				Content: TencentResponse.Choices[0].Messages.Content,
			},
			FinishReason: TencentResponse.Choices[0].FinishReason,
		}
		fullTextResponse.Choices = append(fullTextResponse.Choices, choice)
	}

	return fullTextResponse, nil
}

func (p *TencentProvider) getChatRequestBody(request *types.ChatCompletionRequest) *TencentChatRequest {
	messages := make([]TencentMessage, 0, len(request.Messages))
	for i := 0; i < len(request.Messages); i++ {
		message := request.Messages[i]
		if message.Role == "system" {
			messages = append(messages, TencentMessage{
				Role:    "user",
				Content: message.StringContent(),
			})
			messages = append(messages, TencentMessage{
				Role:    "assistant",
				Content: "Okay",
			})
			continue
		}
		messages = append(messages, TencentMessage{
			Content: message.StringContent(),
			Role:    message.Role,
		})
	}
	stream := 0
	if request.Stream {
		stream = 1
	}
	return &TencentChatRequest{
		Timestamp:   common.GetTimestamp(),
		Expired:     common.GetTimestamp() + 24*60*60,
		QueryID:     common.GetUUID(),
		Temperature: request.Temperature,
		TopP:        request.TopP,
		Stream:      stream,
		Messages:    messages,
	}
}

func (p *TencentProvider) ChatAction(request *types.ChatCompletionRequest, isModelMapped bool, promptTokens int) (usage *types.Usage, errWithCode *types.OpenAIErrorWithStatusCode) {
	requestBody := p.getChatRequestBody(request)
	sign := p.getTencentSign(*requestBody)
	if sign == "" {
		return nil, types.ErrorWrapper(errors.New("get tencent sign failed"), "get_tencent_sign_failed", http.StatusInternalServerError)
	}

	fullRequestURL := p.GetFullRequestURL(p.ChatCompletions, request.Model)
	headers := p.GetRequestHeaders()
	headers["Authorization"] = sign
	if request.Stream {
		headers["Accept"] = "text/event-stream"
	}

	client := common.NewClient()
	req, err := client.NewRequest(p.Context.Request.Method, fullRequestURL, common.WithBody(requestBody), common.WithHeader(headers))
	if err != nil {
		return nil, types.ErrorWrapper(err, "new_request_failed", http.StatusInternalServerError)
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
		tencentResponse := &TencentChatResponse{}
		errWithCode = p.SendRequest(req, tencentResponse)
		if errWithCode != nil {
			return
		}

		usage = tencentResponse.Usage
	}
	return

}

func (p *TencentProvider) streamResponseTencent2OpenAI(TencentResponse *TencentChatResponse) *types.ChatCompletionStreamResponse {
	response := types.ChatCompletionStreamResponse{
		Object:  "chat.completion.chunk",
		Created: common.GetTimestamp(),
		Model:   "tencent-hunyuan",
	}
	if len(TencentResponse.Choices) > 0 {
		var choice types.ChatCompletionStreamChoice
		choice.Delta.Content = TencentResponse.Choices[0].Delta.Content
		if TencentResponse.Choices[0].FinishReason == "stop" {
			choice.FinishReason = &base.StopFinishReason
		}
		response.Choices = append(response.Choices, choice)
	}
	return &response
}

func (p *TencentProvider) sendStreamRequest(req *http.Request) (*types.OpenAIErrorWithStatusCode, string) {
	// 发送请求
	resp, err := common.HttpClient.Do(req)
	if err != nil {
		return types.ErrorWrapper(err, "http_request_failed", http.StatusInternalServerError), ""
	}

	if common.IsFailureStatusCode(resp) {
		return p.HandleErrorResp(resp), ""
	}

	defer resp.Body.Close()

	var responseText string
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
			if len(data) < 5 { // ignore blank line or wrong format
				continue
			}
			if data[:5] != "data:" {
				continue
			}
			data = data[5:]
			dataChan <- data
		}
		stopChan <- true
	}()
	common.SetEventStreamHeaders(p.Context)
	p.Context.Stream(func(w io.Writer) bool {
		select {
		case data := <-dataChan:
			var TencentResponse TencentChatResponse
			err := json.Unmarshal([]byte(data), &TencentResponse)
			if err != nil {
				common.SysError("error unmarshalling stream response: " + err.Error())
				return true
			}
			response := p.streamResponseTencent2OpenAI(&TencentResponse)
			if len(response.Choices) != 0 {
				responseText += response.Choices[0].Delta.Content
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
