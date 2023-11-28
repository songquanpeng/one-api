package providers

import (
	"bufio"
	"encoding/json"
	"io"
	"net/http"
	"one-api/common"
	"one-api/types"
	"strings"
)

type BaiduMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type BaiduChatRequest struct {
	Messages []BaiduMessage `json:"messages"`
	Stream   bool           `json:"stream"`
	UserId   string         `json:"user_id,omitempty"`
}

type BaiduChatResponse struct {
	Id               string       `json:"id"`
	Object           string       `json:"object"`
	Created          int64        `json:"created"`
	Result           string       `json:"result"`
	IsTruncated      bool         `json:"is_truncated"`
	NeedClearHistory bool         `json:"need_clear_history"`
	Usage            *types.Usage `json:"usage"`
	BaiduError
}

func (baiduResponse *BaiduChatResponse) requestHandler(resp *http.Response) (OpenAIResponse any, openAIErrorWithStatusCode *types.OpenAIErrorWithStatusCode) {
	if baiduResponse.ErrorMsg != "" {
		return nil, &types.OpenAIErrorWithStatusCode{
			OpenAIError: types.OpenAIError{
				Message: baiduResponse.ErrorMsg,
				Type:    "baidu_error",
				Param:   "",
				Code:    baiduResponse.ErrorCode,
			},
			StatusCode: resp.StatusCode,
		}
	}

	choice := types.ChatCompletionChoice{
		Index: 0,
		Message: types.ChatCompletionMessage{
			Role:    "assistant",
			Content: baiduResponse.Result,
		},
		FinishReason: "stop",
	}

	fullTextResponse := types.ChatCompletionResponse{
		ID:      baiduResponse.Id,
		Object:  "chat.completion",
		Created: baiduResponse.Created,
		Choices: []types.ChatCompletionChoice{choice},
		Usage:   baiduResponse.Usage,
	}

	return fullTextResponse, nil
}

type BaiduChatStreamResponse struct {
	BaiduChatResponse
	SentenceId int  `json:"sentence_id"`
	IsEnd      bool `json:"is_end"`
}

type BaiduError struct {
	ErrorCode int    `json:"error_code"`
	ErrorMsg  string `json:"error_msg"`
}

func (p *BaiduProvider) getChatRequestBody(request *types.ChatCompletionRequest) *BaiduChatRequest {
	messages := make([]BaiduMessage, 0, len(request.Messages))
	for _, message := range request.Messages {
		if message.Role == "system" {
			messages = append(messages, BaiduMessage{
				Role:    "user",
				Content: message.StringContent(),
			})
			messages = append(messages, BaiduMessage{
				Role:    "assistant",
				Content: "Okay",
			})
		} else {
			messages = append(messages, BaiduMessage{
				Role:    message.Role,
				Content: message.StringContent(),
			})
		}
	}
	return &BaiduChatRequest{
		Messages: messages,
		Stream:   request.Stream,
	}
}

func (p *BaiduProvider) ChatCompleteResponse(request *types.ChatCompletionRequest, isModelMapped bool, promptTokens int) (usage *types.Usage, openAIErrorWithStatusCode *types.OpenAIErrorWithStatusCode) {
	requestBody := p.getChatRequestBody(request)
	fullRequestURL := p.GetFullRequestURL(p.ChatCompletions, request.Model)
	if fullRequestURL == "" {
		return nil, types.ErrorWrapper(nil, "invalid_baidu_config", http.StatusInternalServerError)
	}

	headers := p.GetRequestHeaders()
	if request.Stream {
		headers["Accept"] = "text/event-stream"
	}

	client := common.NewClient()
	req, err := client.NewRequest(p.Context.Request.Method, fullRequestURL, common.WithBody(requestBody), common.WithHeader(headers))
	if err != nil {
		return nil, types.ErrorWrapper(err, "new_request_failed", http.StatusInternalServerError)
	}

	if request.Stream {
		openAIErrorWithStatusCode, usage = p.sendStreamRequest(req)
		if openAIErrorWithStatusCode != nil {
			return
		}

	} else {
		baiduChatRequest := &BaiduChatResponse{}
		openAIErrorWithStatusCode = p.sendRequest(req, baiduChatRequest)
		if openAIErrorWithStatusCode != nil {
			return
		}

		usage = baiduChatRequest.Usage
	}
	return

}

func (p *BaiduProvider) streamResponseBaidu2OpenAI(baiduResponse *BaiduChatStreamResponse) *types.ChatCompletionStreamResponse {
	var choice types.ChatCompletionStreamChoice
	choice.Delta.Content = baiduResponse.Result
	if baiduResponse.IsEnd {
		choice.FinishReason = &stopFinishReason
	}

	response := types.ChatCompletionStreamResponse{
		ID:      baiduResponse.Id,
		Object:  "chat.completion.chunk",
		Created: baiduResponse.Created,
		Model:   "ernie-bot",
		Choices: []types.ChatCompletionStreamChoice{choice},
	}
	return &response
}

func (p *BaiduProvider) sendStreamRequest(req *http.Request) (openAIErrorWithStatusCode *types.OpenAIErrorWithStatusCode, usage *types.Usage) {
	usage = &types.Usage{}
	// 发送请求
	resp, err := common.HttpClient.Do(req)
	if err != nil {
		return types.ErrorWrapper(err, "http_request_failed", http.StatusInternalServerError), nil
	}

	if common.IsFailureStatusCode(resp) {
		return p.handleErrorResp(resp), nil
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
			if len(data) < 6 { // ignore blank line or wrong format
				continue
			}
			data = data[6:]
			dataChan <- data
		}
		stopChan <- true
	}()
	setEventStreamHeaders(p.Context)
	p.Context.Stream(func(w io.Writer) bool {
		select {
		case data := <-dataChan:
			var baiduResponse BaiduChatStreamResponse
			err := json.Unmarshal([]byte(data), &baiduResponse)
			if err != nil {
				common.SysError("error unmarshalling stream response: " + err.Error())
				return true
			}
			if baiduResponse.Usage.TotalTokens != 0 {
				usage.TotalTokens = baiduResponse.Usage.TotalTokens
				usage.PromptTokens = baiduResponse.Usage.PromptTokens
				usage.CompletionTokens = baiduResponse.Usage.TotalTokens - baiduResponse.Usage.PromptTokens
			}
			response := p.streamResponseBaidu2OpenAI(&baiduResponse)
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

	return nil, usage
}
