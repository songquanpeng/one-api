package relay

import (
	"net/http"
	"one-api/common"
	"one-api/common/config"
	"one-api/model"
	"one-api/providers/azure"
	"one-api/providers/openai"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

func RelayOnly(c *gin.Context) {
	provider, _, fail := GetProvider(c, "")
	if fail != nil {
		common.AbortWithMessage(c, http.StatusServiceUnavailable, fail.Error())
		return
	}

	channel := provider.GetChannel()
	if channel.Type != config.ChannelTypeOpenAI && channel.Type != config.ChannelTypeAzure {
		common.AbortWithMessage(c, http.StatusServiceUnavailable, "provider must be of type azureopenai or openai")
		return
	}

	// 获取请求的path
	url := ""
	path := c.Request.URL.Path
	openAIProvider, ok := provider.(*openai.OpenAIProvider)
	if !ok {
		azureProvider, ok := provider.(*azure.AzureProvider)
		if !ok {
			common.AbortWithMessage(c, http.StatusServiceUnavailable, "provider must be of type openai")
			return
		}
		url = azureProvider.GetFullRequestURL(path, "")
	} else {
		url = openAIProvider.GetFullRequestURL(path, "")
	}

	headers := c.Request.Header
	mapHeaders := provider.GetRequestHeaders()
	// 设置请求头
	for k, v := range headers {
		if _, ok := mapHeaders[k]; ok {
			continue
		}
		mapHeaders[k] = strings.Join(v, ", ")
	}

	requester := provider.GetRequester()
	req, err := requester.NewRequest(c.Request.Method, url, requester.WithBody(c.Request.Body), requester.WithHeader(mapHeaders))
	if err != nil {
		common.AbortWithMessage(c, http.StatusBadRequest, err.Error())
		return
	}

	defer req.Body.Close()

	response, errWithCode := requester.SendRequestRaw(req)
	if errWithCode != nil {
		relayResponseWithErr(c, errWithCode)
		return
	}

	errWithCode = responseMultipart(c, response)

	if errWithCode != nil {
		relayResponseWithErr(c, errWithCode)
		return
	}

	requestTime := 0
	requestStartTimeValue := c.Request.Context().Value("requestStartTime")
	if requestStartTimeValue != nil {
		requestStartTime, ok := requestStartTimeValue.(time.Time)
		if ok {
			requestTime = int(time.Since(requestStartTime).Milliseconds())
		}
	}
	model.RecordConsumeLog(c.Request.Context(), c.GetInt("id"), c.GetInt("channel_id"), 0, 0, "", c.GetString("token_name"), 0, "中继:"+path, requestTime)

}
