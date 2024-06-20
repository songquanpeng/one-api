package xunfei

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/songquanpeng/one-api/common"
	"github.com/songquanpeng/one-api/common/helper"
	"github.com/songquanpeng/one-api/common/logger"
	"github.com/songquanpeng/one-api/common/random"
	"github.com/songquanpeng/one-api/relay/adaptor/openai"
	"github.com/songquanpeng/one-api/relay/constant"
	"github.com/songquanpeng/one-api/relay/meta"
	"github.com/songquanpeng/one-api/relay/model"
)

// https://console.xfyun.cn/services/cbm
// https://www.xfyun.cn/doc/spark/Web.html

func requestOpenAI2Xunfei(request model.GeneralOpenAIRequest, xunfeiAppId string, domain string) *ChatRequest {
	messages := make([]Message, 0, len(request.Messages))
	for _, message := range request.Messages {
		messages = append(messages, Message{
			Role:    message.Role,
			Content: message.StringContent(),
		})
	}
	xunfeiRequest := ChatRequest{}
	xunfeiRequest.Header.AppId = xunfeiAppId
	xunfeiRequest.Parameter.Chat.Domain = domain
	xunfeiRequest.Parameter.Chat.Temperature = request.Temperature
	xunfeiRequest.Parameter.Chat.TopK = request.N
	xunfeiRequest.Parameter.Chat.MaxTokens = request.MaxTokens
	xunfeiRequest.Payload.Message.Text = messages

	if strings.HasPrefix(domain, "generalv3") {
		functions := make([]model.Function, len(request.Tools))
		for i, tool := range request.Tools {
			functions[i] = tool.Function
		}
		xunfeiRequest.Payload.Functions = &Functions{
			Text: functions,
		}
	}

	return &xunfeiRequest
}

func getToolCalls(response *ChatResponse) []model.Tool {
	var toolCalls []model.Tool
	if len(response.Payload.Choices.Text) == 0 {
		return toolCalls
	}
	item := response.Payload.Choices.Text[0]
	if item.FunctionCall == nil {
		return toolCalls
	}
	toolCall := model.Tool{
		Id:       fmt.Sprintf("call_%s", random.GetUUID()),
		Type:     "function",
		Function: *item.FunctionCall,
	}
	toolCalls = append(toolCalls, toolCall)
	return toolCalls
}

func responseXunfei2OpenAI(response *ChatResponse) *openai.TextResponse {
	if len(response.Payload.Choices.Text) == 0 {
		response.Payload.Choices.Text = []ChatResponseTextItem{
			{
				Content: "",
			},
		}
	}
	choice := openai.TextResponseChoice{
		Index: 0,
		Message: model.Message{
			Role:      "assistant",
			Content:   response.Payload.Choices.Text[0].Content,
			ToolCalls: getToolCalls(response),
		},
		FinishReason: constant.StopFinishReason,
	}
	fullTextResponse := openai.TextResponse{
		Id:      fmt.Sprintf("chatcmpl-%s", random.GetUUID()),
		Object:  "chat.completion",
		Created: helper.GetTimestamp(),
		Choices: []openai.TextResponseChoice{choice},
		Usage:   response.Payload.Usage.Text,
	}
	return &fullTextResponse
}

