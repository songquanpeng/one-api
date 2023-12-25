package zhipu

import (
	"bufio"
	"encoding/json"
	"io"
	"net/http"
	"one-api/common"
	"one-api/providers/base"
	"one-api/types"
	"strings"
)

func (zhipuResponse *ZhipuResponse) ResponseHandler(resp *http.Response) (OpenAIResponse any, errWithCode *types.OpenAIErrorWithStatusCode) {
	if !zhipuResponse.Success {
		return &types.OpenAIErrorWithStatusCode{
			OpenAIError: types.OpenAIError{
				Message: zhipuResponse.Msg,
				Type:    "zhipu_error",
				Param:   "",
				Code:    zhipuResponse.Code,
			},
			StatusCode: resp.StatusCode,
		}, nil
	}

	fullTextResponse := types.ChatCompletionResponse{
		ID:      zhipuResponse.Data.TaskId,
		Object:  "chat.completion",
		Created: common.GetTimestamp(),
		Model:   zhipuResponse.Model,
		Choices: make([]types.ChatCompletionChoice, 0, len(zhipuResponse.Data.Choices)),
		Usage:   &zhipuResponse.Data.Usage,
	}
	for i, choice := range zhipuResponse.Data.Choices {
		openaiChoice := types.ChatCompletionChoice{
			Index: i,
			Message: types.ChatCompletionMessage{
				Role:    choice.Role,
				Content: strings.Trim(choice.Content, "\""),
			},
			FinishReason: "",
		}
		if i == len(zhipuResponse.Data.Choices)-1 {
			openaiChoice.FinishReason = "stop"
		}
		fullTextResponse.Choices = append(fullTextResponse.Choices, openaiChoice)
	}
	return fullTextResponse, nil

}

func (p *ZhipuProvider) getChatRequestBody(request *types.ChatCompletionRequest) *ZhipuRequest {
	messages := make([]ZhipuMessage, 0, len(request.Messages))
	for _, message := range request.Messages {
		if message.Role == "system" {
			messages = append(messages, ZhipuMessage{
				Role:    "system",
				Content: message.StringContent(),
			})
			messages = append(messages, ZhipuMessage{
				Role:    "user",
				Content: "Okay",
			})
		} else {
			messages = append(messages, ZhipuMessage{
				Role:    message.Role,
				Content: message.StringContent(),
			})
		}
	}
	return &ZhipuRequest{
		Prompt:      messages,
		Temperature: request.Temperature,
		TopP:        request.TopP,
		Incremental: false,
	}
}

func (p *ZhipuProvider) ChatAction(request *types.ChatCompletionRequest, isModelMapped bool, promptTokens int) (usage *types.Usage, errWithCode *types.OpenAIErrorWithStatusCode) {
	requestBody := p.getChatRequestBody(request)
	fullRequestURL := p.GetFullRequestURL(p.ChatCompletions, request.Model)
	headers := p.GetRequestHeaders()
	if request.Stream {
		headers["Accept"] = "text/event-stream"
		fullRequestURL += "/sse-invoke"
	} else {
		fullRequestURL += "/invoke"
	}

	client := common.NewClient()
	req, err := client.NewRequest(p.Context.Request.Method, fullRequestURL, common.WithBody(requestBody), common.WithHeader(headers))
	if err != nil {
		return nil, common.ErrorWrapper(err, "new_request_failed", http.StatusInternalServerError)
	}

	if request.Stream {
		errWithCode, usage = p.sendStreamRequest(req, request.Model)
		if errWithCode != nil {
			return
		}

	} else {
		zhipuResponse := &ZhipuResponse{
			Model: request.Model,
		}
		errWithCode = p.SendRequest(req, zhipuResponse, false)
		if errWithCode != nil {
			return
		}

		usage = &zhipuResponse.Data.Usage
	}
	return

}

