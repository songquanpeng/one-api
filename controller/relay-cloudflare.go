package controller

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"one-api/common"
	"strings"

	"github.com/gin-gonic/gin"
)

type CloudflareRequest struct {
	Messages  []Message `json:"messages"`
	Stream    bool      `json:"stream,omitempty"`
	MaxTokens int       `json:"max_tokens"`
	Prompt    any       `json:"prompt,omitempty"`
}

type CloudflareError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type CloudflareResult struct {
	Response string `json:"response"`
}
type CloudflareResponse struct {
	Result   CloudflareResult  `json:"result"`
	Success  bool              `json:"success"`
	Errors   []CloudflareError `json:"errors"`
	Messages []string          `json:"messages"`
}

type CloudflareStreamResponse struct {
	Reponse string `json:"response"`
}

func requestOpenAI2Cloudflare(textRequest GeneralOpenAIRequest) *CloudflareRequest {
	cloudflareRequest := CloudflareRequest{
		Messages:  textRequest.Messages,
		Stream:    textRequest.Stream,
		MaxTokens: -1,
		Prompt:    textRequest.Prompt,
	}
	return &cloudflareRequest
}

func streamResponseCloudflare2OpenAI(cloudflareStreamResponse *CloudflareStreamResponse) *ChatCompletionsStreamResponse {
	var choice ChatCompletionsStreamResponseChoice
	choice.Delta.Content = cloudflareStreamResponse.Reponse
	choice.FinishReason = &stopFinishReason
	var response ChatCompletionsStreamResponse

	response.Object = "chat.completion.chunk"
	response.Model = "cloudflare"
	response.Choices = []ChatCompletionsStreamResponseChoice{choice}

	return &response
}

func responseCloudflare2OpenAI(cloudflareResponse *CloudflareResponse) *OpenAITextResponse {
	choice := OpenAITextResponseChoice{
		Index: 0,
		Message: Message{
			Role:    "assistant",
			Content: strings.TrimPrefix(cloudflareResponse.Result.Response, " "),
			Name:    nil,
		},
		FinishReason: stopFinishReason,
	}
	PromptTokens := 1
	CompletionTokens := len(strings.TrimPrefix(cloudflareResponse.Result.Response, " "))
	TotalTokens := CompletionTokens + PromptTokens
	fullTextResponse := OpenAITextResponse{
		Id:      fmt.Sprintf("chatcmpl-%s", common.GetUUID()),
		Object:  "chat.completion",
		Created: common.GetTimestamp(),
		Choices: []OpenAITextResponseChoice{choice},
		Usage: Usage{
			PromptTokens:     PromptTokens,
			TotalTokens:      TotalTokens,
			CompletionTokens: CompletionTokens,
		},
	}
	return &fullTextResponse
}

func cloudflareStreamHandler(c *gin.Context, resp *http.Response) (*OpenAIErrorWithStatusCode, string) {
	responseText := ""
	responseId := fmt.Sprintf("chatcmpl-%s", common.GetUUID())
	createdTime := common.GetTimestamp()
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
			if len(data) < 6 { // ignore blank line or wrong format
				continue
			}
			if data[:6] != "data: " && data[:6] != "[DONE]" {
				continue
			}
			dataChan <- data
			data = data[6:]

			if !strings.HasPrefix(data, "[DONE]") {
				var streamResponse CloudflareStreamResponse
				err := json.Unmarshal([]byte(data), &streamResponse)
				if err != nil {
					common.SysError("error unmarshalling stream response: " + err.Error())
					continue // just ignore the error
				}
				responseText += streamResponse.Reponse
			}
		}
		stopChan <- true
	}()
	setEventStreamHeaders(c)
	c.Stream(func(w io.Writer) bool {
		select {
		case data := <-dataChan:
			if strings.HasPrefix(data, "data: [DONE]") {
				c.Render(-1, common.CustomEvent{Data: "data: [DONE]"})
				return false
			}
			data = data[6:]
			// some implementations may add \r at the end of data
			data = strings.TrimSuffix(data, "\r")
			var cloudflareStreamResponse CloudflareStreamResponse
			err := json.Unmarshal([]byte(data), &cloudflareStreamResponse)
			if err != nil {
				common.SysError("error unmarshalling stream response: " + err.Error())
				return true
			}
			responseText += cloudflareStreamResponse.Reponse
			response := streamResponseCloudflare2OpenAI(&cloudflareStreamResponse)
			response.Id = responseId
			response.Created = createdTime
			jsonStr, err := json.Marshal(response)
			if err != nil {
				common.SysError("error marshalling stream response: " + err.Error())
				return true
			}
			c.Render(-1, common.CustomEvent{Data: "data: " + string(jsonStr)})
			return true
		case <-stopChan:
			c.Render(-1, common.CustomEvent{Data: "data: [DONE]"})
			return false
		}
	})
	err := resp.Body.Close()
	if err != nil {
		return errorWrapper(err, "close_response_body_failed", http.StatusInternalServerError), ""
	}
	return nil, responseText
}

func cloudflareHandler(c *gin.Context, resp *http.Response, promptTokens int, model string) (*OpenAIErrorWithStatusCode, *Usage) {
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return errorWrapper(err, "read_response_body_failed", http.StatusInternalServerError), nil
	}
	err = resp.Body.Close()
	if err != nil {
		return errorWrapper(err, "close_response_body_failed", http.StatusInternalServerError), nil
	}
	var cloudflareResponse CloudflareResponse
	err = json.Unmarshal(responseBody, &cloudflareResponse)
	if err != nil {
		return errorWrapper(err, "unmarshal_response_body_failed", http.StatusInternalServerError), nil
	}
	if len(cloudflareResponse.Errors) > 0 {
		return &OpenAIErrorWithStatusCode{
			OpenAIError: OpenAIError{
				Message: cloudflareResponse.Errors[0].Message,
				Type:    "",
				Param:   "",
				Code:    cloudflareResponse.Errors[0].Code,
			},
			StatusCode: resp.StatusCode,
		}, nil
	}
	fullTextResponse := responseCloudflare2OpenAI(&cloudflareResponse)
	// completionTokens := 0
	completionTokens := countTokenText(cloudflareResponse.Result.Response, model)
	usage := Usage{
		PromptTokens:     promptTokens,
		CompletionTokens: completionTokens,
		TotalTokens:      promptTokens + completionTokens,
	}
	fullTextResponse.Usage = usage
	jsonResponse, err := json.Marshal(fullTextResponse)
	if err != nil {
		return errorWrapper(err, "marshal_response_body_failed", http.StatusInternalServerError), nil
	}
	c.Writer.Header().Set("Content-Type", "application/json")
	c.Writer.WriteHeader(resp.StatusCode)
	_, err = c.Writer.Write(jsonResponse)
	return nil, &usage
}

func getCloudflareAccountID(apiKey string) (string, error) {
	split := strings.Split(apiKey, "|")
	if len(split) != 2 {
		return "", errors.New("getCloudflareAccountID: Invalid API key format")
	}
	return split[0], nil
}

func getCloudflareAPI_Token(apiKey string) (string, error) {
	split := strings.Split(apiKey, "|")
	if len(split) != 2 {
		return "", errors.New("getCloudflareAPI_Token: Invalid API key format")
	}
	return split[1], nil
}
