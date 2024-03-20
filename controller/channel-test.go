package controller

import (
	"encoding/json"
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
	if channel.TestModel == "" {
		return errors.New("请填写测速模型后再试"), nil
	}

	// 创建一个 http.Request
	req, err := http.NewRequest("POST", "/v1/chat/completions", nil)
	if err != nil {
		return err, nil
	}
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	request.Model = channel.TestModel

	provider := providers.GetProvider(channel, c)
	if provider == nil {
		return errors.New("channel not implemented"), nil
	}

	newModelName, err := provider.ModelMappingHandler(request.Model)
	if err != nil {
		return err, nil
	}

	request.Model = newModelName

	chatProvider, ok := provider.(providers_base.ChatInterface)
	if !ok {
		return errors.New("channel not implemented"), nil
	}

	chatProvider.SetUsage(&types.Usage{})

	response, openAIErrorWithStatusCode := chatProvider.CreateChatCompletion(&request)

	if openAIErrorWithStatusCode != nil {
		return errors.New(openAIErrorWithStatusCode.Message), &openAIErrorWithStatusCode.OpenAIError
	}

	usage := chatProvider.GetUsage()

	if usage.CompletionTokens == 0 {
		return fmt.Errorf("channel %s, message 补全 tokens 非预期返回 0", channel.Name), nil
	}

	// 转换为JSON字符串
	jsonBytes, _ := json.Marshal(response)
	common.SysLog(fmt.Sprintf("测试模型 %s 返回内容为：%s", channel.Name, string(jsonBytes)))

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
		Model: "",
		// MaxTokens: 1,
		Stream: false,
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
}

var testAllChannelsLock sync.Mutex
var testAllChannelsRunning bool = false

func notifyRootUser(subject string, content string) {
	if common.RootUserEmail == "" {
		common.RootUserEmail = model.GetRootUserEmail()
	}
	err := common.SendEmail(subject, common.RootUserEmail, content)
	if err != nil {
		common.SysError(fmt.Sprintf("failed to send email: %s", err.Error()))
	}
}

// enable & notify
func enableChannel(channelId int, channelName string) {
	model.UpdateChannelStatusById(channelId, common.ChannelStatusEnabled)
	subject := fmt.Sprintf("通道「%s」（#%d）已被启用", channelName, channelId)
	content := fmt.Sprintf("通道「%s」（#%d）已被启用", channelName, channelId)
	notifyRootUser(subject, content)
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
	channels, err := model.GetAllChannels()
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
			isChannelEnabled := channel.Status == common.ChannelStatusEnabled
			tik := time.Now()
			err, openaiErr := testChannel(channel, *testRequest)
			tok := time.Now()
			milliseconds := tok.Sub(tik).Milliseconds()
			if milliseconds > disableThreshold {
				err = fmt.Errorf("响应时间 %.2fs 超过阈值 %.2fs", float64(milliseconds)/1000.0, float64(disableThreshold)/1000.0)
				DisableChannel(channel.Id, channel.Name, err.Error())
			}
			if isChannelEnabled && ShouldDisableChannel(openaiErr, -1) {
				DisableChannel(channel.Id, channel.Name, err.Error())
			}
			if !isChannelEnabled && shouldEnableChannel(err, openaiErr) {
				enableChannel(channel.Id, channel.Name)
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
	if frequency <= 0 {
		return
	}

	for {
		time.Sleep(time.Duration(frequency) * time.Minute)
		common.SysLog("testing all channels")
		_ = testAllChannels(false)
		common.SysLog("channel test finished")
	}
}
