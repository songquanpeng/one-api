package controller

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"one-api/common"
	"one-api/model"
	"one-api/types"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

func testChannel(channel *model.Channel, request types.ChatCompletionRequest) (err error, openaiErr *types.OpenAIError) {
	// 创建一个 http.Request
	req, err := http.NewRequest("POST", "/v1/chat/completions", nil)
	if err != nil {
		return err, nil
	}
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("channel", channel.Type)
	c.Set("channel_id", channel.Id)
	c.Set("channel_name", channel.Name)
	c.Set("model_mapping", channel.GetModelMapping())
	c.Set("api_key", channel.Key)
	c.Set("base_url", channel.GetBaseURL())

	switch channel.Type {
	case common.ChannelTypePaLM:
		request.Model = "PaLM-2"
	case common.ChannelTypeAnthropic:
		request.Model = "claude-2"
	case common.ChannelTypeBaidu:
		request.Model = "ERNIE-Bot"
	case common.ChannelTypeZhipu:
		request.Model = "chatglm_lite"
	case common.ChannelTypeAli:
		request.Model = "qwen-turbo"
	case common.ChannelType360:
		request.Model = "360GPT_S2_V9"
	case common.ChannelTypeXunfei:
		request.Model = "SparkDesk"
		c.Set("api_version", channel.Other)
	case common.ChannelTypeTencent:
		request.Model = "hunyuan"
	case common.ChannelTypeAzure:
		request.Model = "gpt-3.5-turbo"
		c.Set("api_version", channel.Other)
	default:
		request.Model = "gpt-3.5-turbo"
	}

	chatProvider := GetChatProvider(channel.Type, c)
	isModelMapped := false
	modelMap, err := parseModelMapping(c.GetString("model_mapping"))
	if err != nil {
		return err, nil
	}
	if modelMap != nil && modelMap[request.Model] != "" {
		request.Model = modelMap[request.Model]
		isModelMapped = true
	}

	promptTokens := common.CountTokenMessages(request.Messages, request.Model)
	_, openAIErrorWithStatusCode := chatProvider.ChatAction(&request, isModelMapped, promptTokens)
	if openAIErrorWithStatusCode != nil {
		return nil, &openAIErrorWithStatusCode.OpenAIError
	}

	return nil, nil
}

func buildTestRequest() *types.ChatCompletionRequest {
	testRequest := &types.ChatCompletionRequest{
		Messages: []types.ChatCompletionMessage{
			{
				Role:    "user",
				Content: "You just need to output 'hi' next.",
			},
		},
		Model:     "",
		MaxTokens: 1,
		Stream:    false,
	}
	return testRequest
}

func TestChannel(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	channel, err := model.GetChannelById(id, true)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	testRequest := buildTestRequest()
	tik := time.Now()
	err, _ = testChannel(channel, *testRequest)
	tok := time.Now()
	milliseconds := tok.Sub(tik).Milliseconds()
	go channel.UpdateResponseTime(milliseconds)
	consumedTime := float64(milliseconds) / 1000.0
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
			"time":    consumedTime,
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"time":    consumedTime,
	})
	return
}

var testAllChannelsLock sync.Mutex
var testAllChannelsRunning bool = false

// disable & notify
func disableChannel(channelId int, channelName string, reason string) {
	if common.RootUserEmail == "" {
		common.RootUserEmail = model.GetRootUserEmail()
	}
	model.UpdateChannelStatusById(channelId, common.ChannelStatusAutoDisabled)
	subject := fmt.Sprintf("通道「%s」（#%d）已被禁用", channelName, channelId)
	content := fmt.Sprintf("通道「%s」（#%d）已被禁用，原因：%s", channelName, channelId, reason)
	err := common.SendEmail(subject, common.RootUserEmail, content)
	if err != nil {
		common.SysError(fmt.Sprintf("failed to send email: %s", err.Error()))
	}
}

func testAllChannels(notify bool) error {
	if common.RootUserEmail == "" {
		common.RootUserEmail = model.GetRootUserEmail()
	}
	testAllChannelsLock.Lock()
	if testAllChannelsRunning {
		testAllChannelsLock.Unlock()
		return errors.New("测试已在运行中")
	}
	testAllChannelsRunning = true
	testAllChannelsLock.Unlock()
	channels, err := model.GetAllChannels(0, 0, true)
	if err != nil {
		return err
	}
	testRequest := buildTestRequest()
	var disableThreshold = int64(common.ChannelDisableThreshold * 1000)
	if disableThreshold == 0 {
		disableThreshold = 10000000 // a impossible value
	}
	go func() {
		for _, channel := range channels {
			if channel.Status != common.ChannelStatusEnabled {
				continue
			}
			tik := time.Now()
			err, openaiErr := testChannel(channel, *testRequest)
			tok := time.Now()
			milliseconds := tok.Sub(tik).Milliseconds()
			if milliseconds > disableThreshold {
				err = errors.New(fmt.Sprintf("响应时间 %.2fs 超过阈值 %.2fs", float64(milliseconds)/1000.0, float64(disableThreshold)/1000.0))
				disableChannel(channel.Id, channel.Name, err.Error())
			}
			if shouldDisableChannel(openaiErr, -1) {
				disableChannel(channel.Id, channel.Name, err.Error())
			}
			channel.UpdateResponseTime(milliseconds)
			time.Sleep(common.RequestInterval)
		}
		testAllChannelsLock.Lock()
		testAllChannelsRunning = false
		testAllChannelsLock.Unlock()
		if notify {
			err := common.SendEmail("通道测试完成", common.RootUserEmail, "通道测试完成，如果没有收到禁用通知，说明所有通道都正常")
			if err != nil {
				common.SysError(fmt.Sprintf("failed to send email: %s", err.Error()))
			}
		}
	}()
	return nil
}

func TestAllChannels(c *gin.Context) {
	err := testAllChannels(true)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
	})
	return
}

func AutomaticallyTestChannels(frequency int) {
	for {
		time.Sleep(time.Duration(frequency) * time.Minute)
		common.SysLog("testing all channels")
		_ = testAllChannels(false)
		common.SysLog("channel test finished")
	}
}
