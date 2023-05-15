package middleware

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"one-api/common"
	"one-api/model"
	"strconv"
)

func Distribute() func(c *gin.Context) {
	return func(c *gin.Context) {
		var channel *model.Channel
		channelId, ok := c.Get("channelId")
		if ok {
			id, err := strconv.Atoi(channelId.(string))
			if err != nil {
				c.JSON(http.StatusOK, gin.H{
					"error": gin.H{
						"message": "无效的渠道 ID",
						"type":    "one_api_error",
					},
				})
				c.Abort()
				return
			}
			channel, err = model.GetChannelById(id, true)
			if err != nil {
				c.JSON(200, gin.H{
					"error": gin.H{
						"message": "无效的渠道 ID",
						"type":    "one_api_error",
					},
				})
				c.Abort()
				return
			}
			if channel.Status != common.ChannelStatusEnabled {
				c.JSON(200, gin.H{
					"error": gin.H{
						"message": "该渠道已被禁用",
						"type":    "one_api_error",
					},
				})
				c.Abort()
				return
			}
		} else {
			// Select a channel for the user
			var err error
			channel, err = model.GetRandomChannel()
			if err != nil {
				c.JSON(200, gin.H{
					"error": gin.H{
						"message": "无可用渠道",
						"type":    "one_api_error",
					},
				})
				c.Abort()
				return
			}
		}
		c.Set("channel", channel.Type)
		c.Set("channel_id", channel.Id)
		c.Set("channel_name", channel.Name)
		c.Request.Header.Set("Authorization", fmt.Sprintf("Bearer %s", channel.Key))
		if channel.Type == common.ChannelTypeCustom || channel.Type == common.ChannelTypeAzure {
			c.Set("base_url", channel.BaseURL)
			if channel.Type == common.ChannelTypeAzure {
				c.Set("api_version", channel.Other)
			}
		}
		c.Next()
	}
}
