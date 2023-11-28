package providers

import (
	"encoding/json"
	"io"
	"net/http"
	"one-api/common"
	"one-api/types"
	"time"

	"github.com/gorilla/websocket"
)

type XunfeiMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type XunfeiChatRequest struct {
	Header struct {
		AppId string `json:"app_id"`
	} `json:"header"`
	Parameter struct {
		Chat struct {
			Domain      string  `json:"domain,omitempty"`
			Temperature float64 `json:"temperature,omitempty"`
			TopK        int     `json:"top_k,omitempty"`
			MaxTokens   int     `json:"max_tokens,omitempty"`
			Auditing    bool    `json:"auditing,omitempty"`
		} `json:"chat"`
	} `json:"parameter"`
	Payload struct {
		Message struct {
			Text []XunfeiMessage `json:"text"`
		} `json:"message"`
	} `json:"payload"`
}

type XunfeiChatResponseTextItem struct {
	Content string `json:"content"`
	Role    string `json:"role"`
	Index   int    `json:"index"`
}

type XunfeiChatResponse struct {
	Header struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Sid     string `json:"sid"`
		Status  int    `json:"status"`
	} `json:"header"`
	Payload struct {
		Choices struct {
			Status int                          `json:"status"`
			Seq    int                          `json:"seq"`
			Text   []XunfeiChatResponseTextItem `json:"text"`
		} `json:"choices"`
		Usage struct {
			//Text struct {
			//	QuestionTokens   string `json:"question_tokens"`
			//	PromptTokens     string `json:"prompt_tokens"`
			//	CompletionTokens string `json:"completion_tokens"`
			//	TotalTokens      string `json:"total_tokens"`
			//} `json:"text"`
			Text types.Usage `json:"text"`
		} `json:"usage"`
	} `json:"payload"`
}

func (p *XunfeiProvider) ChatCompleteResponse(request *types.ChatCompletionRequest, isModelMapped bool, promptTokens int) (usage *types.Usage, openAIErrorWithStatusCode *types.OpenAIErrorWithStatusCode) {
	authUrl := p.GetFullRequestURL(p.ChatCompletions, request.Model)

	if request.Stream {
		return p.sendStreamRequest(request, authUrl)
	} else {
		return p.sendRequest(request, authUrl)
	}
}

func (p *XunfeiProvider) sendRequest(request *types.ChatCompletionRequest, authUrl string) (usage *types.Usage, openAIErrorWithStatusCode *types.OpenAIErrorWithStatusCode) {
	usage = &types.Usage{}
	dataChan, stopChan, err := p.xunfeiMakeRequest(request, authUrl)
	if err != nil {
		return nil, types.ErrorWrapper(err, "make xunfei request err", http.StatusInternalServerError)
	}

	var content string
	var xunfeiResponse XunfeiChatResponse
	stop := false
	for !stop {
		select {
		case xunfeiResponse = <-dataChan:
			if len(xunfeiResponse.Payload.Choices.Text) == 0 {
				continue
			}
			content += xunfeiResponse.Payload.Choices.Text[0].Content
			usage.PromptTokens += xunfeiResponse.Payload.Usage.Text.PromptTokens
			usage.CompletionTokens += xunfeiResponse.Payload.Usage.Text.CompletionTokens
			usage.TotalTokens += xunfeiResponse.Payload.Usage.Text.TotalTokens
		case stop = <-stopChan:
		}
	}

	xunfeiResponse.Payload.Choices.Text[0].Content = content

	response := p.responseXunfei2OpenAI(&xunfeiResponse)
	jsonResponse, err := json.Marshal(response)
	if err != nil {
		return nil, types.ErrorWrapper(err, "marshal_response_body_failed", http.StatusInternalServerError)
	}
	p.Context.Writer.Header().Set("Content-Type", "application/json")
	_, _ = p.Context.Writer.Write(jsonResponse)
	return usage, nil
}

