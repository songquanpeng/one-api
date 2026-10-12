package controller

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/songquanpeng/one-api/common/config"
	"github.com/songquanpeng/one-api/common/helper"
	"github.com/songquanpeng/one-api/model"
	"github.com/songquanpeng/one-api/relay/channeltype"
)

func GetAllChannels(c *gin.Context) {
	p, _ := strconv.Atoi(c.Query("p"))
	if p < 0 {
		p = 0
	}
	channels, err := model.GetAllChannels(p*config.ItemsPerPage, config.ItemsPerPage, "limited")
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
		"data":    channels,
	})
	return
}

func SearchChannels(c *gin.Context) {
	keyword := c.Query("keyword")
	channels, err := model.SearchChannels(keyword)
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
		"data":    channels,
	})
	return
}

func GetChannel(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	channel, err := model.GetChannelById(id, false)
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
		"data":    channel,
	})
	return
}

func AddChannel(c *gin.Context) {
	channel := model.Channel{}
	err := c.ShouldBindJSON(&channel)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	channel.CreatedTime = helper.GetTimestamp()
	keys := strings.Split(channel.Key, "\n")
	channels := make([]model.Channel, 0, len(keys))
	for _, key := range keys {
		if key == "" {
			continue
		}
		localChannel := channel
		localChannel.Key = key
		channels = append(channels, localChannel)
	}
	err = model.BatchInsertChannels(channels)
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

func DeleteChannel(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	channel := model.Channel{Id: id}
	err := channel.Delete()
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

func DeleteDisabledChannel(c *gin.Context) {
	rows, err := model.DeleteDisabledChannel()
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
		"data":    rows,
	})
	return
}

// BatchDeleteChannels 批量删除勾选的渠道
func BatchDeleteChannels(c *gin.Context) {
	var body struct {
		Ids []int `json:"ids"`
	}
	if err := c.ShouldBindJSON(&body); err != nil || len(body.Ids) == 0 {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "参数错误：需要非空的 ids 数组",
		})
		return
	}
	rows, err := model.BatchDeleteChannels(body.Ids)
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
		"data":    rows,
	})
	return
}

// FetchUpstreamModels 从上游 API 拉取模型列表（OpenAI 兼容 /models 端点），
// 用于渠道编辑页「获取模型」一键填充。请求体：{"base_url": "...", "key": "...", "type": 1}
func FetchUpstreamModels(c *gin.Context) {
	var body struct {
		BaseURL string `json:"base_url"`
		Key     string `json:"key"`
		Type    int    `json:"type"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	baseURL := strings.TrimRight(body.BaseURL, "/")
	if baseURL == "" {
		baseURL = channeltype.ChannelBaseURLs[body.Type]
	}
	if baseURL == "" {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "base_url 为空，无法获取模型列表",
		})
		return
	}
	models, err := fetchUpstreamModels(baseURL, body.Key)
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
		"data":    models,
	})
	return
}

// fetchUpstreamModels 依次尝试 {base}/models 与 {base}/v1/models，解析 OpenAI 兼容响应
func fetchUpstreamModels(baseURL string, key string) ([]string, error) {
	candidates := []string{baseURL + "/models"}
	if !strings.HasSuffix(baseURL, "/v1") {
		candidates = append(candidates, baseURL+"/v1/models")
	}
	var lastErr error
	for _, url := range candidates {
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			lastErr = err
			continue
		}
		if key != "" {
			req.Header.Set("Authorization", "Bearer "+key)
		}
		client := &http.Client{Timeout: 15 * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			lastErr = err
			continue
		}
		defer resp.Body.Close()
		raw, err := io.ReadAll(resp.Body)
		if err != nil {
			lastErr = err
			continue
		}
		if resp.StatusCode != http.StatusOK {
			lastErr = fmt.Errorf("上游返回状态码 %d", resp.StatusCode)
			continue
		}
		var parsed struct {
			Data []struct {
				Id string `json:"id"`
			} `json:"data"`
		}
		if err := json.Unmarshal(raw, &parsed); err != nil {
			lastErr = fmt.Errorf("解析上游响应失败: %s", err.Error())
			continue
		}
		if len(parsed.Data) == 0 {
			lastErr = fmt.Errorf("上游返回空模型列表")
			continue
		}
		models := make([]string, 0, len(parsed.Data))
		seen := make(map[string]bool)
		for _, item := range parsed.Data {
			if item.Id != "" && !seen[item.Id] {
				seen[item.Id] = true
				models = append(models, item.Id)
			}
		}
		return models, nil
	}
	if lastErr == nil {
		lastErr = fmt.Errorf("无法连接上游")
	}
	return nil, lastErr
}

func UpdateChannel(c *gin.Context) {
	channel := model.Channel{}
	err := c.ShouldBindJSON(&channel)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	err = channel.Update()
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
		"data":    channel,
	})
	return
}
