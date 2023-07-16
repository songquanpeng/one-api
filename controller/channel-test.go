package controller

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"one-api/common"
	"one-api/model"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

func formatFloat(input float64) float64 {
	if input == float64(int64(input)) {
		return input
	}
	return float64(int64(input*10)) / 10
}

func testChannel(channel *model.Channel, request ChatRequest) error {
	switch channel.Type {
	case common.ChannelTypeAzure:
		request.Model = "gpt-35-turbo"
	default:
		request.Model = "gpt-3.5-turbo"
	}
	requestURL := common.ChannelBaseURLs[channel.Type]
	if channel.Type == common.ChannelTypeAzure {
		requestURL = fmt.Sprintf("%s/openai/deployments/%s/chat/completions?api-version=2023-03-15-preview", channel.BaseURL, request.Model)
	} else if channel.Type == common.ChannelTypeChatGPTWeb {
		if channel.BaseURL != "" {
			requestURL = channel.BaseURL
		}
	} else {
		if channel.BaseURL != "" {
			requestURL = channel.BaseURL
		}
		requestURL += "/v1/chat/completions"
	}

	jsonData, err := json.Marshal(request)

	if channel.Type == common.ChannelTypeChatGPTWeb {
		// Get system message from Message json, Role == "system"
		var systemMessage Message

		for _, message := range request.Messages {
			if message.Role == "system" {
				systemMessage = message
				break
			}
		}

		var prompt string

		// Get all the Message, Roles from request.Messages, and format it into string by
		// ||> role: content
		for _, message := range request.Messages {
			// Exclude system message
			if message.Role == "system" {
				continue
			}
			prompt += "||> " + message.Role + ": " + message.Content + "\n"
		}

		// Construct json data without adding escape character
		map1 := make(map[string]interface{})

		map1["prompt"] = prompt
		map1["systemMessage"] = systemMessage.Content

		if request.Temperature != 0 {
			map1["temperature"] = formatFloat(request.Temperature)
		}
		if request.TopP != 0 {
			map1["top_p"] = formatFloat(request.TopP)
		}

		// Convert map to json string
		jsonData, err = json.Marshal(map1)
	}
	if err != nil {
		return err
	}
	req, err := http.NewRequest("POST", requestURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	if channel.Type == common.ChannelTypeAzure {
		req.Header.Set("api-key", channel.Key)
	} else {
		req.Header.Set("Authorization", "Bearer "+channel.Key)
	}
	req.Header.Set("Content-Type", "application/json")

	if channel.EnableIpRandomization {
		// Generate random IP
		ip := common.GenerateIP()
		req.Header.Set("X-Forwarded-For", ip)
		req.Header.Set("X-Real-IP", ip)
		req.Header.Set("X-Client-IP", ip)
		req.Header.Set("X-Forwarded-Host", ip)
		req.Header.Set("X-Originating-IP", ip)
		req.RemoteAddr = ip
		req.Header.Set("X-Remote-IP", ip)
		req.Header.Set("X-Remote-Addr", ip)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		// Print the body in string
		if resp.Body != nil {
			buf := new(bytes.Buffer)
			buf.ReadFrom(resp.Body)
			return errors.New("error response: " + strconv.Itoa(resp.StatusCode) + " " + buf.String())
		}

		return errors.New("error response: " + strconv.Itoa(resp.StatusCode))
	}

	var streamResponseText = ""

	scanner := bufio.NewScanner(resp.Body)

	if channel.Type != common.ChannelTypeChatGPTWeb {
		scanner.Split(func(data []byte, atEOF bool) (advance int, token []byte, err error) {
			if atEOF && len(data) == 0 {
				return 0, nil, nil
			}

			if i := strings.Index(string(data), "\n"); i >= 0 {
				return i + 2, data[0:i], nil
			}

			if atEOF {
				return len(data), data, nil
			}

			return 0, nil, nil
		})
	}

	for scanner.Scan() {
		data := scanner.Text()
		if len(data) < 6 { // must be something wrong!
			continue
		}

		if channel.Type != common.ChannelTypeChatGPTWeb {
			// If data has event: event content inside, remove it, it can be prefix or inside the data
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

				// Trim whitespace and newlines from the modified data string
				data = strings.TrimSpace(data)
			}
			if !strings.HasPrefix(data, "data:") {
				continue
			}
			data = data[6:]
			if !strings.HasPrefix(data, "[DONE]") {
				var streamResponse ChatCompletionsStreamResponse
				err = json.Unmarshal([]byte(data), &streamResponse)
				if err != nil {
					// Prinnt the body in string
					buf := new(bytes.Buffer)
					buf.ReadFrom(resp.Body)
					common.SysError("error unmarshalling stream response: " + err.Error() + " " + buf.String())
					return err
				}
				for _, choice := range streamResponse.Choices {
					streamResponseText += choice.Delta.Content
				}
			}

		} else if channel.Type == common.ChannelTypeChatGPTWeb {
			var chatResponse ChatGptWebChatResponse
			err = json.Unmarshal([]byte(data), &chatResponse)
			if err != nil {
				// Print the body in string
				buf := new(bytes.Buffer)
				buf.ReadFrom(resp.Body)
				common.SysError("error unmarshalling chat response: " + err.Error() + " " + buf.String())
				return err
			}

			// if response role is assistant and contains delta, append the content to streamResponseText
			if chatResponse.Role == "assistant" && chatResponse.Detail != nil {
				for _, choice := range chatResponse.Detail.Choices {
					streamResponseText += choice.Delta.Content
				}
			}

		}
	}

	defer resp.Body.Close()

	// Check if streaming is complete and streamResponseText is populated
	if streamResponseText == "" {
		return errors.New("Streaming not complete")
	}

	return nil
}

func buildTestRequest() *ChatRequest {
	testRequest := &ChatRequest{
		Model:  "", // this will be set later
		Stream: true,
	}
	testMessage := Message{
		Role:    "user",
		Content: "Hello ChatGPT!",
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
	testRequest := buildTestRequest()
	tik := time.Now()
	err = testChannel(channel, *testRequest)
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
	model.UpdateChannelStatusById(channelId, common.ChannelStatusDisabled)
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
			err := testChannel(channel, *testRequest)
			tok := time.Now()
			milliseconds := tok.Sub(tik).Milliseconds()
			if err != nil || milliseconds > disableThreshold {
				if milliseconds > disableThreshold {
					err = errors.New(fmt.Sprintf("响应时间 %.2fs 超过阈值 %.2fs", float64(milliseconds)/1000.0, float64(disableThreshold)/1000.0))
				}
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
