package controller

import (
	"bufio"
	"crypto/md5"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"io"
	"net/http"
	"one-api/common"
	"strings"
	"sync"
	"time"
)

// https://cloud.baidu.com/doc/WENXINWORKSHOP/s/flfmc9do2

type BaiduTokenResponse struct {
	RefreshToken  string `json:"refresh_token"`
	ExpiresIn     int    `json:"expires_in"`
	SessionKey    string `json:"session_key"`
	AccessToken   string `json:"access_token"`
	Scope         string `json:"scope"`
	SessionSecret string `json:"session_secret"`
}

type BaiduMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type BaiduChatRequest struct {
	Messages []BaiduMessage `json:"messages"`
	Stream   bool           `json:"stream"`
	UserId   string         `json:"user_id,omitempty"`
}

type BaiduError struct {
	ErrorCode int    `json:"error_code"`
	ErrorMsg  string `json:"error_msg"`
}

type BaiduChatResponse struct {
	Id               string `json:"id"`
	Object           string `json:"object"`
	Created          int64  `json:"created"`
	Result           string `json:"result"`
	IsTruncated      bool   `json:"is_truncated"`
	NeedClearHistory bool   `json:"need_clear_history"`
	Usage            Usage  `json:"usage"`
	BaiduError
}

type BaiduChatStreamResponse struct {
	BaiduChatResponse
	SentenceId int  `json:"sentence_id"`
	IsEnd      bool `json:"is_end"`
}

type BaiduEmbeddingRequest struct {
	Input []string `json:"input"`
}

type BaiduEmbeddingData struct {
	Object    string    `json:"object"`
	Embedding []float64 `json:"embedding"`
	Index     int       `json:"index"`
}

type BaiduEmbeddingResponse struct {
	Id      string               `json:"id"`
	Object  string               `json:"object"`
	Created int64                `json:"created"`
	Data    []BaiduEmbeddingData `json:"data"`
	Usage   Usage                `json:"usage"`
	BaiduError
}

type BaiduAccessToken struct {
	RefreshToken     string    `json:"refresh_token"`
	ExpiresIn        int       `json:"expires_in"`
	SessionKey       string    `json:"session_key"`
	AccessToken      string    `json:"access_token"`
	Scope            string    `json:"scope"`
	SessionSecret    string    `json:"session_secret"`
	Error            string    `json:"error,omitempty"`
	ErrorDescription string    `json:"error_description,omitempty"`
	ExpiresAt        time.Time `json:"expires_at,omitempty"`
	SecretKey        string    `json:"secret_key,omitempty"`
	ApiKey           string    `json:"api_key,omitempty"`
}

var baiduAccessTokens sync.Map

func requestOpenAI2Baidu(request GeneralOpenAIRequest) *BaiduChatRequest {
	messages := make([]BaiduMessage, 0, len(request.Messages))
	for _, message := range request.Messages {
		if message.Role == "system" {
			messages = append(messages, BaiduMessage{
				Role:    "user",
				Content: message.Content,
			})
			messages = append(messages, BaiduMessage{
				Role:    "assistant",
				Content: "Okay",
			})
		} else {
			messages = append(messages, BaiduMessage{
				Role:    message.Role,
				Content: message.Content,
			})
		}
	}
	return &BaiduChatRequest{
		Messages: messages,
		Stream:   request.Stream,
	}
}

func responseBaidu2OpenAI(response *BaiduChatResponse) *OpenAITextResponse {
	choice := OpenAITextResponseChoice{
		Index: 0,
		Message: Message{
			Role:    "assistant",
			Content: response.Result,
		},
		FinishReason: "stop",
	}
	fullTextResponse := OpenAITextResponse{
		Id:      response.Id,
		Object:  "chat.completion",
		Created: response.Created,
		Choices: []OpenAITextResponseChoice{choice},
		Usage:   response.Usage,
	}
	return &fullTextResponse
}

func streamResponseBaidu2OpenAI(baiduResponse *BaiduChatStreamResponse) *ChatCompletionsStreamResponse {
	var choice ChatCompletionsStreamResponseChoice
	choice.Delta.Content = baiduResponse.Result
	if baiduResponse.IsEnd {
		choice.FinishReason = &stopFinishReason
	}
	response := ChatCompletionsStreamResponse{
		Id:      baiduResponse.Id,
		Object:  "chat.completion.chunk",
		Created: baiduResponse.Created,
		Model:   "ernie-bot",
		Choices: []ChatCompletionsStreamResponseChoice{choice},
	}
	return &response
}

