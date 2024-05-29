package xunfei

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"one-api/common"
	"one-api/common/config"
	"one-api/common/requester"
	"one-api/common/utils"
	"one-api/types"
	"strings"

	"github.com/gorilla/websocket"
)

type xunfeiHandler struct {
	Usage   *types.Usage
	Request *types.ChatCompletionRequest
}

func (p *XunfeiProvider) CreateChatCompletion(request *types.ChatCompletionRequest) (*types.ChatCompletionResponse, *types.OpenAIErrorWithStatusCode) {
	wsConn, errWithCode := p.getChatRequest(request)
	if errWithCode != nil {
		return nil, errWithCode
	}

	xunfeiRequest := p.convertFromChatOpenai(request)

	chatHandler := &xunfeiHandler{
		Usage:   p.Usage,
		Request: request,
	}

	stream, errWithCode := requester.SendWSJsonRequest[XunfeiChatResponse](wsConn, xunfeiRequest, chatHandler.handlerNotStream)
	if errWithCode != nil {
		return nil, errWithCode
	}

	return chatHandler.convertToChatOpenai(stream)

}

func (p *XunfeiProvider) CreateChatCompletionStream(request *types.ChatCompletionRequest) (requester.StreamReaderInterface[string], *types.OpenAIErrorWithStatusCode) {
	wsConn, errWithCode := p.getChatRequest(request)
	if errWithCode != nil {
		return nil, errWithCode
	}

	xunfeiRequest := p.convertFromChatOpenai(request)

	chatHandler := &xunfeiHandler{
		Usage:   p.Usage,
		Request: request,
	}

	return requester.SendWSJsonRequest[string](wsConn, xunfeiRequest, chatHandler.handlerStream)
}

func (p *XunfeiProvider) getChatRequest(request *types.ChatCompletionRequest) (*websocket.Conn, *types.OpenAIErrorWithStatusCode) {
	_, errWithCode := p.GetSupportedAPIUri(config.RelayModeChatCompletions)
	if errWithCode != nil {
		return nil, errWithCode
	}

	authUrl := p.GetFullRequestURL(request.Model)

	wsConn, err := p.wsRequester.NewRequest(authUrl, nil)
	if err != nil {
		return nil, common.ErrorWrapper(err, "ws_request_failed", http.StatusInternalServerError)
	}

	return wsConn, nil
}

func (p *XunfeiProvider) convertFromChatOpenai(request *types.ChatCompletionRequest) *XunfeiChatRequest {
	request.ClearEmptyMessages()
	messages := make([]XunfeiMessage, 0, len(request.Messages))
	for _, message := range request.Messages {
		if message.FunctionCall != nil || message.ToolCalls != nil {
			useToolName := ""
			useToolArgs := ""
			if message.ToolCalls != nil {
				useToolName = message.ToolCalls[0].Function.Name
				useToolArgs = message.ToolCalls[0].Function.Arguments
			} else {
				useToolName = message.FunctionCall.Name
				useToolArgs = message.FunctionCall.Arguments
			}
			messages = append(messages, XunfeiMessage{
				Role:    message.Role,
				Content: fmt.Sprintf("使用工具：%s，参数：%s", useToolName, useToolArgs),
			})
		} else if message.Role == types.ChatMessageRoleFunction || message.Role == types.ChatMessageRoleTool {
			messages = append(messages, XunfeiMessage{
				Role:    types.ChatMessageRoleUser,
				Content: "这是函数调用返回的内容，请回答之前的问题：\n" + message.StringContent(),
			})
		} else {
			messages = append(messages, XunfeiMessage{
				Role:    message.Role,
				Content: message.StringContent(),
			})
		}
	}

	xunfeiRequest := XunfeiChatRequest{}

	if request.Tools != nil {
		functions := make([]*types.ChatCompletionFunction, 0, len(request.Tools))
		for _, tool := range request.Tools {
			functions = append(functions, &tool.Function)
		}
		xunfeiRequest.Payload.Functions = &XunfeiChatPayloadFunctions{}
		xunfeiRequest.Payload.Functions.Text = functions
	} else if request.Functions != nil {
		xunfeiRequest.Payload.Functions = &XunfeiChatPayloadFunctions{}
		xunfeiRequest.Payload.Functions.Text = request.Functions
	}

	xunfeiRequest.Header.AppId = p.apiId
	xunfeiRequest.Parameter.Chat.Domain = p.domain
	xunfeiRequest.Parameter.Chat.Temperature = request.Temperature
	xunfeiRequest.Parameter.Chat.TopK = request.N
	xunfeiRequest.Parameter.Chat.MaxTokens = request.MaxTokens
	xunfeiRequest.Payload.Message.Text = messages
	return &xunfeiRequest
}

