package cloudflare

import (
	"bufio"
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/songquanpeng/one-api/common/ctxkey"
	"github.com/songquanpeng/one-api/common/render"

	"github.com/gin-gonic/gin"
	"github.com/songquanpeng/one-api/common"
	"github.com/songquanpeng/one-api/common/helper"
	"github.com/songquanpeng/one-api/common/logger"
	"github.com/songquanpeng/one-api/relay/adaptor/openai"
	"github.com/songquanpeng/one-api/relay/model"
)

func ConvertCompletionsRequest(textRequest model.GeneralOpenAIRequest) *Request {
	p, _ := textRequest.Prompt.(string)
	return &Request{
		Prompt:      p,
		MaxTokens:   textRequest.MaxTokens,
		Stream:      textRequest.Stream,
		Temperature: textRequest.Temperature,
	}
}

func StreamHandler(c *gin.Context, resp *http.Response, promptTokens int, modelName string) (*model.ErrorWithStatusCode, *model.Usage) {
	scanner := bufio.NewScanner(resp.Body)
	scanner.Split(bufio.ScanLines)

	common.SetEventStreamHeaders(c)
	id := helper.GetResponseID(c)
	responseModel := c.GetString(ctxkey.OriginalModel)
	var responseText string

	for scanner.Scan() {
		data := scanner.Text()
		if len(data) < len("data: ") {
			continue
		}
		data = strings.TrimPrefix(data, "data: ")
		data = strings.TrimSuffix(data, "\r")

		if data == "[DONE]" {
			break
		}

		var response openai.ChatCompletionsStreamResponse
		err := json.Unmarshal([]byte(data), &response)
		if err != nil {
			logger.SysError("error unmarshalling stream response: " + err.Error())
			continue
		}
		for _, v := range response.Choices {
			v.Delta.Role = "assistant"
			responseText += v.Delta.StringContent()
		}
		response.Id = id
		response.Model = modelName
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
	var response openai.TextResponse
	err = json.Unmarshal(responseBody, &response)
	if err != nil {
		return openai.ErrorWrapper(err, "unmarshal_response_body_failed", http.StatusInternalServerError), nil
	}
	response.Model = modelName
	var responseText string
	for _, v := range response.Choices {
		responseText += v.Message.Content.(string)
	}
	usage := openai.ResponseText2Usage(responseText, modelName, promptTokens)
	response.Usage = *usage
	response.Id = helper.GetResponseID(c)
	jsonResponse, err := json.Marshal(response)
	if err != nil {
		return openai.ErrorWrapper(err, "marshal_response_body_failed", http.StatusInternalServerError), nil
	}
	c.Writer.Header().Set("Content-Type", "application/json")
	c.Writer.WriteHeader(resp.StatusCode)
	_, _ = c.Writer.Write(jsonResponse)
	return nil, usage
}