func embeddingRequestOpenAI2Baidu(request GeneralOpenAIRequest) *BaiduEmbeddingRequest {
	baiduEmbeddingRequest := BaiduEmbeddingRequest{
		Input: nil,
	}
	switch request.Input.(type) {
	case string:
		baiduEmbeddingRequest.Input = []string{request.Input.(string)}
	case []string:
		baiduEmbeddingRequest.Input = request.Input.([]string)
	}
	return &baiduEmbeddingRequest
}

func embeddingResponseBaidu2OpenAI(response *BaiduEmbeddingResponse) *OpenAIEmbeddingResponse {
	openAIEmbeddingResponse := OpenAIEmbeddingResponse{
		Object: "list",
		Data:   make([]OpenAIEmbeddingResponseItem, 0, len(response.Data)),
		Model:  "baidu-embedding",
		Usage:  response.Usage,
	}
	for _, item := range response.Data {
		openAIEmbeddingResponse.Data = append(openAIEmbeddingResponse.Data, OpenAIEmbeddingResponseItem{
			Object:    item.Object,
			Index:     item.Index,
			Embedding: item.Embedding,
		})
	}
	return &openAIEmbeddingResponse
}

func baiduStreamHandler(c *gin.Context, resp *http.Response) (*OpenAIErrorWithStatusCode, *Usage) {
	var usage Usage
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
			data = data[6:]
			dataChan <- data
		}
		stopChan <- true
	}()
	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Header().Set("Transfer-Encoding", "chunked")
	c.Writer.Header().Set("X-Accel-Buffering", "no")
	c.Stream(func(w io.Writer) bool {
		select {
		case data := <-dataChan:
			var baiduResponse BaiduChatStreamResponse
			err := json.Unmarshal([]byte(data), &baiduResponse)
			if err != nil {
				common.SysError("error unmarshalling stream response: " + err.Error())
				return true
			}
			usage.PromptTokens += baiduResponse.Usage.PromptTokens
			usage.CompletionTokens += baiduResponse.Usage.CompletionTokens
			usage.TotalTokens += baiduResponse.Usage.TotalTokens
			response := streamResponseBaidu2OpenAI(&baiduResponse)
			jsonResponse, err := json.Marshal(response)
			if err != nil {
				common.SysError("error marshalling stream response: " + err.Error())
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
		return errorWrapper(err, "close_response_body_failed", http.StatusInternalServerError), nil
	}
	return nil, &usage
}

func baiduHandler(c *gin.Context, resp *http.Response) (*OpenAIErrorWithStatusCode, *Usage) {
	var baiduResponse BaiduChatResponse
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return errorWrapper(err, "read_response_body_failed", http.StatusInternalServerError), nil
	}
	err = resp.Body.Close()
	if err != nil {
		return errorWrapper(err, "close_response_body_failed", http.StatusInternalServerError), nil
	}
	err = json.Unmarshal(responseBody, &baiduResponse)
	if err != nil {
		return errorWrapper(err, "unmarshal_response_body_failed", http.StatusInternalServerError), nil
	}
	if baiduResponse.ErrorMsg != "" {
		return &OpenAIErrorWithStatusCode{
			OpenAIError: OpenAIError{
				Message: baiduResponse.ErrorMsg,
				Type:    "baidu_error",
				Param:   "",
				Code:    baiduResponse.ErrorCode,
			},
			StatusCode: resp.StatusCode,
		}, nil
	}
	fullTextResponse := responseBaidu2OpenAI(&baiduResponse)
	jsonResponse, err := json.Marshal(fullTextResponse)
	if err != nil {
		return errorWrapper(err, "marshal_response_body_failed", http.StatusInternalServerError), nil
	}
	c.Writer.Header().Set("Content-Type", "application/json")
	c.Writer.WriteHeader(resp.StatusCode)
	_, err = c.Writer.Write(jsonResponse)
	return nil, &fullTextResponse.Usage
}

func baiduEmbeddingHandler(c *gin.Context, resp *http.Response) (*OpenAIErrorWithStatusCode, *Usage) {
	var baiduResponse BaiduEmbeddingResponse
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return errorWrapper(err, "read_response_body_failed", http.StatusInternalServerError), nil
	}
	err = resp.Body.Close()
	if err != nil {
		return errorWrapper(err, "close_response_body_failed", http.StatusInternalServerError), nil
	}
	err = json.Unmarshal(responseBody, &baiduResponse)
	if err != nil {
		return errorWrapper(err, "unmarshal_response_body_failed", http.StatusInternalServerError), nil
	}
	if baiduResponse.ErrorMsg != "" {
		return &OpenAIErrorWithStatusCode{
			OpenAIError: OpenAIError{
				Message: baiduResponse.ErrorMsg,
				Type:    "baidu_error",
				Param:   "",
				Code:    baiduResponse.ErrorCode,
			},
			StatusCode: resp.StatusCode,
		}, nil
	}
	fullTextResponse := embeddingResponseBaidu2OpenAI(&baiduResponse)
	jsonResponse, err := json.Marshal(fullTextResponse)
	if err != nil {
		return errorWrapper(err, "marshal_response_body_failed", http.StatusInternalServerError), nil
	}
	c.Writer.Header().Set("Content-Type", "application/json")
	c.Writer.WriteHeader(resp.StatusCode)
	_, err = c.Writer.Write(jsonResponse)
	return nil, &fullTextResponse.Usage
}

