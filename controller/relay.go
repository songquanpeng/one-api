package controller

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"io"
	"net/http"
	"one-api/common"
)

func Relay(c *gin.Context) {
	channelType := c.GetInt("channel")
	baseURL := common.ChannelBaseURLs[channelType]
	if channelType == common.ChannelTypeCustom {
		baseURL = c.GetString("base_url")
	}
	req, err := http.NewRequest(c.Request.Method, fmt.Sprintf("%s%s", baseURL, c.Request.URL.String()), c.Request.Body)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"error": gin.H{
				"message": err.Error(),
				"type":    "one_api_error",
			},
		})
		return
	}
	//req.Header = c.Request.Header.Clone()
	// Fix HTTP Decompression failed
	// https://github.com/stoplightio/prism/issues/1064#issuecomment-824682360
	//req.Header.Del("Accept-Encoding")
	req.Header.Set("Authorization", c.Request.Header.Get("Authorization"))
	req.Header.Set("Content-Type", c.Request.Header.Get("Content-Type"))
	acceptHeader := c.Request.Header.Get("Accept")
	if acceptHeader != "" {
		req.Header.Set("Accept", acceptHeader)
	}
	connectionHeader := c.Request.Header.Get("Connection")
	if connectionHeader != "" {
		req.Header.Set("Connection", connectionHeader)
	}
	lastEventIDHeader := c.Request.Header.Get("Last-Event-ID")
	if lastEventIDHeader != "" {
		req.Header.Set("Last-Event-ID", lastEventIDHeader)
	}
	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"error": gin.H{
				"message": err.Error(),
				"type":    "one_api_error",
			},
		})
		return
	}
	for k, v := range resp.Header {
		c.Writer.Header().Set(k, v[0])
	}
	_, err = io.Copy(c.Writer, resp.Body)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"error": gin.H{
				"message": err.Error(),
				"type":    "one_api_error",
			},
		})
		return
	}
}
