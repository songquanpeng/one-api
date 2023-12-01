package ali

import (
	"bufio"
	"encoding/json"
	"io"
	"net/http"
	"one-api/common"
	"one-api/types"
	"strings"
)

// 阿里云响应处理
func (aliResponse *AliChatResponse) ResponseHandler(resp *http.Response) (OpenAIResponse any, errWithCode *types.OpenAIErrorWithStatusCode) {
	if aliResponse.Code != "" {
		errWithCode = &types.OpenAIErrorWithStatusCode{
			OpenAIError: types.OpenAIError{
				Message: aliResponse.Message,
				Type:    aliResponse.Code,
				Param:   aliResponse.RequestId,
				Code:    aliResponse.Code,
			},
			StatusCode: resp.StatusCode,
		}

		return
	}

	choice := types.ChatCompletionChoice{
		Index: 0,
		Message: types.ChatCompletionMessage{
			Role:    "assistant",
			Content: aliResponse.Output.Text,
		},
		FinishReason: aliResponse.Output.FinishReason,
	}

	OpenAIResponse = types.ChatCompletionResponse{
		ID:      aliResponse.RequestId,
		Object:  "chat.completion",
		Created: common.GetTimestamp(),
		Choices: []types.ChatCompletionChoice{choice},
		Usage: &types.Usage{
			PromptTokens:     aliResponse.Usage.InputTokens,
			CompletionTokens: aliResponse.Usage.OutputTokens,
			TotalTokens:      aliResponse.Usage.InputTokens + aliResponse.Usage.OutputTokens,
		},
	}

	return
}

// 获取聊天请求体
func (p *AliProvider) getChatRequestBody(request *types.ChatCompletionRequest) *AliChatRequest {
	messages := make([]AliMessage, 0, len(request.Messages))
	prompt := ""
	for i := 0; i < len(request.Messages); i++ {
		message := request.Messages[i]
		if message.Role == "system" {
			messages = append(messages, AliMessage{
				User: message.StringContent(),
				Bot:  "Okay",
			})
			continue
		} else {
			if i == len(request.Messages)-1 {
				prompt = message.StringContent()
				break
			}
			messages = append(messages, AliMessage{
				User: message.StringContent(),
				Bot:  request.Messages[i+1].StringContent(),
			})
			i++
		}
	}
	return &AliChatRequest{
		Model: request.Model,
		Input: AliInput{
			Prompt:  prompt,
			History: messages,
		},
	}
}

// 聊天
func (p *AliProvider) ChatAction(request *types.ChatCompletionRequest, isModelMapped bool, promptTokens int) (usage *types.Usage, errWithCode *types.OpenAIErrorWithStatusCode) {

	requestBody := p.getChatRequestBody(request)
	fullRequestURL := p.GetFullRequestURL(p.ChatCompletions, request.Model)
	headers := p.GetRequestHeaders()
	if request.Stream {
		headers["Accept"] = "text/event-stream"
		headers["X-DashScope-SSE"] = "enable"
	}

	client := common.NewClient()
	req, err := client.NewRequest(p.Context.Request.Method, fullRequestURL, common.WithBody(requestBody), common.WithHeader(headers))
	if err != nil {
		return nil, common.ErrorWrapper(err, "new_request_failed", http.StatusInternalServerError)
	}

	if request.Stream {
		usage, errWithCode = p.sendStreamRequest(req)
		if errWithCode != nil {
			return
		}

		if usage == nil {
			usage = &types.Usage{
				PromptTokens:     0,
				CompletionTokens: 0,
				TotalTokens:      0,
			}
		}

	} else {
		aliResponse := &AliChatResponse{}
		errWithCode = p.SendRequest(req, aliResponse, false)
		if errWithCode != nil {
			return
		}

		usage = &types.Usage{
			PromptTokens:     aliResponse.Usage.InputTokens,
			CompletionTokens: aliResponse.Usage.OutputTokens,
			TotalTokens:      aliResponse.Usage.InputTokens + aliResponse.Usage.OutputTokens,
		}
	}
	return
}

// 阿里云响应转OpenAI响应
func (p *AliProvider) streamResponseAli2OpenAI(aliResponse *AliChatResponse) *types.ChatCompletionStreamResponse {
	var choice types.ChatCompletionStreamChoice
	choice.Delta.Content = aliResponse.Output.Text
	if aliResponse.Output.FinishReason != "null" {
		finishReason := aliResponse.Output.FinishReason
		choice.FinishReason = &finishReason
	}

	response := types.ChatCompletionStreamResponse{
		ID:      aliResponse.RequestId,
		Object:  "chat.completion.chunk",
		Created: common.GetTimestamp(),
		Model:   "ernie-bot",
		Choices: []types.ChatCompletionStreamChoice{choice},
	}
	return &response
}

// 发送流请求
func (p *AliProvider) sendStreamRequest(req *http.Request) (usage *types.Usage, errWithCode *types.OpenAIErrorWithStatusCode) {
	defer req.Body.Close()

	usage = &types.Usage{}
	// 发送请求
	resp, err := common.HttpClient.Do(req)
	if err != nil {
		return nil, common.ErrorWrapper(err, "http_request_failed", http.StatusInternalServerError)
	}

	if common.IsFailureStatusCode(resp) {
		return nil, common.HandleErrorResp(resp)
	}

	defer resp.Body.Close()

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
	lastResponseText := ""
	p.Context.Stream(func(w io.Writer) bool {
		select {
		case data := <-dataChan:
			var aliResponse AliChatResponse
			err := json.Unmarshal([]byte(data), &aliResponse)
			if err != nil {
				common.SysError("error unmarshalling stream response: " + err.Error())
				return true
			}
			if aliResponse.Usage.OutputTokens != 0 {
				usage.PromptTokens = aliResponse.Usage.InputTokens
				usage.CompletionTokens = aliResponse.Usage.OutputTokens
				usage.TotalTokens = aliResponse.Usage.InputTokens + aliResponse.Usage.OutputTokens
			}
			response := p.streamResponseAli2OpenAI(&aliResponse)
			response.Choices[0].Delta.Content = strings.TrimPrefix(response.Choices[0].Delta.Content, lastResponseText)
			lastResponseText = aliResponse.Output.Text
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

	return
}