func getBaiduAccessToken(apiKey string) (string, error) {
	var accessToken BaiduAccessToken
	if val, ok := baiduAccessTokens.Load(md5.Sum([]byte(apiKey))); ok {
		if accessToken, ok = val.(BaiduAccessToken); ok {
			// 提前1小时刷新
			if time.Now().Add(time.Hour).After(accessToken.ExpiresAt) {
				go refreshBaiduAccessToken(&accessToken)
			}
			return accessToken.AccessToken, nil
		}
	}

	splits := strings.Split(apiKey, "|")
	if len(splits) == 1 {
		accessToken.AccessToken = apiKey
		accessToken.ExpiresAt = time.Now().Add(30 * 24 * time.Hour)
		return apiKey, nil
	}

	var token string
	var err error
	if token, err = initBaiduAccessToken(splits[0], splits[1], ""); err != nil {
		return "", err
	}

	return token, nil
}

func refreshBaiduAccessToken(accessToken *BaiduAccessToken) error {
	if accessToken.RefreshToken == "" {
		return nil
	}
	_, err := initBaiduAccessToken(accessToken.SecretKey, accessToken.ApiKey, accessToken.RefreshToken)
	return err
}

func initBaiduAccessToken(secretKey, apiKey, refreshToken string) (string, error) {
	url := "https://aip.baidubce.com/oauth/2.0/token?grant_type=client_credentials&client_id=" + apiKey + "&client_secret=" + secretKey
	if refreshToken != "" {
		url += "&refresh_token=" + refreshToken
	}
	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return "", errors.New(fmt.Sprintf("initBaiduAccessToken err: %s", err.Error()))
	}

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept", "application/json")

	client := http.Client{
		Timeout: 5 * time.Second,
	}

	res, err := client.Do(req)
	if err != nil {
		return "", errors.New(fmt.Sprintf("initBaiduAccessToken request err: %s", err.Error()))
	}
	defer res.Body.Close()

	var accessToken BaiduAccessToken
	err = json.NewDecoder(res.Body).Decode(&accessToken)
	if err != nil {
		return "", errors.New(fmt.Sprintf("initBaiduAccessToken decode access token err: %s", err.Error()))
	}
	if accessToken.Error != "" {
		return "", errors.New(accessToken.Error + ": " + accessToken.ErrorDescription)
	}

	if accessToken.AccessToken == "" {
		return "", errors.New("initBaiduAccessToken get access token empty")
	}

	accessToken.ExpiresAt = time.Now().Add(time.Duration(accessToken.ExpiresIn) * time.Second)
	accessToken.SecretKey = secretKey
	accessToken.ApiKey = apiKey
	baiduAccessTokens.Store(md5.Sum([]byte(secretKey+"|"+apiKey)), accessToken)
	return accessToken.AccessToken, nil
}
