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

	OpenAIResponse = types.ChatCompletionResponse{
		ID:      aliResponse.RequestId,
		Object:  "chat.completion",
		Created: common.GetTimestamp(),
		Model:   aliResponse.Model,
		Choices: aliResponse.Output.ToChatCompletionChoices(),
		Usage: &types.Usage{
			PromptTokens:     aliResponse.Usage.InputTokens,
			CompletionTokens: aliResponse.Usage.OutputTokens,
			TotalTokens:      aliResponse.Usage.InputTokens + aliResponse.Usage.OutputTokens,
		},
	}

	return
}

const AliEnableSearchModelSuffix = "-internet"

// 获取聊天请求体
func (p *AliProvider) getChatRequestBody(request *types.ChatCompletionRequest) *AliChatRequest {
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
		usage, errWithCode = p.sendStreamRequest(req, request.Model)
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
		aliResponse := &AliChatResponse{
			Model: request.Model,
		}
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
	// chatChoice := aliResponse.Output.ToChatCompletionChoices()
	// jsonBody, _ := json.MarshalIndent(chatChoice, "", "  ")
	// fmt.Println("requestBody:", string(jsonBody))
	var choice types.ChatCompletionStreamChoice
	choice.Index = aliResponse.Output.Choices[0].Index
	choice.Delta.Content = aliResponse.Output.Choices[0].Message.StringContent()
	// fmt.Println("choice.Delta.Content:", chatChoice[0].Message)
	if aliResponse.Output.Choices[0].FinishReason != "null" {
		finishReason := aliResponse.Output.Choices[0].FinishReason
		choice.FinishReason = &finishReason
	}

	response := types.ChatCompletionStreamResponse{
		ID:      aliResponse.RequestId,
		Object:  "chat.completion.chunk",
		Created: common.GetTimestamp(),
		Model:   aliResponse.Model,
		Choices: []types.ChatCompletionStreamChoice{choice},
	}
	return &response
}

// 发送流请求
func (p *AliProvider) sendStreamRequest(req *http.Request, model string) (usage *types.Usage, errWithCode *types.OpenAIErrorWithStatusCode) {
	defer req.Body.Close()

	usage = &types.Usage{}
	// 发送请求
	client := common.GetHttpClient(p.Channel.Proxy)
	resp, err := client.Do(req)
	if err != nil {
		return nil, common.ErrorWrapper(err, "http_request_failed", http.StatusInternalServerError)
	}
	common.PutHttpClient(client)

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
	index := 0
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
			aliResponse.Model = model
			aliResponse.Output.Choices[0].Index = index
			index++
			response := p.streamResponseAli2OpenAI(&aliResponse)
			response.Choices[0].Delta.Content = strings.TrimPrefix(response.Choices[0].Delta.Content, lastResponseText)
			lastResponseText = aliResponse.Output.Choices[0].Message.StringContent()
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
