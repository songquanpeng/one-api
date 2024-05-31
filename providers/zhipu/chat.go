package zhipu

import (
	"encoding/json"
	"io"
	"net/http"
	"one-api/common"
	"one-api/common/config"
	"one-api/common/requester"
	"one-api/types"
	"strings"
)

type zhipuStreamHandler struct {
	Usage   *types.Usage
	Request *types.ChatCompletionRequest
}

func (p *ZhipuProvider) CreateChatCompletion(request *types.ChatCompletionRequest) (*types.ChatCompletionResponse, *types.OpenAIErrorWithStatusCode) {
	req, errWithCode := p.getChatRequest(request)
	if errWithCode != nil {
		return nil, errWithCode
	}
	defer req.Body.Close()

	zhipuChatResponse := &ZhipuResponse{}
	// 发送请求
	_, errWithCode = p.Requester.SendRequest(req, zhipuChatResponse, false)
	if errWithCode != nil {
		return nil, errWithCode
	}

	return p.convertToChatOpenai(zhipuChatResponse, request)
}

func (p *ZhipuProvider) CreateChatCompletionStream(request *types.ChatCompletionRequest) (requester.StreamReaderInterface[string], *types.OpenAIErrorWithStatusCode) {
	req, errWithCode := p.getChatRequest(request)
	if errWithCode != nil {
		return nil, errWithCode
	}
	defer req.Body.Close()

	// 发送请求
	resp, errWithCode := p.Requester.SendRequestRaw(req)
	if errWithCode != nil {
		return nil, errWithCode
	}

	chatHandler := &zhipuStreamHandler{
		Usage:   p.Usage,
		Request: request,
	}

	return requester.RequestStream[string](p.Requester, resp, chatHandler.handlerStream)
}

func (p *ZhipuProvider) getChatRequest(request *types.ChatCompletionRequest) (*http.Request, *types.OpenAIErrorWithStatusCode) {
	url, errWithCode := p.GetSupportedAPIUri(config.RelayModeChatCompletions)
	if errWithCode != nil {
		return nil, errWithCode
	}

	// 获取请求地址
	fullRequestURL := p.GetFullRequestURL(url)
	if fullRequestURL == "" {
		return nil, common.ErrorWrapper(nil, "invalid_zhipu_config", http.StatusInternalServerError)
	}

	// 获取请求头
	headers := p.GetRequestHeaders()

	zhipuRequest := p.convertFromChatOpenai(request)

	// 创建请求
	req, err := p.Requester.NewRequest(http.MethodPost, fullRequestURL, p.Requester.WithBody(zhipuRequest), p.Requester.WithHeader(headers))
	if err != nil {
		return nil, common.ErrorWrapper(err, "new_request_failed", http.StatusInternalServerError)
	}

	return req, nil
}

func (p *ZhipuProvider) convertToChatOpenai(response *ZhipuResponse, request *types.ChatCompletionRequest) (openaiResponse *types.ChatCompletionResponse, errWithCode *types.OpenAIErrorWithStatusCode) {
	aiError := errorHandle(&response.Error)
	if aiError != nil {
		errWithCode = &types.OpenAIErrorWithStatusCode{
			OpenAIError: *aiError,
			StatusCode:  http.StatusBadRequest,
		}
		return
	}

	openaiResponse = &types.ChatCompletionResponse{
		ID:      response.ID,
		Object:  "chat.completion",
		Created: response.Created,
		Model:   request.Model,
		Choices: response.Choices,
		Usage:   response.Usage,
	}

	if len(openaiResponse.Choices) > 0 && openaiResponse.Choices[0].Message.ToolCalls != nil && request.Functions != nil {
		for i := range openaiResponse.Choices {
			openaiResponse.Choices[i].CheckChoice(request)
		}
	}

	*p.Usage = *response.Usage

	return
}

