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
	host := common.ChannelHosts[channelType]
	req, err := http.NewRequest(c.Request.Method, fmt.Sprintf("%s%s", host, c.Request.URL.String()), c.Request.Body)
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
	//body, err := io.ReadAll(resp.Body)
	//_, err = c.Writer.Write(body)
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