func (h *xunfeiHandler) convertToChatOpenai(stream requester.StreamReaderInterface[XunfeiChatResponse]) (*types.ChatCompletionResponse, *types.OpenAIErrorWithStatusCode) {
	var content string
	var xunfeiResponse XunfeiChatResponse
	dataChan, errChan := stream.Recv()

	stop := false
	for !stop {
		select {
		case response := <-dataChan:
			if len(response.Payload.Choices.Text) == 0 {
				continue
			}
			xunfeiResponse = response
			content += xunfeiResponse.Payload.Choices.Text[0].Content
		case err := <-errChan:
			if err != nil && !errors.Is(err, io.EOF) {
				return nil, common.ErrorWrapper(err, "xunfei_failed", http.StatusInternalServerError)
			}

			if errors.Is(err, io.EOF) {
				stop = true
			}
		}
	}

	if len(xunfeiResponse.Payload.Choices.Text) == 0 {
		xunfeiResponse.Payload.Choices.Text = []XunfeiChatResponseTextItem{{}}
	}
	xunfeiResponse.Payload.Choices.Text[0].Content = content

	choice := types.ChatCompletionChoice{
		Index:        0,
		FinishReason: types.FinishReasonStop,
	}

	xunfeiText := xunfeiResponse.Payload.Choices.Text[0]

	if xunfeiText.FunctionCall != nil {
		choice.Message = types.ChatCompletionMessage{
			Role: "assistant",
		}

		if h.Request.Tools != nil {
			choice.Message.ToolCalls = []*types.ChatCompletionToolCalls{
				{
					Id:       xunfeiResponse.Header.Sid,
					Type:     "function",
					Function: xunfeiText.FunctionCall,
				},
			}
			choice.FinishReason = types.FinishReasonToolCalls
		} else {
			choice.Message.FunctionCall = xunfeiText.FunctionCall
			choice.FinishReason = types.FinishReasonFunctionCall
		}

	} else {
		choice.Message = types.ChatCompletionMessage{
			Role:    "assistant",
			Content: xunfeiText.Content,
		}
	}

	fullTextResponse := &types.ChatCompletionResponse{
		ID:      xunfeiResponse.Header.Sid,
		Object:  "chat.completion",
		Model:   h.Request.Model,
		Created: utils.GetTimestamp(),
		Choices: []types.ChatCompletionChoice{choice},
		Usage:   &xunfeiResponse.Payload.Usage.Text,
	}

	return fullTextResponse, nil
}