func (p *XunfeiProvider) sendStreamRequest(request *types.ChatCompletionRequest, authUrl string) (usage *types.Usage, openAIErrorWithStatusCode *types.OpenAIErrorWithStatusCode) {
	usage = &types.Usage{}
	dataChan, stopChan, err := p.xunfeiMakeRequest(request, authUrl)
	if err != nil {
		return nil, types.ErrorWrapper(err, "make xunfei request err", http.StatusInternalServerError)
	}
	setEventStreamHeaders(p.Context)
	p.Context.Stream(func(w io.Writer) bool {
		select {
		case xunfeiResponse := <-dataChan:
			usage.PromptTokens += xunfeiResponse.Payload.Usage.Text.PromptTokens
			usage.CompletionTokens += xunfeiResponse.Payload.Usage.Text.CompletionTokens
			usage.TotalTokens += xunfeiResponse.Payload.Usage.Text.TotalTokens
			response := p.streamResponseXunfei2OpenAI(&xunfeiResponse)
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
	return usage, nil
}

func (p *XunfeiProvider) requestOpenAI2Xunfei(request *types.ChatCompletionRequest) *XunfeiChatRequest {
	messages := make([]XunfeiMessage, 0, len(request.Messages))
	for _, message := range request.Messages {
		if message.Role == "system" {
			messages = append(messages, XunfeiMessage{
				Role:    "user",
				Content: message.StringContent(),
			})
			messages = append(messages, XunfeiMessage{
				Role:    "assistant",
				Content: "Okay",
			})
		} else {
			messages = append(messages, XunfeiMessage{
				Role:    message.Role,
				Content: message.StringContent(),
			})
		}
	}
	xunfeiRequest := XunfeiChatRequest{}
	xunfeiRequest.Header.AppId = p.apiId
	xunfeiRequest.Parameter.Chat.Domain = p.domain
	xunfeiRequest.Parameter.Chat.Temperature = request.Temperature
	xunfeiRequest.Parameter.Chat.TopK = request.N
	xunfeiRequest.Parameter.Chat.MaxTokens = request.MaxTokens
	xunfeiRequest.Payload.Message.Text = messages
	return &xunfeiRequest
}

func (p *XunfeiProvider) responseXunfei2OpenAI(response *XunfeiChatResponse) *types.ChatCompletionResponse {
	if len(response.Payload.Choices.Text) == 0 {
		response.Payload.Choices.Text = []XunfeiChatResponseTextItem{
			{
				Content: "",
			},
		}
	}
	choice := types.ChatCompletionChoice{
		Index: 0,
		Message: types.ChatCompletionMessage{
			Role:    "assistant",
			Content: response.Payload.Choices.Text[0].Content,
		},
		FinishReason: stopFinishReason,
	}
	fullTextResponse := types.ChatCompletionResponse{
		Object:  "chat.completion",
		Created: common.GetTimestamp(),
		Choices: []types.ChatCompletionChoice{choice},
		Usage:   &response.Payload.Usage.Text,
	}
	return &fullTextResponse
}

func (p *XunfeiProvider) xunfeiMakeRequest(textRequest *types.ChatCompletionRequest, authUrl string) (chan XunfeiChatResponse, chan bool, error) {
	d := websocket.Dialer{
		HandshakeTimeout: 5 * time.Second,
	}
	conn, resp, err := d.Dial(authUrl, nil)
	if err != nil || resp.StatusCode != 101 {
		return nil, nil, err
	}
	data := p.requestOpenAI2Xunfei(textRequest)
	err = conn.WriteJSON(data)
	if err != nil {
		return nil, nil, err
	}

	dataChan := make(chan XunfeiChatResponse)
	stopChan := make(chan bool)
	go func() {
		for {
			_, msg, err := conn.ReadMessage()
			if err != nil {
				common.SysError("error reading stream response: " + err.Error())
				break
			}
			var response XunfeiChatResponse
			err = json.Unmarshal(msg, &response)
			if err != nil {
				common.SysError("error unmarshalling stream response: " + err.Error())
				break
			}
			dataChan <- response
			if response.Payload.Choices.Status == 2 {
				err := conn.Close()
				if err != nil {
					common.SysError("error closing websocket connection: " + err.Error())
				}
				break
			}
		}
		stopChan <- true
	}()

	return dataChan, stopChan, nil
}

func (p *XunfeiProvider) streamResponseXunfei2OpenAI(xunfeiResponse *XunfeiChatResponse) *types.ChatCompletionStreamResponse {
	if len(xunfeiResponse.Payload.Choices.Text) == 0 {
		xunfeiResponse.Payload.Choices.Text = []XunfeiChatResponseTextItem{
			{
				Content: "",
			},
		}
	}
	var choice types.ChatCompletionStreamChoice
	choice.Delta.Content = xunfeiResponse.Payload.Choices.Text[0].Content
	if xunfeiResponse.Payload.Choices.Status == 2 {
		choice.FinishReason = &stopFinishReason
	}
	response := types.ChatCompletionStreamResponse{
		Object:  "chat.completion.chunk",
		Created: common.GetTimestamp(),
		Model:   "SparkDesk",
		Choices: []types.ChatCompletionStreamChoice{choice},
	}
	return &response
}
