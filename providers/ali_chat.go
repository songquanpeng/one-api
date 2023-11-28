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

type AliMessage struct {
	User string `json:"user"`
	Bot  string `json:"bot"`
}

type AliInput struct {
	Prompt  string       `json:"prompt"`
	History []AliMessage `json:"history"`
}

type AliParameters struct {
	TopP         float64 `json:"top_p,omitempty"`
	TopK         int     `json:"top_k,omitempty"`
	Seed         uint64  `json:"seed,omitempty"`
	EnableSearch bool    `json:"enable_search,omitempty"`
}

type AliChatRequest struct {
	Model      string        `json:"model"`
	Input      AliInput      `json:"input"`
	Parameters AliParameters `json:"parameters,omitempty"`
}

type AliOutput struct {
	Text         string `json:"text"`
	FinishReason string `json:"finish_reason"`
}

type AliChatResponse struct {
	Output AliOutput `json:"output"`
	Usage  AliUsage  `json:"usage"`
	AliError
}

func (aliResponse *AliChatResponse) requestHandler(resp *http.Response) (OpenAIResponse any, openAIErrorWithStatusCode *types.OpenAIErrorWithStatusCode) {
	if aliResponse.Code != "" {
		return nil, &types.OpenAIErrorWithStatusCode{
			OpenAIError: types.OpenAIError{
				Message: aliResponse.Message,
				Type:    aliResponse.Code,
				Param:   aliResponse.RequestId,
				Code:    aliResponse.Code,
			},
			StatusCode: resp.StatusCode,
		}
	}

	choice := types.ChatCompletionChoice{
		Index: 0,
		Message: types.ChatCompletionMessage{
			Role:    "assistant",
			Content: aliResponse.Output.Text,
		},
		FinishReason: aliResponse.Output.FinishReason,
	}

	fullTextResponse := types.ChatCompletionResponse{
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

	return fullTextResponse, nil
}

func (p *AliAIProvider) getChatRequestBody(request *types.ChatCompletionRequest) *AliChatRequest {
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

func (p *AliAIProvider) ChatCompleteResponse(request *types.ChatCompletionRequest, isModelMapped bool, promptTokens int) (usage *types.Usage, openAIErrorWithStatusCode *types.OpenAIErrorWithStatusCode) {

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
		return nil, types.ErrorWrapper(err, "new_request_failed", http.StatusInternalServerError)
	}

	if request.Stream {
		openAIErrorWithStatusCode, usage = p.sendStreamRequest(req)
		if openAIErrorWithStatusCode != nil {
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
		openAIErrorWithStatusCode = p.sendRequest(req, aliResponse)
		if openAIErrorWithStatusCode != nil {
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

func (p *AliAIProvider) streamResponseAli2OpenAI(aliResponse *AliChatResponse) *types.ChatCompletionStreamResponse {
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

func (p *AliAIProvider) sendStreamRequest(req *http.Request) (openAIErrorWithStatusCode *types.OpenAIErrorWithStatusCode, usage *types.Usage) {
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
	setEventStreamHeaders(p.Context)
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

	return nil, usage
}
