package zhipu

import (
	"bufio"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	"github.com/songquanpeng/one-api/common"
	"github.com/songquanpeng/one-api/common/helper"
	"github.com/songquanpeng/one-api/common/logger"
	"github.com/songquanpeng/one-api/relay/adaptor/openai"
	"github.com/songquanpeng/one-api/relay/constant"
	"github.com/songquanpeng/one-api/relay/model"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"
)

// https://open.bigmodel.cn/doc/api#chatglm_std
// chatglm_std, chatglm_lite
// https://open.bigmodel.cn/api/paas/v3/model-api/chatglm_std/invoke
// https://open.bigmodel.cn/api/paas/v3/model-api/chatglm_std/sse-invoke

var zhipuTokens sync.Map
var expSeconds int64 = 24 * 3600

func GetToken(apikey string) string {
	data, ok := zhipuTokens.Load(apikey)
	if ok {
		tokenData := data.(tokenData)
		if time.Now().Before(tokenData.ExpiryTime) {
			return tokenData.Token
		}
	}

	split := strings.Split(apikey, ".")
	if len(split) != 2 {
		logger.SysError("invalid zhipu key: " + apikey)
		return ""
	}

	id := split[0]
	secret := split[1]

	expMillis := time.Now().Add(time.Duration(expSeconds)*time.Second).UnixNano() / 1e6
	expiryTime := time.Now().Add(time.Duration(expSeconds) * time.Second)

	timestamp := time.Now().UnixNano() / 1e6

	payload := jwt.MapClaims{
		"api_key":   id,
		"exp":       expMillis,
		"timestamp": timestamp,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, payload)

	token.Header["alg"] = "HS256"
	token.Header["sign_type"] = "SIGN"

	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		return ""
	}

	zhipuTokens.Store(apikey, tokenData{
		Token:      tokenString,
		ExpiryTime: expiryTime,
	})

	return tokenString
}

func ConvertRequest(request model.GeneralOpenAIRequest) *Request {
	messages := make([]Message, 0, len(request.Messages))
	for _, message := range request.Messages {
		messages = append(messages, Message{
			Role:    message.Role,
			Content: message.StringContent(),
		})
	}
	return &Request{
		Prompt:      messages,
		Temperature: request.Temperature,
		TopP:        request.TopP,
		Incremental: false,
	}
}

func responseZhipu2OpenAI(response *Response) *openai.TextResponse {
	fullTextResponse := openai.TextResponse{
		Id:      response.Data.TaskId,
		Object:  "chat.completion",
		Created: helper.GetTimestamp(),
		Choices: make([]openai.TextResponseChoice, 0, len(response.Data.Choices)),
		Usage:   response.Data.Usage,
	}
	for i, choice := range response.Data.Choices {
		openaiChoice := openai.TextResponseChoice{
			Index: i,
			Message: model.Message{
				Role:    choice.Role,
				Content: strings.Trim(choice.Content, "\""),
			},
			FinishReason: "",
		}
		if i == len(response.Data.Choices)-1 {
			openaiChoice.FinishReason = "stop"
		}
		fullTextResponse.Choices = append(fullTextResponse.Choices, openaiChoice)
	}
	return &fullTextResponse
}

func streamResponseZhipu2OpenAI(zhipuResponse string) *openai.ChatCompletionsStreamResponse {
	var choice openai.ChatCompletionsStreamResponseChoice
	choice.Delta.Content = zhipuResponse
	response := openai.ChatCompletionsStreamResponse{
		Object:  "chat.completion.chunk",
		Created: helper.GetTimestamp(),
		Model:   "chatglm",
		Choices: []openai.ChatCompletionsStreamResponseChoice{choice},
	}
	return &response
}

func streamMetaResponseZhipu2OpenAI(zhipuResponse *StreamMetaResponse) (*openai.ChatCompletionsStreamResponse, *model.Usage) {
	var choice openai.ChatCompletionsStreamResponseChoice
	choice.Delta.Content = ""
	choice.FinishReason = &constant.StopFinishReason
	response := openai.ChatCompletionsStreamResponse{
		Id:      zhipuResponse.RequestId,
		Object:  "chat.completion.chunk",
		Created: helper.GetTimestamp(),
		Model:   "chatglm",
		Choices: []openai.ChatCompletionsStreamResponseChoice{choice},
	}
	return &response, &zhipuResponse.Usage
}