func (p *ZhipuProvider) streamResponseZhipu2OpenAI(zhipuResponse string) *types.ChatCompletionStreamResponse {
	var choice types.ChatCompletionStreamChoice
	choice.Delta.Content = zhipuResponse
	response := types.ChatCompletionStreamResponse{
		Object:  "chat.completion.chunk",
		Created: common.GetTimestamp(),
		Model:   "chatglm",
		Choices: []types.ChatCompletionStreamChoice{choice},
	}
	return &response
}

func (p *ZhipuProvider) streamMetaResponseZhipu2OpenAI(zhipuResponse *ZhipuStreamMetaResponse) (*types.ChatCompletionStreamResponse, *types.Usage) {
	var choice types.ChatCompletionStreamChoice
	choice.Delta.Content = ""
	choice.FinishReason = &base.StopFinishReason
	response := types.ChatCompletionStreamResponse{
		ID:      zhipuResponse.RequestId,
		Object:  "chat.completion.chunk",
		Created: common.GetTimestamp(),
		Model:   zhipuResponse.Model,
		Choices: []types.ChatCompletionStreamChoice{choice},
	}
	return &response, &zhipuResponse.Usage
}

func (p *ZhipuProvider) sendStreamRequest(req *http.Request, model string) (*types.OpenAIErrorWithStatusCode, *types.Usage) {
	defer req.Body.Close()

	// 发送请求
	resp, err := common.HttpClient.Do(req)
	if err != nil {
		return common.ErrorWrapper(err, "http_request_failed", http.StatusInternalServerError), nil
	}

	if common.IsFailureStatusCode(resp) {
		return common.HandleErrorResp(resp), nil
	}

	defer resp.Body.Close()

	var usage *types.Usage
	scanner := bufio.NewScanner(resp.Body)
	scanner.Split(func(data []byte, atEOF bool) (advance int, token []byte, err error) {
		if atEOF && len(data) == 0 {
			return 0, nil, nil
		}
		if i := strings.Index(string(data), "\n\n"); i >= 0 && strings.Contains(string(data), ":") {
			return i + 2, data[0:i], nil
		}
		if atEOF {
			return len(data), data, nil
		}
		return 0, nil, nil
	})
	dataChan := make(chan string)
	metaChan := make(chan string)
	stopChan := make(chan bool)
	go func() {
		for scanner.Scan() {
			data := scanner.Text()
			lines := strings.Split(data, "\n")
			for i, line := range lines {
				if len(line) < 5 {
					continue
				}
				if line[:5] == "data:" {
					dataChan <- line[5:]
					if i != len(lines)-1 {
						dataChan <- "\n"
					}
				} else if line[:5] == "meta:" {
					metaChan <- line[5:]
				}
			}
		}
		stopChan <- true
	}()
	common.SetEventStreamHeaders(p.Context)
	p.Context.Stream(func(w io.Writer) bool {
		select {
		case data := <-dataChan:
			response := p.streamResponseZhipu2OpenAI(data)
			response.Model = model
			jsonResponse, err := json.Marshal(response)
			if err != nil {
				common.SysError("error marshalling stream response: " + err.Error())
				return true
			}
			p.Context.Render(-1, common.CustomEvent{Data: "data: " + string(jsonResponse)})
			return true
		case data := <-metaChan:
			var zhipuResponse ZhipuStreamMetaResponse
			err := json.Unmarshal([]byte(data), &zhipuResponse)
			if err != nil {
				common.SysError("error unmarshalling stream response: " + err.Error())
				return true
			}
			zhipuResponse.Model = model
			response, zhipuUsage := p.streamMetaResponseZhipu2OpenAI(&zhipuResponse)
			jsonResponse, err := json.Marshal(response)
			if err != nil {
				common.SysError("error marshalling stream response: " + err.Error())
				return true
			}
			usage = zhipuUsage
			p.Context.Render(-1, common.CustomEvent{Data: "data: " + string(jsonResponse)})
			return true
		case <-stopChan:
			p.Context.Render(-1, common.CustomEvent{Data: "data: [DONE]"})
			return false
		}
	})
	return nil, usage
}
