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
	req.Header = c.Request.Header.Clone()
	// Fix HTTP Decompression failed
	// https://github.com/stoplightio/prism/issues/1064#issuecomment-824682360
	req.Header.Del("Accept-Encoding")
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
