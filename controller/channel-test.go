package controller

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"github.com/gin-gonic/gin"
	"net/http"
	"one-api/common"
	"one-api/model"
	"strconv"
	"strings"
	"sync"
	"time"
)

func testChannel(channel *model.Channel, request ChatRequest) (err error, openaiErr *OpenAIError) {
	switch channel.Type {
	case common.ChannelTypePaLM:
		fallthrough
	case common.ChannelTypeAnthropic:
		fallthrough
	case common.ChannelTypeBaidu:
		fallthrough
	case common.ChannelTypeZhipu:
		fallthrough
	case common.ChannelTypeAli:
		fallthrough
	case common.ChannelType360:
		fallthrough
	case common.ChannelTypeXunfei:
		return errors.New("该渠道类型当前版本不支持测试，请手动测试"), nil
	case common.ChannelTypeAzure:
		request.Model = "gpt-35-turbo"
		defer func() {
			if err != nil {
				err = errors.New("请确保已在 Azure 上创建了 gpt-35-turbo 模型，并且 apiVersion 已正确填写！")
			}
		}()
	default:
		request.Model = "gpt-3.5-turbo"
	}
	requestURL := common.ChannelBaseURLs[channel.Type]
	if channel.Type == common.ChannelTypeAzure {
		requestURL = fmt.Sprintf("%s/openai/deployments/%s/chat/completions?api-version=2023-03-15-preview", channel.GetBaseURL(), request.Model)
	} else {
		if channel.GetBaseURL() != "" {
			requestURL = channel.GetBaseURL()
		}
		requestURL += "/v1/chat/completions"
	}
	// for Cloudflare AI gateway: https://github.com/songquanpeng/one-api/pull/639
	requestURL = strings.Replace(requestURL, "/v1/v1", "/v1", 1)

	jsonData, err := json.Marshal(request)
	if err != nil {
		return err, nil
	}
	req, err := http.NewRequest("POST", requestURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return err, nil
	}
	if channel.Type == common.ChannelTypeAzure {
		req.Header.Set("api-key", channel.Key)
	} else {
		req.Header.Set("Authorization", "Bearer "+channel.Key)
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := httpClient.Do(req)
	if err != nil {
		return err, nil
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		// Log: Channel (channel_id) Request Error (StatusCode)
		log.Print(fmt.Sprintf("Channel (%d) Request Error (%d)", channel.Id, resp.StatusCode))
		return errors.New(fmt.Sprintf("status code %d", resp.StatusCode)), nil
	}

	if channel.AllowStreaming == common.ChannelAllowStreamEnabled {
		responseText := ""
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
		for scanner.Scan() {
			data := scanner.Text()
			if len(data) < 6 { // ignore blank line or wrong format
				continue
			}
			// ChatGPT Next Web
			if strings.HasPrefix(data, "event:") || strings.Contains(data, "event:") {
				// Remove event: event in the front or back
				data = strings.TrimPrefix(data, "event: event")
				data = strings.TrimSuffix(data, "event: event")
				// Remove everything, only keep `data: {...}` <--- this is the json
				// Find the start and end indices of `data: {...}` substring
				startIndex := strings.Index(data, "data:")
				endIndex := strings.LastIndex(data, "}")

				// If both indices are found and end index is greater than start index
				if startIndex != -1 && endIndex != -1 && endIndex > startIndex {
					// Extract the `data: {...}` substring
					data = data[startIndex : endIndex+1]
				}
			}
			if !strings.HasPrefix(data, "data:") {
				continue
			}
			data = strings.TrimPrefix(data, "data:")
			if !strings.HasPrefix(data, "[DONE]") {
				var streamResponse ChatCompletionsStreamResponse
				err := json.Unmarshal([]byte(data), &streamResponse)
				if err == nil {
					for _, choice := range streamResponse.Choices {
						responseText += choice.Delta.Content
					}
				}
			}
		}

		if responseText == "" {
			return errors.New("Empty response"), nil
		}
	} else if channel.AllowNonStreaming == common.ChannelAllowNonStreamEnabled {
		var response TextResponse
		err = json.NewDecoder(resp.Body).Decode(&response)
		if err != nil {
			return err, nil
		}
		if response.Usage.CompletionTokens == 0 {
			return errors.New(fmt.Sprintf("type %s, code %v, message %s", response.Error.Type, response.Error.Code, response.Error.Message)), &response.Error
		}
	}

	return nil, nil
}

func buildTestRequest(stream bool) *ChatRequest {
	testRequest := &ChatRequest{
		Model:     "", // this will be set later
		MaxTokens: 1,
		Stream:    stream,
	}
	testMessage := Message{
		Role:    "user",
		Content: "hi",
	}
	testRequest.Messages = append(testRequest.Messages, testMessage)
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
	testRequest := buildTestRequest(channel.AllowStreaming == common.ChannelAllowStreamEnabled)
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

// enable
func enableChannel(channelId int, channelName string) {
	if common.RootUserEmail == "" {
		common.RootUserEmail = model.GetRootUserEmail()
	}
	model.UpdateChannelStatusById(channelId, common.ChannelStatusEnabled)
	subject := fmt.Sprintf("通道「%s」（#%d）已被重新启用", channelName, channelId)
	content := fmt.Sprintf("通道「%s」（#%d）通道现在满足了服务器要求, 已被重新启用并重新开始运行", channelName, channelId)
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
	var disableThreshold = int64(common.ChannelDisableThreshold * 1000)
	if disableThreshold == 0 {
		disableThreshold = 10000000 // a impossible value
	}
	go func() {
		for _, channel := range channels {
			if channel.Status == common.ChannelStatusManuallyDisabled || channel.Status == common.ChannelStatusUnknown {
				continue
			}
			tik := time.Now()
			testRequest := buildTestRequest(channel.AllowStreaming == common.ChannelAllowStreamEnabled)
			err, openaiErr := testChannel(channel, *testRequest)
			tok := time.Now()
			milliseconds := tok.Sub(tik).Milliseconds()
			channelBeninDisabled := false
			if milliseconds > disableThreshold {
				err = errors.New(fmt.Sprintf("响应时间 %.2fs 超过阈值 %.2fs", float64(milliseconds)/1000.0, float64(disableThreshold)/1000.0))
				disableChannel(channel.Id, channel.Name, err.Error())
				channelBeninDisabled = true
			}
			if shouldDisableChannel(openaiErr, -1) {
				disableChannel(channel.Id, channel.Name, err.Error())
				channelBeninDisabled = true
			}
			if channel.Status == common.ChannelStatusAutoDisabled && common.AutoReEnableFailedChannelEnabled && !channelBeninDisabled {
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
