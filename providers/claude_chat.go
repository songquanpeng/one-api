package providers

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"one-api/common"
	"one-api/types"
	"strings"
)

type ClaudeMetadata struct {
	UserId string `json:"user_id"`
}

type ClaudeRequest struct {
	Model             string   `json:"model"`
	Prompt            string   `json:"prompt"`
	MaxTokensToSample int      `json:"max_tokens_to_sample"`
	StopSequences     []string `json:"stop_sequences,omitempty"`
	Temperature       float64  `json:"temperature,omitempty"`
	TopP              float64  `json:"top_p,omitempty"`
	TopK              int      `json:"top_k,omitempty"`
	//ClaudeMetadata    `json:"metadata,omitempty"`
	Stream bool `json:"stream,omitempty"`
}

type ClaudeResponse struct {
	Completion string       `json:"completion"`
	StopReason string       `json:"stop_reason"`
	Model      string       `json:"model"`
	Error      ClaudeError  `json:"error"`
	Usage      *types.Usage `json:"usage,omitempty"`
}

func (claudeResponse *ClaudeResponse) requestHandler(resp *http.Response) (OpenAIResponse any, openAIErrorWithStatusCode *types.OpenAIErrorWithStatusCode) {
	if claudeResponse.Error.Type != "" {
		return nil, &types.OpenAIErrorWithStatusCode{
			OpenAIError: types.OpenAIError{
				Message: claudeResponse.Error.Message,
				Type:    claudeResponse.Error.Type,
				Param:   "",
				Code:    claudeResponse.Error.Type,
			},
			StatusCode: resp.StatusCode,
		}
	}

	choice := types.ChatCompletionChoice{
		Index: 0,
		Message: types.ChatCompletionMessage{
			Role:    "assistant",
			Content: strings.TrimPrefix(claudeResponse.Completion, " "),
			Name:    nil,
		},
		FinishReason: stopReasonClaude2OpenAI(claudeResponse.StopReason),
	}
	fullTextResponse := types.ChatCompletionResponse{
		ID:      fmt.Sprintf("chatcmpl-%s", common.GetUUID()),
		Object:  "chat.completion",
		Created: common.GetTimestamp(),
		Choices: []types.ChatCompletionChoice{choice},
	}

	completionTokens := common.CountTokenText(claudeResponse.Completion, claudeResponse.Model)
	claudeResponse.Usage.CompletionTokens = completionTokens
	claudeResponse.Usage.TotalTokens = claudeResponse.Usage.PromptTokens + completionTokens

	fullTextResponse.Usage = claudeResponse.Usage

	return fullTextResponse, nil
}

func (p *ClaudeProvider) getChatRequestBody(request *types.ChatCompletionRequest) (requestBody *ClaudeRequest) {
	claudeRequest := ClaudeRequest{
		Model:             request.Model,
		Prompt:            "",
		MaxTokensToSample: request.MaxTokens,
		StopSequences:     nil,
		Temperature:       request.Temperature,
		TopP:              request.TopP,
		Stream:            request.Stream,
	}
	if claudeRequest.MaxTokensToSample == 0 {
		claudeRequest.MaxTokensToSample = 1000000
	}
	prompt := ""
	for _, message := range request.Messages {
		if message.Role == "user" {
			prompt += fmt.Sprintf("\n\nHuman: %s", message.Content)
		} else if message.Role == "assistant" {
			prompt += fmt.Sprintf("\n\nAssistant: %s", message.Content)
		} else if message.Role == "system" {
			prompt += fmt.Sprintf("\n\nSystem: %s", message.Content)
		}
	}
	prompt += "\n\nAssistant:"
	claudeRequest.Prompt = prompt
	return &claudeRequest
}

func (p *ClaudeProvider) ChatCompleteResponse(request *types.ChatCompletionRequest, isModelMapped bool, promptTokens int) (usage *types.Usage, openAIErrorWithStatusCode *types.OpenAIErrorWithStatusCode) {
	requestBody := p.getChatRequestBody(request)
	fullRequestURL := p.GetFullRequestURL(p.ChatCompletions, request.Model)
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
		var responseText string
		openAIErrorWithStatusCode, responseText = p.sendStreamRequest(req)
		if openAIErrorWithStatusCode != nil {
			return
		}

		usage.PromptTokens = promptTokens
		usage.CompletionTokens = common.CountTokenText(responseText, request.Model)
		usage.TotalTokens = promptTokens + usage.CompletionTokens

	} else {
		var claudeResponse = &ClaudeResponse{
			Usage: &types.Usage{
				PromptTokens: promptTokens,
			},
		}
		openAIErrorWithStatusCode = p.sendRequest(req, claudeResponse)
		if openAIErrorWithStatusCode != nil {
			return
		}

		usage = claudeResponse.Usage
	}
	return

}

func (p *ClaudeProvider) streamResponseClaude2OpenAI(claudeResponse *ClaudeResponse) *types.ChatCompletionStreamResponse {
	var choice types.ChatCompletionStreamChoice
	choice.Delta.Content = claudeResponse.Completion
	finishReason := stopReasonClaude2OpenAI(claudeResponse.StopReason)
	if finishReason != "null" {
		choice.FinishReason = &finishReason
	}
	var response types.ChatCompletionStreamResponse
	response.Object = "chat.completion.chunk"
	response.Model = claudeResponse.Model
	response.Choices = []types.ChatCompletionStreamChoice{choice}
	return &response
}

func (p *ClaudeProvider) sendStreamRequest(req *http.Request) (*types.OpenAIErrorWithStatusCode, string) {
	// 发送请求
	resp, err := common.HttpClient.Do(req)
	if err != nil {
		return types.ErrorWrapper(err, "http_request_failed", http.StatusInternalServerError), ""
	}

	if common.IsFailureStatusCode(resp) {
		return p.handleErrorResp(resp), ""
	}

	defer resp.Body.Close()

	responseText := ""
	responseId := fmt.Sprintf("chatcmpl-%s", common.GetUUID())
	createdTime := common.GetTimestamp()
	scanner := bufio.NewScanner(resp.Body)
	scanner.Split(func(data []byte, atEOF bool) (advance int, token []byte, err error) {
		if atEOF && len(data) == 0 {
			return 0, nil, nil
		}
		if i := strings.Index(string(data), "\r\n\r\n"); i >= 0 {
			return i + 4, data[0:i], nil
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
			if !strings.HasPrefix(data, "event: completion") {
				continue
			}
			data = strings.TrimPrefix(data, "event: completion\r\ndata: ")
			dataChan <- data
		}
		stopChan <- true
	}()
	setEventStreamHeaders(p.Context)
	p.Context.Stream(func(w io.Writer) bool {
		select {
		case data := <-dataChan:
			// some implementations may add \r at the end of data
			data = strings.TrimSuffix(data, "\r")
			var claudeResponse ClaudeResponse
			err := json.Unmarshal([]byte(data), &claudeResponse)
			if err != nil {
				common.SysError("error unmarshalling stream response: " + err.Error())
				return true
			}
			responseText += claudeResponse.Completion
			response := p.streamResponseClaude2OpenAI(&claudeResponse)
			response.ID = responseId
			response.Created = createdTime
			jsonStr, err := json.Marshal(response)
			if err != nil {
				common.SysError("error marshalling stream response: " + err.Error())
				return true
			}
			p.Context.Render(-1, common.CustomEvent{Data: "data: " + string(jsonStr)})
			return true
		case <-stopChan:
			p.Context.Render(-1, common.CustomEvent{Data: "data: [DONE]"})
			return false
		}
	})

	return nil, responseText
}
