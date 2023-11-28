package providers

import (
	"bufio"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"one-api/common"
	"one-api/types"
	"strings"
)

type TencentMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type TencentChatRequest struct {
	AppId    int64  `json:"app_id"`    // 腾讯云账号的 APPID
	SecretId string `json:"secret_id"` // 官网 SecretId
	// Timestamp当前 UNIX 时间戳，单位为秒，可记录发起 API 请求的时间。
	// 例如1529223702，如果与当前时间相差过大，会引起签名过期错误
	Timestamp int64 `json:"timestamp"`
	// Expired 签名的有效期，是一个符合 UNIX Epoch 时间戳规范的数值，
	// 单位为秒；Expired 必须大于 Timestamp 且 Expired-Timestamp 小于90天
	Expired int64  `json:"expired"`
	QueryID string `json:"query_id"` //请求 Id，用于问题排查
	// Temperature 较高的数值会使输出更加随机，而较低的数值会使其更加集中和确定
	// 默认 1.0，取值区间为[0.0,2.0]，非必要不建议使用,不合理的取值会影响效果
	// 建议该参数和 top_p 只设置1个，不要同时更改 top_p
	Temperature float64 `json:"temperature"`
	// TopP 影响输出文本的多样性，取值越大，生成文本的多样性越强
	// 默认1.0，取值区间为[0.0, 1.0]，非必要不建议使用, 不合理的取值会影响效果
	// 建议该参数和 temperature 只设置1个，不要同时更改
	TopP float64 `json:"top_p"`
	// Stream 0：同步，1：流式 （默认，协议：SSE)
	// 同步请求超时：60s，如果内容较长建议使用流式
	Stream int `json:"stream"`
	// Messages 会话内容, 长度最多为40, 按对话时间从旧到新在数组中排列
	// 输入 content 总数最大支持 3000 token。
	Messages []TencentMessage `json:"messages"`
}

type TencentUsage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
	TotalTokens  int `json:"total_tokens"`
}

type TencentResponseChoices struct {
	FinishReason string         `json:"finish_reason,omitempty"` // 流式结束标志位，为 stop 则表示尾包
	Messages     TencentMessage `json:"messages,omitempty"`      // 内容，同步模式返回内容，流模式为 null 输出 content 内容总数最多支持 1024token。
	Delta        TencentMessage `json:"delta,omitempty"`         // 内容，流模式返回内容，同步模式为 null 输出 content 内容总数最多支持 1024token。
}

type TencentChatResponse struct {
	Choices []TencentResponseChoices `json:"choices,omitempty"` // 结果
	Created string                   `json:"created,omitempty"` // unix 时间戳的字符串
	Id      string                   `json:"id,omitempty"`      // 会话 id
	Usage   *types.Usage             `json:"usage,omitempty"`   // token 数量
	Error   TencentError             `json:"error,omitempty"`   // 错误信息 注意：此字段可能返回 null，表示取不到有效值
	Note    string                   `json:"note,omitempty"`    // 注释
	ReqID   string                   `json:"req_id,omitempty"`  // 唯一请求 Id，每次请求都会返回。用于反馈接口入参
}

func (TencentResponse *TencentChatResponse) requestHandler(resp *http.Response) (OpenAIResponse any, openAIErrorWithStatusCode *types.OpenAIErrorWithStatusCode) {
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

func (p *TencentProvider) ChatCompleteResponse(request *types.ChatCompletionRequest, isModelMapped bool, promptTokens int) (usage *types.Usage, openAIErrorWithStatusCode *types.OpenAIErrorWithStatusCode) {
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
		openAIErrorWithStatusCode, responseText = p.sendStreamRequest(req)
		if openAIErrorWithStatusCode != nil {
			return
		}

		usage.PromptTokens = promptTokens
		usage.CompletionTokens = common.CountTokenText(responseText, request.Model)
		usage.TotalTokens = promptTokens + usage.CompletionTokens

	} else {
		tencentResponse := &TencentChatResponse{}
		openAIErrorWithStatusCode = p.sendRequest(req, tencentResponse)
		if openAIErrorWithStatusCode != nil {
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
			choice.FinishReason = &stopFinishReason
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
		return p.handleErrorResp(resp), ""
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
	setEventStreamHeaders(p.Context)
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
