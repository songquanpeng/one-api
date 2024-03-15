package lark

import (
	"bufio"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/songquanpeng/one-api/common"
	"github.com/songquanpeng/one-api/common/helper"
	"github.com/songquanpeng/one-api/common/logger"
	"github.com/songquanpeng/one-api/relay/channel/openai"
	"github.com/songquanpeng/one-api/relay/model"
	"io"
	"net/http"
	"strings"
)

// https://help.aliyun.com/document_detail/613695.html?spm=a2c4g.2399480.0.0.1adb778fAdzP9w#341800c0f8w0r

const EnableSearchModelSuffix = "-internet"

func ConvertRequest(request model.GeneralOpenAIRequest) *ChatRequest {
	messages := make([]Message, 0, len(request.Messages))
	for i := 0; i < len(request.Messages); i++ {
		message := request.Messages[i]
		messages = append(messages, Message{
			Content: message.StringContent(),
			Role:    strings.ToLower(message.Role),
		})
	}
	enableSearch := false
	larkModel := request.Model
	if strings.HasSuffix(larkModel, EnableSearchModelSuffix) {
		enableSearch = true
		larkModel = strings.TrimSuffix(larkModel, EnableSearchModelSuffix)
	}
	if request.TopP >= 1 {
		request.TopP = 0.9999
	}
	return &ChatRequest{
		Model:    Model{Name: larkModel},
		Messages: messages,
		Parameters: Parameters{
			EnableSearch:      enableSearch,
			IncrementalOutput: request.Stream,
			Seed:              uint64(request.Seed),
			MaxTokens:         request.MaxTokens,
			Temperature:       request.Temperature,
			TopP:              request.TopP,
		},
	}
}

func ConvertEmbeddingRequest(request model.GeneralOpenAIRequest) *EmbeddingRequest {
	return &EmbeddingRequest{
		Model: "text-embedding-v1",
		Input: struct {
			Texts []string `json:"texts"`
		}{
			Texts: request.ParseInput(),
		},
	}
}

func EmbeddingHandler(c *gin.Context, resp *http.Response) (*model.ErrorWithStatusCode, *model.Usage) {
	var larkResponse EmbeddingResponse
	err := json.NewDecoder(resp.Body).Decode(&larkResponse)
	if err != nil {
		return openai.ErrorWrapper(err, "unmarshal_response_body_failed", http.StatusInternalServerError), nil
	}

	err = resp.Body.Close()
	if err != nil {
		return openai.ErrorWrapper(err, "close_response_body_failed", http.StatusInternalServerError), nil
	}

	if larkResponse.Code != "" {
		return &model.ErrorWithStatusCode{
			Error: model.Error{
				Message: larkResponse.Message,
				Type:    larkResponse.Code,
				Param:   larkResponse.RequestId,
				Code:    larkResponse.Code,
			},
			StatusCode: resp.StatusCode,
		}, nil
	}

	fullTextResponse := embeddingResponseAli2OpenAI(&larkResponse)
	jsonResponse, err := json.Marshal(fullTextResponse)
	if err != nil {
		return openai.ErrorWrapper(err, "marshal_response_body_failed", http.StatusInternalServerError), nil
	}
	c.Writer.Header().Set("Content-Type", "application/json")
	c.Writer.WriteHeader(resp.StatusCode)
	_, err = c.Writer.Write(jsonResponse)
	return nil, &fullTextResponse.Usage
}

func embeddingResponseAli2OpenAI(response *EmbeddingResponse) *openai.EmbeddingResponse {
	openAIEmbeddingResponse := openai.EmbeddingResponse{
		Object: "list",
		Data:   make([]openai.EmbeddingResponseItem, 0, len(response.Output.Embeddings)),
		Model:  "text-embedding-v1",
		Usage:  model.Usage{TotalTokens: response.Usage.TotalTokens},
	}

	for _, item := range response.Output.Embeddings {
		openAIEmbeddingResponse.Data = append(openAIEmbeddingResponse.Data, openai.EmbeddingResponseItem{
			Object:    `embedding`,
			Index:     item.TextIndex,
			Embedding: item.Embedding,
		})
	}
	return &openAIEmbeddingResponse
}