func StreamHandler(c *gin.Context, resp *http.Response) (*model.ErrorWithStatusCode, *model.Usage) {
	var usage *model.Usage
	scanner := bufio.NewScanner(resp.Body)
	scanner.Split(func(data []byte, atEOF bool) (advance int, token []byte, err error) {
		if atEOF && len(data) == 0 {
			return 0, nil, nil
		}
		if i := strings.Index(string(data), "\n\n"); i >= 0 && strings.Index(string(data), ":") >= 0 {
			return i + 2, data[0:i], nil
		}
		if atEOF {
			return len(data), data, nil
		}
		return 0, nil, nil
	})
	dataChan := make(chan string)
	metaChan := make(chan string)
	stopChan := make(chan bool)
	go func() {
		for scanner.Scan() {
			data := scanner.Text()
			lines := strings.Split(data, "\n")
			for i, line := range lines {
				if len(line) < 5 {
					continue
				}
				if line[:5] == "data:" {
					dataChan <- line[5:]
					if i != len(lines)-1 {
						dataChan <- "\n"
					}
				} else if line[:5] == "meta:" {
					metaChan <- line[5:]
				}
			}
		}
		stopChan <- true
	}()
	common.SetEventStreamHeaders(c)
	c.Stream(func(w io.Writer) bool {
		select {
		case data := <-dataChan:
			response := streamResponseZhipu2OpenAI(data)
			jsonResponse, err := json.Marshal(response)
			if err != nil {
				logger.SysError("error marshalling stream response: " + err.Error())
				return true
			}
			c.Render(-1, common.CustomEvent{Data: "data: " + string(jsonResponse)})
			return true
		case data := <-metaChan:
			var zhipuResponse StreamMetaResponse
			err := json.Unmarshal([]byte(data), &zhipuResponse)
			if err != nil {
				logger.SysError("error unmarshalling stream response: " + err.Error())
				return true
			}
			response, zhipuUsage := streamMetaResponseZhipu2OpenAI(&zhipuResponse)
			jsonResponse, err := json.Marshal(response)
			if err != nil {
				logger.SysError("error marshalling stream response: " + err.Error())
				return true
			}
			usage = zhipuUsage
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
	return nil, usage
}

func Handler(c *gin.Context, resp *http.Response) (*model.ErrorWithStatusCode, *model.Usage) {
	var zhipuResponse Response
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return openai.ErrorWrapper(err, "read_response_body_failed", http.StatusInternalServerError), nil
	}
	err = resp.Body.Close()
	if err != nil {
		return openai.ErrorWrapper(err, "close_response_body_failed", http.StatusInternalServerError), nil
	}
	err = json.Unmarshal(responseBody, &zhipuResponse)
	if err != nil {
		return openai.ErrorWrapper(err, "unmarshal_response_body_failed", http.StatusInternalServerError), nil
	}
	if !zhipuResponse.Success {
		return &model.ErrorWithStatusCode{
			Error: model.Error{
				Message: zhipuResponse.Msg,
				Type:    "zhipu_error",
				Param:   "",
				Code:    zhipuResponse.Code,
			},
			StatusCode: resp.StatusCode,
		}, nil
	}
	fullTextResponse := responseZhipu2OpenAI(&zhipuResponse)
	fullTextResponse.Model = "chatglm"
	jsonResponse, err := json.Marshal(fullTextResponse)
	if err != nil {
		return openai.ErrorWrapper(err, "marshal_response_body_failed", http.StatusInternalServerError), nil
	}
	c.Writer.Header().Set("Content-Type", "application/json")
	c.Writer.WriteHeader(resp.StatusCode)
	_, err = c.Writer.Write(jsonResponse)
	return nil, &fullTextResponse.Usage
}

func EmbeddingsHandler(c *gin.Context, resp *http.Response) (*model.ErrorWithStatusCode, *model.Usage) {
	var zhipuResponse EmbeddingResponse
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return openai.ErrorWrapper(err, "read_response_body_failed", http.StatusInternalServerError), nil
	}
	err = resp.Body.Close()
	if err != nil {
		return openai.ErrorWrapper(err, "close_response_body_failed", http.StatusInternalServerError), nil
	}
	err = json.Unmarshal(responseBody, &zhipuResponse)
	if err != nil {
		return openai.ErrorWrapper(err, "unmarshal_response_body_failed", http.StatusInternalServerError), nil
	}
	fullTextResponse := embeddingResponseZhipu2OpenAI(&zhipuResponse)
	jsonResponse, err := json.Marshal(fullTextResponse)
	if err != nil {
		return openai.ErrorWrapper(err, "marshal_response_body_failed", http.StatusInternalServerError), nil
	}
	c.Writer.Header().Set("Content-Type", "application/json")
	c.Writer.WriteHeader(resp.StatusCode)
	_, err = c.Writer.Write(jsonResponse)
	return nil, &fullTextResponse.Usage
}

func embeddingResponseZhipu2OpenAI(response *EmbeddingResponse) *openai.EmbeddingResponse {
	openAIEmbeddingResponse := openai.EmbeddingResponse{
		Object: "list",
		Data:   make([]openai.EmbeddingResponseItem, 0, len(response.Embeddings)),
		Model:  response.Model,
		Usage: model.Usage{
			PromptTokens:     response.PromptTokens,
			CompletionTokens: response.CompletionTokens,
			TotalTokens:      response.Usage.TotalTokens,
		},
	}

	for _, item := range response.Embeddings {
		openAIEmbeddingResponse.Data = append(openAIEmbeddingResponse.Data, openai.EmbeddingResponseItem{
			Object:    `embedding`,
			Index:     item.Index,
			Embedding: item.Embedding,
		})
	}
	return &openAIEmbeddingResponse
}