func streamResponseXunfei2OpenAI(xunfeiResponse *ChatResponse) *openai.ChatCompletionsStreamResponse {
	if len(xunfeiResponse.Payload.Choices.Text) == 0 {
		xunfeiResponse.Payload.Choices.Text = []ChatResponseTextItem{
			{
				Content: "",
			},
		}
	}
	var choice openai.ChatCompletionsStreamResponseChoice
	choice.Delta.Content = xunfeiResponse.Payload.Choices.Text[0].Content
	choice.Delta.ToolCalls = getToolCalls(xunfeiResponse)
	if xunfeiResponse.Payload.Choices.Status == 2 {
		choice.FinishReason = &constant.StopFinishReason
	}
	response := openai.ChatCompletionsStreamResponse{
		Id:      fmt.Sprintf("chatcmpl-%s", random.GetUUID()),
		Object:  "chat.completion.chunk",
		Created: helper.GetTimestamp(),
		Model:   "SparkDesk",
		Choices: []openai.ChatCompletionsStreamResponseChoice{choice},
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

func StreamHandler(c *gin.Context, meta *meta.Meta, textRequest model.GeneralOpenAIRequest, appId string, apiSecret string, apiKey string) (*model.ErrorWithStatusCode, *model.Usage) {
	domain, authUrl := getXunfeiAuthUrl(meta.Config.APIVersion, apiKey, apiSecret)
	dataChan, stopChan, err := xunfeiMakeRequest(textRequest, domain, authUrl, appId)
	if err != nil {
		return openai.ErrorWrapper(err, "xunfei_request_failed", http.StatusInternalServerError), nil
	}
	common.SetEventStreamHeaders(c)
	var usage model.Usage
	c.Stream(func(w io.Writer) bool {
		select {
		case xunfeiResponse := <-dataChan:
			usage.PromptTokens += xunfeiResponse.Payload.Usage.Text.PromptTokens
			usage.CompletionTokens += xunfeiResponse.Payload.Usage.Text.CompletionTokens
			usage.TotalTokens += xunfeiResponse.Payload.Usage.Text.TotalTokens
			response := streamResponseXunfei2OpenAI(&xunfeiResponse)
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
	return nil, &usage
}

func Handler(c *gin.Context, meta *meta.Meta, textRequest model.GeneralOpenAIRequest, appId string, apiSecret string, apiKey string) (*model.ErrorWithStatusCode, *model.Usage) {
	domain, authUrl := getXunfeiAuthUrl(meta.Config.APIVersion, apiKey, apiSecret)
	dataChan, stopChan, err := xunfeiMakeRequest(textRequest, domain, authUrl, appId)
	if err != nil {
		return openai.ErrorWrapper(err, "xunfei_request_failed", http.StatusInternalServerError), nil
	}
	var usage model.Usage
	var content string
	var xunfeiResponse ChatResponse
	stop := false
	for !stop {
		select {
		case xunfeiResponse = <-dataChan:
			if len(xunfeiResponse.Payload.Choices.Text) == 0 {
				continue
			}
			content += xunfeiResponse.Payload.Choices.Text[0].Content
			usage.PromptTokens += xunfeiResponse.Payload.Usage.Text.PromptTokens
			usage.CompletionTokens += xunfeiResponse.Payload.Usage.Text.CompletionTokens
			usage.TotalTokens += xunfeiResponse.Payload.Usage.Text.TotalTokens
		case stop = <-stopChan:
		}
	}
	if len(xunfeiResponse.Payload.Choices.Text) == 0 {
		return openai.ErrorWrapper(errors.New("xunfei empty response detected"), "xunfei_empty_response_detected", http.StatusInternalServerError), nil
	}
	xunfeiResponse.Payload.Choices.Text[0].Content = content

	response := responseXunfei2OpenAI(&xunfeiResponse)
	jsonResponse, err := json.Marshal(response)
	if err != nil {
		return openai.ErrorWrapper(err, "marshal_response_body_failed", http.StatusInternalServerError), nil
	}
	c.Writer.Header().Set("Content-Type", "application/json")
	_, _ = c.Writer.Write(jsonResponse)
	return nil, &usage
}

func xunfeiMakeRequest(textRequest model.GeneralOpenAIRequest, domain, authUrl, appId string) (chan ChatResponse, chan bool, error) {
	d := websocket.Dialer{
		HandshakeTimeout: 5 * time.Second,
	}
	conn, resp, err := d.Dial(authUrl, nil)
	if err != nil || resp.StatusCode != 101 {
		return nil, nil, err
	}
	data := requestOpenAI2Xunfei(textRequest, appId, domain)
	err = conn.WriteJSON(data)
	if err != nil {
		return nil, nil, err
	}
	_, msg, err := conn.ReadMessage()
	if err != nil {
		return nil, nil, err
	}

	dataChan := make(chan ChatResponse)
	stopChan := make(chan bool)
	go func() {
		for {
			if msg == nil {
				_, msg, err = conn.ReadMessage()
				if err != nil {
					logger.SysError("error reading stream response: " + err.Error())
					break
				}
			}
			var response ChatResponse
			err = json.Unmarshal(msg, &response)
			if err != nil {
				logger.SysError("error unmarshalling stream response: " + err.Error())
				break
			}
			msg = nil
			dataChan <- response
			if response.Payload.Choices.Status == 2 {
				err := conn.Close()
				if err != nil {
					logger.SysError("error closing websocket connection: " + err.Error())
				}
				break
			}
		}
		stopChan <- true
	}()

	return dataChan, stopChan, nil
}

func parseAPIVersionByModelName(modelName string) string {
	parts := strings.Split(modelName, "-")
	if len(parts) == 2 {
		return parts[1]
	}
	return ""
}

// https://www.xfyun.cn/doc/spark/Web.html#_1-%E6%8E%A5%E5%8F%A3%E8%AF%B4%E6%98%8E
func apiVersion2domain(apiVersion string) string {
	switch apiVersion {
	case "v1.1":
		return "general"
	case "v2.1":
		return "generalv2"
	case "v3.1":
		return "generalv3"
	case "v3.5":
		return "generalv3.5"
	}
	return "general" + apiVersion
}

func getXunfeiAuthUrl(apiVersion string, apiKey string, apiSecret string) (string, string) {
	domain := apiVersion2domain(apiVersion)
	authUrl := buildXunfeiAuthUrl(fmt.Sprintf("wss://spark-api.xf-yun.com/%s/chat", apiVersion), apiKey, apiSecret)
	return domain, authUrl
}