func responseAli2OpenAI(response *ChatResponse) *openai.TextResponse {
	choice := openai.TextResponseChoice{
		Index: 0,
		Message: model.Message{
			Role:    "assistant",
			Content: response.Choice.Message.Content,
		},
		FinishReason: response.Choice.FinishReason,
	}
	fullTextResponse := openai.TextResponse{
		Id:      response.RequestId,
		Object:  "chat.completion",
		Created: helper.GetTimestamp(),
		Choices: []openai.TextResponseChoice{choice},
		Usage: model.Usage{
			PromptTokens:     response.Usage.PromptTokens,
			CompletionTokens: response.Usage.CompletionTokens,
			TotalTokens:      response.Usage.TotalTokens,
		},
	}
	return &fullTextResponse
}

func streamResponseLark2OpenAI(larkResponse *ChatResponse) *openai.ChatCompletionsStreamResponse {
	var choice openai.ChatCompletionsStreamResponseChoice
	choice.Delta.Content = larkResponse.Choice.Message.Content
	if larkResponse.Choice.FinishReason != "null" {
		finishReason := larkResponse.Choice.FinishReason
		choice.FinishReason = &finishReason
	}
	response := openai.ChatCompletionsStreamResponse{
		Id:      larkResponse.RequestId,
		Object:  "chat.completion.chunk",
		Created: helper.GetTimestamp(),
		Model:   "skylark2",
		Choices: []openai.ChatCompletionsStreamResponseChoice{choice},
	}
	return &response
}

func StreamHandler(c *gin.Context, resp *http.Response) (*model.ErrorWithStatusCode, *model.Usage) {
	var usage model.Usage
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
	common.SetEventStreamHeaders(c)
	//lastResponseText := ""
	c.Stream(func(w io.Writer) bool {
		select {
		case data := <-dataChan:
			var larkResponse ChatResponse
			err := json.Unmarshal([]byte(data), &larkResponse)
			if err != nil {
				logger.SysError("error unmarshalling stream response: " + err.Error())
				return true
			}
			if larkResponse.Usage.CompletionTokens != 0 {
				usage.PromptTokens = larkResponse.Usage.TotalTokens
				usage.CompletionTokens = larkResponse.Usage.CompletionTokens
				usage.TotalTokens = larkResponse.Usage.TotalTokens
			}
			response := streamResponseLark2OpenAI(&larkResponse)
			//response.Choices[0].Delta.Content = strings.TrimPrefix(response.Choices[0].Delta.Content, lastResponseText)
			//lastResponseText = larkResponse.Output.Text
			jsonResponse, err := json.Marshal(response)
			if err != nil {
				logger.SysError("error marshalling stream response: " + err.Error())
				return true
			}
			c.Render(-1, common.CustomEvent{Data: "data: " + string(jsonResponse)})
			return true
		case <-stopChan:
			c.Render(-1, common.CustomEvent{Data: "data: [DONE]"})
			return false
		}
	})
	err := resp.Body.Close()
	if err != nil {
		return openai.ErrorWrapper(err, "close_response_body_failed", http.StatusInternalServerError), nil
	}
	return nil, &usage
}

func Handler(c *gin.Context, resp *http.Response) (*model.ErrorWithStatusCode, *model.Usage) {
	var larkResponse ChatResponse
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return openai.ErrorWrapper(err, "read_response_body_failed", http.StatusInternalServerError), nil
	}
	err = resp.Body.Close()
	if err != nil {
		return openai.ErrorWrapper(err, "close_response_body_failed", http.StatusInternalServerError), nil
	}
	err = json.Unmarshal(responseBody, &larkResponse)
	if err != nil {
		return openai.ErrorWrapper(err, "unmarshal_response_body_failed", http.StatusInternalServerError), nil
	}
	if larkResponse.Code != "" {
		return &model.ErrorWithStatusCode{
			Error: model.Error{
				Message: larkResponse.Message,
				Type:    larkResponse.Code,
				Param:   larkResponse.RequestId,
				Code:    larkResponse.Code,
			},
			StatusCode: resp.StatusCode,
		}, nil
	}
	fullTextResponse := responseAli2OpenAI(&larkResponse)
	fullTextResponse.Model = "skylark2"
	jsonResponse, err := json.Marshal(fullTextResponse)
	if err != nil {
		return openai.ErrorWrapper(err, "marshal_response_body_failed", http.StatusInternalServerError), nil
	}
	c.Writer.Header().Set("Content-Type", "application/json")
	c.Writer.WriteHeader(resp.StatusCode)
	_, err = c.Writer.Write(jsonResponse)
	return nil, &fullTextResponse.Usage
}
