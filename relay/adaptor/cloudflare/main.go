package cloudflare

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

func ConvertRequest(textRequest model.GeneralOpenAIRequest) *Request {
	var promptBuilder strings.Builder
	for _, message := range textRequest.Messages {
		promptBuilder.WriteString(message.StringContent())
		promptBuilder.WriteString("\n") // 添加换行符来分隔每个消息
	}

	return &Request{
		MaxTokens:   textRequest.MaxTokens,
		Prompt:      promptBuilder.String(),
		Stream:      textRequest.Stream,
		Temperature: textRequest.Temperature,
	}
}

func ResponseCloudflare2OpenAI(cloudflareResponse *Response) *openai.TextResponse {
	choice := openai.TextResponseChoice{
		Index: 0,
		Message: model.Message{
			Role:    "assistant",
			Content: cloudflareResponse.Result.Response,
		},
		FinishReason: "stop",
	}
	fullTextResponse := openai.TextResponse{
		Object:  "chat.completion",
		Created: helper.GetTimestamp(),
		Choices: []openai.TextResponseChoice{choice},
	}
	return &fullTextResponse
}

func StreamResponseCloudflare2OpenAI(cloudflareResponse *StreamResponse) *openai.ChatCompletionsStreamResponse {
	var choice openai.ChatCompletionsStreamResponseChoice
	choice.Delta.Content = cloudflareResponse.Response
	choice.Delta.Role = "assistant"
	openaiResponse := openai.ChatCompletionsStreamResponse{
		Object:  "chat.completion.chunk",
		Choices: []openai.ChatCompletionsStreamResponseChoice{choice},
		Created: helper.GetTimestamp(),
	}
	return &openaiResponse
}

func StreamHandler(c *gin.Context, resp *http.Response, promptTokens int, modelName string) (*model.ErrorWithStatusCode, *model.Usage) {
	scanner := bufio.NewScanner(resp.Body)
	scanner.Split(bufio.ScanLines)

	common.SetEventStreamHeaders(c)
	id := helper.GetResponseID(c)
	responseModel := c.GetString("original_model")
	var responseText string

	for scanner.Scan() {
		data := scanner.Text()
		if len(data) < len("data: ") {
			continue
		}
		data = strings.TrimPrefix(data, "data: ")
		data = strings.TrimSuffix(data, "\r")

		var cloudflareResponse StreamResponse
		err := json.Unmarshal([]byte(data), &cloudflareResponse)
		if err != nil {
			logger.SysError("error unmarshalling stream response: " + err.Error())
			continue
		}

		response := StreamResponseCloudflare2OpenAI(&cloudflareResponse)
		if response == nil {
			continue
		}

		responseText += cloudflareResponse.Response
		response.Id = id
		response.Model = responseModel

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

	usage := openai.ResponseText2Usage(responseText, responseModel, promptTokens)
	return nil, usage
}

func Handler(c *gin.Context, resp *http.Response, promptTokens int, modelName string) (*model.ErrorWithStatusCode, *model.Usage) {
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return openai.ErrorWrapper(err, "read_response_body_failed", http.StatusInternalServerError), nil
	}
	err = resp.Body.Close()
	if err != nil {
		return openai.ErrorWrapper(err, "close_response_body_failed", http.StatusInternalServerError), nil
	}
	var cloudflareResponse Response
	err = json.Unmarshal(responseBody, &cloudflareResponse)
	if err != nil {
		return openai.ErrorWrapper(err, "unmarshal_response_body_failed", http.StatusInternalServerError), nil
	}
	fullTextResponse := ResponseCloudflare2OpenAI(&cloudflareResponse)
	fullTextResponse.Model = modelName
	usage := openai.ResponseText2Usage(cloudflareResponse.Result.Response, modelName, promptTokens)
	fullTextResponse.Usage = *usage
	fullTextResponse.Id = helper.GetResponseID(c)
	jsonResponse, err := json.Marshal(fullTextResponse)
	if err != nil {
		return openai.ErrorWrapper(err, "marshal_response_body_failed", http.StatusInternalServerError), nil
	}
	c.Writer.Header().Set("Content-Type", "application/json")
	c.Writer.WriteHeader(resp.StatusCode)
	_, err = c.Writer.Write(jsonResponse)
	return nil, usage
}
