package controller

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/songquanpeng/one-api/common"
	"github.com/songquanpeng/one-api/common/config"
	"github.com/songquanpeng/one-api/common/logger"
	"github.com/songquanpeng/one-api/common/message"
	"github.com/songquanpeng/one-api/middleware"
	"github.com/songquanpeng/one-api/model"
	"github.com/songquanpeng/one-api/monitor"
	"github.com/songquanpeng/one-api/relay/constant"
	"github.com/songquanpeng/one-api/relay/helper"
	relaymodel "github.com/songquanpeng/one-api/relay/model"
	"github.com/songquanpeng/one-api/relay/util"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

func buildTestRequest() *relaymodel.GeneralOpenAIRequest {
	testRequest := &relaymodel.GeneralOpenAIRequest{
		MaxTokens: 2,
		Stream:    false,
		Model:     "gpt-3.5-turbo",
	}
	testMessage := relaymodel.Message{
		Role:    "user",
		Content: "hi",
	}
	testRequest.Messages = append(testRequest.Messages, testMessage)
	return testRequest
}

func testChannel(channel *model.Channel) (err error, openaiErr *relaymodel.Error) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = &http.Request{
		Method: "POST",
		URL:    &url.URL{Path: "/v1/chat/completions"},
		Body:   nil,
		Header: make(http.Header),
	}
	c.Request.Header.Set("Authorization", "Bearer "+channel.Key)
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set("channel", channel.Type)
	c.Set("base_url", channel.GetBaseURL())
	middleware.SetupContextForSelectedChannel(c, channel, "")
	meta := util.GetRelayMeta(c)
	apiType := constant.ChannelType2APIType(channel.Type)
	adaptor := helper.GetAdaptor(apiType)
	if adaptor == nil {
		return fmt.Errorf("invalid api type: %d, adaptor is nil", apiType), nil
	}
	adaptor.Init(meta)
	modelName := adaptor.GetModelList()[0]
	if !strings.Contains(channel.Models, modelName) {
		modelNames := strings.Split(channel.Models, ",")
		if len(modelNames) > 0 {
			modelName = modelNames[0]
		}
	}
	request := buildTestRequest()
	request.Model = modelName
	meta.OriginModelName, meta.ActualModelName = modelName, modelName
	convertedRequest, err := adaptor.ConvertRequest(c, constant.RelayModeChatCompletions, request)
	if err != nil {
		return err, nil
	}
	jsonData, err := json.Marshal(convertedRequest)
	if err != nil {
		return err, nil
	}
	requestBody := bytes.NewBuffer(jsonData)
	c.Request.Body = io.NopCloser(requestBody)
	resp, err := adaptor.DoRequest(c, meta, requestBody)
	if err != nil {
		return err, nil
	}
	if resp.StatusCode != http.StatusOK {
		err := util.RelayErrorHandler(resp)
		return fmt.Errorf("status code %d: %s", resp.StatusCode, err.Error.Message), &err.Error
	}
	usage, respErr := adaptor.DoResponse(c, resp, meta)
	if respErr != nil {
		return fmt.Errorf("%s", respErr.Error.Message), &respErr.Error
	}
	if usage == nil {
		return errors.New("usage is nil"), nil
	}
	result := w.Result()
	// print result.Body
	respBody, err := io.ReadAll(result.Body)
	if err != nil {
		return err, nil
	}
	logger.SysLog(fmt.Sprintf("testing channel #%d, response: \n%s", channel.Id, string(respBody)))
	return nil, nil
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
	tik := time.Now()
	err, _ = testChannel(channel)
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

func testChannels(notify bool, scope string) error {
	if config.RootUserEmail == "" {
		config.RootUserEmail = model.GetRootUserEmail()
	}
	testAllChannelsLock.Lock()
	if testAllChannelsRunning {
		testAllChannelsLock.Unlock()
		return errors.New("测试已在运行中")
	}
	testAllChannelsRunning = true
	testAllChannelsLock.Unlock()
	channels, err := model.GetAllChannels(0, 0, scope)
	if err != nil {
		return err
	}
	var disableThreshold = int64(config.ChannelDisableThreshold * 1000)
	if disableThreshold == 0 {
		disableThreshold = 10000000 // a impossible value
	}
	go func() {
		for _, channel := range channels {
			isChannelEnabled := channel.Status == common.ChannelStatusEnabled
			tik := time.Now()
			err, openaiErr := testChannel(channel)
			tok := time.Now()
			milliseconds := tok.Sub(tik).Milliseconds()
			if isChannelEnabled && milliseconds > disableThreshold {
				err = errors.New(fmt.Sprintf("响应时间 %.2fs 超过阈值 %.2fs", float64(milliseconds)/1000.0, float64(disableThreshold)/1000.0))
				if config.AutomaticDisableChannelEnabled {
					monitor.DisableChannel(channel.Id, channel.Name, err.Error())
				} else {
					_ = message.Notify(message.ByAll, fmt.Sprintf("渠道 %s （%d）测试超时", channel.Name, channel.Id), "", err.Error())
				}
			}
			if isChannelEnabled && util.ShouldDisableChannel(openaiErr, -1) {
				monitor.DisableChannel(channel.Id, channel.Name, err.Error())
			}
			if !isChannelEnabled && util.ShouldEnableChannel(err, openaiErr) {
				monitor.EnableChannel(channel.Id, channel.Name)
			}
			channel.UpdateResponseTime(milliseconds)
			time.Sleep(config.RequestInterval)
		}
		testAllChannelsLock.Lock()
		testAllChannelsRunning = false
		testAllChannelsLock.Unlock()
		if notify {
			err := message.Notify(message.ByAll, "渠道测试完成", "", "渠道测试完成，如果没有收到禁用通知，说明所有渠道都正常")
			if err != nil {
				logger.SysError(fmt.Sprintf("failed to send email: %s", err.Error()))
			}
		}
	}()
	return nil
}

func TestChannels(c *gin.Context) {
	scope := c.Query("scope")
	if scope == "" {
		scope = "all"
	}
	err := testChannels(true, scope)
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
		logger.SysLog("testing all channels")
		_ = testChannels(false, "all")
		logger.SysLog("channel test finished")
	}
}