func (h *xunfeiHandler) handlerData(rawLine *[]byte, isFinished *bool) (*XunfeiChatResponse, error) {
	// 如果rawLine 前缀不为{，则直接返回
	if !strings.HasPrefix(string(*rawLine), "{") {
		*rawLine = nil
		return nil, nil
	}

	var xunfeiChatResponse XunfeiChatResponse
	err := json.Unmarshal(*rawLine, &xunfeiChatResponse)
	if err != nil {
		return nil, common.ErrorToOpenAIError(err)
	}

	aiError := errorHandle(&xunfeiChatResponse)
	if aiError != nil {
		return nil, aiError
	}

	if xunfeiChatResponse.Payload.Choices.Status == 2 {
		*isFinished = true
	}

	h.Usage.PromptTokens = xunfeiChatResponse.Payload.Usage.Text.PromptTokens
	h.Usage.CompletionTokens = xunfeiChatResponse.Payload.Usage.Text.CompletionTokens
	h.Usage.TotalTokens = xunfeiChatResponse.Payload.Usage.Text.TotalTokens

	return &xunfeiChatResponse, nil
}

func (h *xunfeiHandler) handlerNotStream(rawLine *[]byte, dataChan chan XunfeiChatResponse, errChan chan error) {
	isFinished := false
	xunfeiChatResponse, err := h.handlerData(rawLine, &isFinished)
	if err != nil {
		errChan <- err
		return
	}

	if *rawLine == nil {
		return
	}

	dataChan <- *xunfeiChatResponse

	if isFinished {
		errChan <- io.EOF
		*rawLine = requester.StreamClosed
	}
}

func (h *xunfeiHandler) handlerStream(rawLine *[]byte, dataChan chan string, errChan chan error) {
	isFinished := false
	xunfeiChatResponse, err := h.handlerData(rawLine, &isFinished)
	if err != nil {
		errChan <- err
		return
	}

	if *rawLine == nil {
		return
	}

	h.convertToOpenaiStream(xunfeiChatResponse, dataChan)

	if isFinished {
		errChan <- io.EOF
		*rawLine = requester.StreamClosed
	}
}

func (h *xunfeiHandler) convertToOpenaiStream(xunfeiChatResponse *XunfeiChatResponse, dataChan chan string) {
	if len(xunfeiChatResponse.Payload.Choices.Text) == 0 {
		xunfeiChatResponse.Payload.Choices.Text = []XunfeiChatResponseTextItem{{}}
	}

	choice := types.ChatCompletionStreamChoice{
		Index: 0,
		Delta: types.ChatCompletionStreamChoiceDelta{
			Role: types.ChatMessageRoleAssistant,
		},
	}
	xunfeiText := xunfeiChatResponse.Payload.Choices.Text[0]

	if xunfeiText.FunctionCall != nil {
		if h.Request.Tools != nil {
			choice.Delta.ToolCalls = []*types.ChatCompletionToolCalls{
				{
					Id:       xunfeiChatResponse.Header.Sid,
					Index:    0,
					Type:     "function",
					Function: xunfeiText.FunctionCall,
				},
			}
			choice.FinishReason = types.FinishReasonToolCalls
		} else {
			choice.Delta.FunctionCall = xunfeiText.FunctionCall
			choice.FinishReason = types.FinishReasonFunctionCall
		}

	} else {
		choice.Delta.Content = xunfeiChatResponse.Payload.Choices.Text[0].Content
		if xunfeiChatResponse.Payload.Choices.Status == 2 {
			choice.FinishReason = types.FinishReasonStop
		}
	}

	chatCompletion := types.ChatCompletionStreamResponse{
		ID:      xunfeiChatResponse.Header.Sid,
		Object:  "chat.completion.chunk",
		Created: utils.GetTimestamp(),
		Model:   h.Request.Model,
	}

	if xunfeiText.FunctionCall == nil {
		chatCompletion.Choices = []types.ChatCompletionStreamChoice{choice}
		responseBody, _ := json.Marshal(chatCompletion)
		dataChan <- string(responseBody)
	} else {
		choices := choice.ConvertOpenaiStream()
		for _, choice := range choices {
			chatCompletionCopy := chatCompletion
			chatCompletionCopy.Choices = []types.ChatCompletionStreamChoice{choice}
			responseBody, _ := json.Marshal(chatCompletionCopy)
			dataChan <- string(responseBody)
		}
	}
}
