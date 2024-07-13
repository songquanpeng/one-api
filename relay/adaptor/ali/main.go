package ali

import (
	"bufio"
	"encoding/json"
	"github.com/songquanpeng/one-api/common/render"
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/songquanpeng/one-api/common"
	"github.com/songquanpeng/one-api/common/helper"
	"github.com/songquanpeng/one-api/common/logger"
	"github.com/songquanpeng/one-api/relay/adaptor/openai"
	"github.com/songquanpeng/one-api/relay/model"
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
	aliModel := request.Model
	if strings.HasSuffix(aliModel, EnableSearchModelSuffix) {
		enableSearch = true
		aliModel = strings.TrimSuffix(aliModel, EnableSearchModelSuffix)
	}
	if request.TopP >= 1 {
		request.TopP = 0.9999
	}
	return &ChatRequest{
		Model: aliModel,
		Input: Input{
			Messages: messages,
		},
		Parameters: Parameters{
			EnableSearch:      enableSearch,
			IncrementalOutput: request.Stream,
			Seed:              uint64(request.Seed),
			MaxTokens:         request.MaxTokens,
			Temperature:       request.Temperature,
			TopP:              request.TopP,
			TopK:              request.TopK,
			ResultFormat:      "message",
			Tools:             request.Tools,
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

func ConvertImageRequest(request model.ImageRequest) *ImageRequest {
	var imageRequest ImageRequest
	imageRequest.Input.Prompt = request.Prompt
	imageRequest.Model = request.Model
	imageRequest.Parameters.Size = strings.Replace(request.Size, "x", "*", -1)
	imageRequest.Parameters.N = request.N
	imageRequest.ResponseFormat = request.ResponseFormat

	return &imageRequest
}

func EmbeddingHandler(c *gin.Context, resp *http.Response) (*model.ErrorWithStatusCode, *model.Usage) {
	var aliResponse EmbeddingResponse
	err := json.NewDecoder(resp.Body).Decode(&aliResponse)
	if err != nil {
		return openai.ErrorWrapper(err, "unmarshal_response_body_failed", http.StatusInternalServerError), nil
	}

	err = resp.Body.Close()
	if err != nil {
		return openai.ErrorWrapper(err, "close_response_body_failed", http.StatusInternalServerError), nil
	}

	if aliResponse.Code != "" {
		return &model.ErrorWithStatusCode{
			Error: model.Error{
				Message: aliResponse.Message,
				Type:    aliResponse.Code,
				Param:   aliResponse.RequestId,
				Code:    aliResponse.Code,
			},
			StatusCode: resp.StatusCode,
		}, nil
	}

	fullTextResponse := embeddingResponseAli2OpenAI(&aliResponse)
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
	fullTextResponse := openai.TextResponse{
		Id:      response.RequestId,
		Object:  "chat.completion",
		Created: helper.GetTimestamp(),
		Choices: response.Output.Choices,
		Usage: model.Usage{
			PromptTokens:     response.Usage.InputTokens,
			CompletionTokens: response.Usage.OutputTokens,
			TotalTokens:      response.Usage.InputTokens + response.Usage.OutputTokens,
		},
	}
	return &fullTextResponse
}

func streamResponseAli2OpenAI(aliResponse *ChatResponse) *openai.ChatCompletionsStreamResponse {
	if len(aliResponse.Output.Choices) == 0 {
		return nil
	}
	aliChoice := aliResponse.Output.Choices[0]
	var choice openai.ChatCompletionsStreamResponseChoice
	choice.Delta = aliChoice.Message
	if aliChoice.FinishReason != "null" {
		finishReason := aliChoice.FinishReason
		choice.FinishReason = &finishReason
	}
	response := openai.ChatCompletionsStreamResponse{
		Id:      aliResponse.RequestId,
		Object:  "chat.completion.chunk",
		Created: helper.GetTimestamp(),
		Model:   "qwen",
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

	common.SetEventStreamHeaders(c)

	for scanner.Scan() {
		data := scanner.Text()
		if len(data) < 5 || data[:5] != "data:" {
			continue
		}
		data = data[5:]

		var aliResponse ChatResponse
		err := json.Unmarshal([]byte(data), &aliResponse)
		if err != nil {
			logger.SysError("error unmarshalling stream response: " + err.Error())
			continue
		}
		if aliResponse.Usage.OutputTokens != 0 {
			usage.PromptTokens = aliResponse.Usage.InputTokens
			usage.CompletionTokens = aliResponse.Usage.OutputTokens
			usage.TotalTokens = aliResponse.Usage.InputTokens + aliResponse.Usage.OutputTokens
		}
		response := streamResponseAli2OpenAI(&aliResponse)
		if response == nil {
			continue
		}
		err = render.ObjectData(c, response)
		if err != nil {
			logger.SysError(err.Error())
		}
	}

	if err := scanner.Err(); err != nil {
		logger.SysError("error reading stream: " + err.Error())
	}

	render.Done(c)

	err := resp.Body.Close()
	if err != nil {
		return openai.ErrorWrapper(err, "close_response_body_failed", http.StatusInternalServerError), nil
	}
	return nil, &usage
}

func Handler(c *gin.Context, resp *http.Response) (*model.ErrorWithStatusCode, *model.Usage) {
	ctx := c.Request.Context()
	var aliResponse ChatResponse
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return openai.ErrorWrapper(err, "read_response_body_failed", http.StatusInternalServerError), nil
	}
	err = resp.Body.Close()
	if err != nil {
		return openai.ErrorWrapper(err, "close_response_body_failed", http.StatusInternalServerError), nil
	}
	logger.Debugf(ctx, "response body: %s\n", responseBody)
	err = json.Unmarshal(responseBody, &aliResponse)
	if err != nil {
		return openai.ErrorWrapper(err, "unmarshal_response_body_failed", http.StatusInternalServerError), nil
	}
	if aliResponse.Code != "" {
		return &model.ErrorWithStatusCode{
			Error: model.Error{
				Message: aliResponse.Message,
				Type:    aliResponse.Code,
				Param:   aliResponse.RequestId,
				Code:    aliResponse.Code,
			},
			StatusCode: resp.StatusCode,
		}, nil
	}
	fullTextResponse := responseAli2OpenAI(&aliResponse)
	fullTextResponse.Model = "qwen"
	jsonResponse, err := json.Marshal(fullTextResponse)
	if err != nil {
		return openai.ErrorWrapper(err, "marshal_response_body_failed", http.StatusInternalServerError), nil
	}
	c.Writer.Header().Set("Content-Type", "application/json")
	c.Writer.WriteHeader(resp.StatusCode)
	_, err = c.Writer.Write(jsonResponse)
	return nil, &fullTextResponse.Usage
}