func (p *ZhipuProvider) convertFromChatOpenai(request *types.ChatCompletionRequest) *ZhipuRequest {
	request.ClearEmptyMessages()
	for i := range request.Messages {
		request.Messages[i].Role = convertRole(request.Messages[i].Role)
		if request.Messages[i].FunctionCall != nil {
			request.Messages[i].FuncToToolCalls()
		}
	}

	zhipuRequest := &ZhipuRequest{
		Model:       request.Model,
		Messages:    request.Messages,
		Stream:      request.Stream,
		Temperature: request.Temperature,
		TopP:        convertTopP(request.TopP),
		MaxTokens:   request.MaxTokens,
		Stop:        request.Stop,
		ToolChoice:  request.ToolChoice,
	}

	// 如果有图片的话，并且是base64编码的图片，需要把前缀去掉
	if zhipuRequest.Model == "glm-4v" {
		for i := range zhipuRequest.Messages {
			contentList, ok := zhipuRequest.Messages[i].Content.([]any)
			if !ok {
				continue
			}
			for j := range contentList {
				contentMap, ok := contentList[j].(map[string]any)
				if !ok || contentMap["type"] != "image_url" {
					continue
				}
				imageUrl, ok := contentMap["image_url"].(map[string]any)
				if !ok {
					continue
				}
				url, ok := imageUrl["url"].(string)
				if !ok || !strings.HasPrefix(url, "data:image/") {
					continue
				}
				imageUrl["url"] = strings.Split(url, ",")[1]
				contentMap["image_url"] = imageUrl
				contentList[j] = contentMap
			}
			zhipuRequest.Messages[i].Content = contentList
		}
	}

	if request.Functions != nil {
		zhipuRequest.Tools = make([]ZhipuTool, 0, len(request.Functions))
		for _, function := range request.Functions {
			zhipuRequest.Tools = append(zhipuRequest.Tools, ZhipuTool{
				Type:     "function",
				Function: function,
			})
		}
	} else if request.Tools != nil {
		zhipuRequest.Tools = make([]ZhipuTool, 0, len(request.Tools))
		for _, tool := range request.Tools {
			zhipuRequest.Tools = append(zhipuRequest.Tools, ZhipuTool{
				Type:     "function",
				Function: &tool.Function,
			})
		}
	}

	p.pluginHandle(zhipuRequest)
	return zhipuRequest
}

func (p *ZhipuProvider) pluginHandle(request *ZhipuRequest) {
	if p.Channel.Plugin == nil {
		return
	}

	plugin := p.Channel.Plugin.Data()

	// 检测是否开启了 retrieval 插件
	if pRetrieval, ok := plugin["retrieval"]; ok {
		if knowledgeId, ok := pRetrieval["knowledge_id"].(string); ok && knowledgeId != "" {
			retrieval := ZhipuTool{
				Type: "retrieval",
				Retrieval: &ZhipuRetrieval{
					KnowledgeId: knowledgeId,
				},
			}

			if promptTemplate, ok := pRetrieval["prompt_template"].(string); ok && promptTemplate != "" {
				retrieval.Retrieval.PromptTemplate = promptTemplate
			}

			request.Tools = append(request.Tools, retrieval)

			// 如果开启了 retrieval 插件，web_search 无效
			return
		}
	}

	// 检测是否开启了 web_search 插件
	if pWeb, ok := plugin["web_search"]; ok {
		if enable, ok := pWeb["enable"].(bool); ok && enable {
			request.Tools = append(request.Tools, ZhipuTool{
				Type: "web_search",
				WebSearch: &ZhipuWebSearch{
					Enable: true,
				},
			})
		}
	}
}

// 转换为OpenAI聊天流式请求体
func (h *zhipuStreamHandler) handlerStream(rawLine *[]byte, dataChan chan string, errChan chan error) {
	// 如果rawLine 前缀不为data: 或者 meta:，则直接返回
	if !strings.HasPrefix(string(*rawLine), "data: ") {
		*rawLine = nil
		return
	}

	*rawLine = (*rawLine)[6:]

	if strings.HasPrefix(string(*rawLine), "[DONE]") {
		errChan <- io.EOF
		*rawLine = requester.StreamClosed
		return
	}

	zhipuResponse := &ZhipuStreamResponse{}
	err := json.Unmarshal(*rawLine, zhipuResponse)
	if err != nil {
		errChan <- common.ErrorToOpenAIError(err)
		return
	}

	aiError := errorHandle(&zhipuResponse.Error)
	if aiError != nil {
		errChan <- aiError
		return
	}

	h.convertToOpenaiStream(zhipuResponse, dataChan)
}

func (h *zhipuStreamHandler) convertToOpenaiStream(zhipuResponse *ZhipuStreamResponse, dataChan chan string) {
	streamResponse := types.ChatCompletionStreamResponse{
		ID:      zhipuResponse.ID,
		Object:  "chat.completion.chunk",
		Created: zhipuResponse.Created,
		Model:   h.Request.Model,
	}

	if zhipuResponse.Choices[0].Delta.ToolCalls != nil {
		choice := zhipuResponse.Choices[0]
		choice.CheckChoice(h.Request)
		choices := choice.ConvertOpenaiStream()
		for _, choice := range choices {
			chatCompletionCopy := streamResponse
			chatCompletionCopy.Choices = []types.ChatCompletionStreamChoice{choice}
			responseBody, _ := json.Marshal(chatCompletionCopy)
			dataChan <- string(responseBody)
		}
	} else {
		streamResponse.Choices = zhipuResponse.Choices
		responseBody, _ := json.Marshal(streamResponse)
		dataChan <- string(responseBody)
	}

	if zhipuResponse.Usage != nil {
		*h.Usage = *zhipuResponse.Usage
	} else {
		h.Usage.CompletionTokens += common.CountTokenText(zhipuResponse.GetResponseText(), h.Request.Model)
		h.Usage.TotalTokens = h.Usage.PromptTokens + h.Usage.CompletionTokens
	}
}
