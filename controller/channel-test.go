package controller

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"one-api/common"
	"one-api/model"
	"one-api/providers"
	providers_base "one-api/providers/base"
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

	setChannelToContext(c, channel)
	// 创建映射
	channelTypeToModel := map[int]string{
		common.ChannelTypePaLM:      "PaLM-2",
		common.ChannelTypeAnthropic: "claude-2",
		common.ChannelTypeBaidu:     "ERNIE-Bot",
		common.ChannelTypeZhipu:     "chatglm_lite",
		common.ChannelTypeAli:       "qwen-turbo",
		common.ChannelType360:       "360GPT_S2_V9",
		common.ChannelTypeXunfei:    "SparkDesk",
		common.ChannelTypeTencent:   "hunyuan",
		common.ChannelTypeAzure:     "gpt-3.5-turbo",
	}

	// 从映射中获取模型名称
	model, ok := channelTypeToModel[channel.Type]
	if !ok {
		model = "gpt-3.5-turbo" // 默认值
	}
	request.Model = model

	provider := providers.GetProvider(channel.Type, c)
	if provider == nil {
		return errors.New("channel not implemented"), nil
	}
	chatProvider, ok := provider.(providers_base.ChatInterface)
	if !ok {
		return errors.New("channel not implemented"), nil
	}

	modelMap, err := parseModelMapping(channel.GetModelMapping())
	if err != nil {
		return err, nil
	}
	if modelMap != nil && modelMap[request.Model] != "" {
		request.Model = modelMap[request.Model]
	}

	promptTokens := common.CountTokenMessages(request.Messages, request.Model)
	_, openAIErrorWithStatusCode := chatProvider.ChatAction(&request, true, promptTokens)
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
				err = fmt.Errorf("响应时间 %.2fs 超过阈值 %.2fs", float64(milliseconds)/1000.0, float64(disableThreshold)/1000.0)
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
}

func AutomaticallyTestChannels(frequency int) {
	for {
		time.Sleep(time.Duration(frequency) * time.Minute)
		common.SysLog("testing all channels")
		_ = testAllChannels(false)
		common.SysLog("channel test finished")
	}
}
