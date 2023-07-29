package controller

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"io"
	"net/http"
	"net/url"
	"one-api/common"
	"strings"
	"time"
)

// https://console.xfyun.cn/services/cbm
// https://www.xfyun.cn/doc/spark/Web.html

type XunfeiMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type XunfeiChatRequest struct {
	Header struct {
		AppId string `json:"app_id"`
	} `json:"header"`
	Parameter struct {
		Chat struct {
			Domain      string  `json:"domain,omitempty"`
			Temperature float64 `json:"temperature,omitempty"`
			TopK        int     `json:"top_k,omitempty"`
			MaxTokens   int     `json:"max_tokens,omitempty"`
			Auditing    bool    `json:"auditing,omitempty"`
		} `json:"chat"`
	} `json:"parameter"`
	Payload struct {
		Message struct {
			Text []XunfeiMessage `json:"text"`
		} `json:"message"`
	} `json:"payload"`
}

type XunfeiChatResponseTextItem struct {
	Content string `json:"content"`
	Role    string `json:"role"`
	Index   int    `json:"index"`
}

type XunfeiChatResponse struct {
	Header struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Sid     string `json:"sid"`
		Status  int    `json:"status"`
	} `json:"header"`
	Payload struct {
		Choices struct {
			Status int                          `json:"status"`
			Seq    int                          `json:"seq"`
			Text   []XunfeiChatResponseTextItem `json:"text"`
		} `json:"choices"`
	} `json:"payload"`
	Usage struct {
		//Text struct {
		//	QuestionTokens   string `json:"question_tokens"`
		//	PromptTokens     string `json:"prompt_tokens"`
		//	CompletionTokens string `json:"completion_tokens"`
		//	TotalTokens      string `json:"total_tokens"`
		//} `json:"text"`
		Text Usage `json:"text"`
	} `json:"usage"`
}

func requestOpenAI2Xunfei(request GeneralOpenAIRequest, xunfeiAppId string) *XunfeiChatRequest {
	messages := make([]XunfeiMessage, 0, len(request.Messages))
	for _, message := range request.Messages {
		if message.Role == "system" {
			messages = append(messages, XunfeiMessage{
				Role:    "user",
				Content: message.Content,
			})
			messages = append(messages, XunfeiMessage{
				Role:    "assistant",
				Content: "Okay",
			})
		} else {
			messages = append(messages, XunfeiMessage{
				Role:    message.Role,
				Content: message.Content,
			})
		}
	}
	xunfeiRequest := XunfeiChatRequest{}
	xunfeiRequest.Header.AppId = xunfeiAppId
	xunfeiRequest.Parameter.Chat.Domain = "general"
	xunfeiRequest.Parameter.Chat.Temperature = request.Temperature
	xunfeiRequest.Parameter.Chat.TopK = request.N
	xunfeiRequest.Parameter.Chat.MaxTokens = request.MaxTokens
	xunfeiRequest.Payload.Message.Text = messages
	return &xunfeiRequest
}

func responseXunfei2OpenAI(response *XunfeiChatResponse) *OpenAITextResponse {
	if len(response.Payload.Choices.Text) == 0 {
		response.Payload.Choices.Text = []XunfeiChatResponseTextItem{
			{
				Content: "",
			},
		}
	}
	choice := OpenAITextResponseChoice{
		Index: 0,
		Message: Message{
			Role:    "assistant",
			Content: response.Payload.Choices.Text[0].Content,
		},
	}
	fullTextResponse := OpenAITextResponse{
		Object:  "chat.completion",
		Created: common.GetTimestamp(),
		Choices: []OpenAITextResponseChoice{choice},
		Usage:   response.Usage.Text,
	}
	return &fullTextResponse
}

func streamResponseXunfei2OpenAI(xunfeiResponse *XunfeiChatResponse) *ChatCompletionsStreamResponse {
	if len(xunfeiResponse.Payload.Choices.Text) == 0 {
		xunfeiResponse.Payload.Choices.Text = []XunfeiChatResponseTextItem{
			{
				Content: "",
			},
		}
	}
	var choice ChatCompletionsStreamResponseChoice
	choice.Delta.Content = xunfeiResponse.Payload.Choices.Text[0].Content
	response := ChatCompletionsStreamResponse{
		Object:  "chat.completion.chunk",
		Created: common.GetTimestamp(),
		Model:   "SparkDesk",
		Choices: []ChatCompletionsStreamResponseChoice{choice},
	}
	return &response
}

func buildXunfeiAuthUrl(hostUrl string, apiKey, apiSecret string) string {
	HmacWithShaToBase64 := func(algorithm, data, key string) string {
		mac := hmac.New(sha256.New, []byte(key))
		mac.Write([]byte(data))
		encodeData := mac.Sum(nil)
		return base64.StdEncoding.EncodeToString(encodeData)
	}
	ul, err := url.Parse(hostUrl)
	if err != nil {
		fmt.Println(err)
	}
	date := time.Now().UTC().Format(time.RFC1123)
	signString := []string{"host: " + ul.Host, "date: " + date, "GET " + ul.Path + " HTTP/1.1"}
	sign := strings.Join(signString, "\n")
	sha := HmacWithShaToBase64("hmac-sha256", sign, apiSecret)
	authUrl := fmt.Sprintf("hmac username=\"%s\", algorithm=\"%s\", headers=\"%s\", signature=\"%s\"", apiKey,
		"hmac-sha256", "host date request-line", sha)
	authorization := base64.StdEncoding.EncodeToString([]byte(authUrl))
	v := url.Values{}
	v.Add("host", ul.Host)
	v.Add("date", date)
	v.Add("authorization", authorization)
	callUrl := hostUrl + "?" + v.Encode()
	return callUrl
}

func xunfeiStreamHandler(c *gin.Context, textRequest GeneralOpenAIRequest, appId string, apiSecret string, apiKey string) (*OpenAIErrorWithStatusCode, *Usage) {
	var usage Usage
	d := websocket.Dialer{
		HandshakeTimeout: 5 * time.Second,
	}
	hostUrl := "wss://aichat.xf-yun.com/v1/chat"
	conn, resp, err := d.Dial(buildXunfeiAuthUrl(hostUrl, apiKey, apiSecret), nil)
	if err != nil || resp.StatusCode != 101 {
		return errorWrapper(err, "dial_failed", http.StatusInternalServerError), nil
	}
	data := requestOpenAI2Xunfei(textRequest, appId)
	err = conn.WriteJSON(data)
	if err != nil {
		return errorWrapper(err, "write_json_failed", http.StatusInternalServerError), nil
	}
	dataChan := make(chan XunfeiChatResponse)
	stopChan := make(chan bool)
	go func() {
		for {
			_, msg, err := conn.ReadMessage()
			if err != nil {
				common.SysError("error reading stream response: " + err.Error())
				break
			}
			var response XunfeiChatResponse
			err = json.Unmarshal(msg, &response)
			if err != nil {
				common.SysError("error unmarshalling stream response: " + err.Error())
				break
			}
			dataChan <- response
			if response.Payload.Choices.Status == 2 {
				break
			}
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
		case xunfeiResponse := <-dataChan:
			usage.PromptTokens += xunfeiResponse.Usage.Text.PromptTokens
			usage.CompletionTokens += xunfeiResponse.Usage.Text.CompletionTokens
			usage.TotalTokens += xunfeiResponse.Usage.Text.TotalTokens
			response := streamResponseXunfei2OpenAI(&xunfeiResponse)
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
	return nil, &usage
}

func xunfeiHandler(c *gin.Context, resp *http.Response) (*OpenAIErrorWithStatusCode, *Usage) {
	var xunfeiResponse XunfeiChatResponse
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return errorWrapper(err, "read_response_body_failed", http.StatusInternalServerError), nil
	}
	err = resp.Body.Close()
	if err != nil {
		return errorWrapper(err, "close_response_body_failed", http.StatusInternalServerError), nil
	}
	err = json.Unmarshal(responseBody, &xunfeiResponse)
	if err != nil {
		return errorWrapper(err, "unmarshal_response_body_failed", http.StatusInternalServerError), nil
	}
	if xunfeiResponse.Header.Code != 0 {
		return &OpenAIErrorWithStatusCode{
			OpenAIError: OpenAIError{
				Message: xunfeiResponse.Header.Message,
				Type:    "xunfei_error",
				Param:   "",
				Code:    xunfeiResponse.Header.Code,
			},
			StatusCode: resp.StatusCode,
		}, nil
	}
	fullTextResponse := responseXunfei2OpenAI(&xunfeiResponse)
	jsonResponse, err := json.Marshal(fullTextResponse)
	if err != nil {
		return errorWrapper(err, "marshal_response_body_failed", http.StatusInternalServerError), nil
	}
	c.Writer.Header().Set("Content-Type", "application/json")
	c.Writer.WriteHeader(resp.StatusCode)
	_, err = c.Writer.Write(jsonResponse)
	return nil, &fullTextResponse.Usage
}
