package controller

import (
	"bufio"
	"fmt"
	"github.com/gin-gonic/gin"
	"io"
	"net/http"
	"one-api/common"
	"one-api/model"
	"strings"
)

func Relay(c *gin.Context) {
	channelType := c.GetInt("channel")
	tokenId := c.GetInt("token_id")
	isUnlimitedQuota := c.GetBool("unlimited_quota")
	baseURL := common.ChannelBaseURLs[channelType]
	if channelType == common.ChannelTypeCustom {
		baseURL = c.GetString("base_url")
	}
	requestURL := c.Request.URL.String()
	req, err := http.NewRequest(c.Request.Method, fmt.Sprintf("%s%s", baseURL, requestURL), c.Request.Body)
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
	req.Header.Set("Accept", c.Request.Header.Get("Accept"))
	req.Header.Set("Connection", c.Request.Header.Get("Connection"))
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

	defer func() {
		err := req.Body.Close()
		if err != nil {
			common.SysError("Error closing request body: " + err.Error())
		}
		if !isUnlimitedQuota && requestURL == "/v1/chat/completions" {
			err := model.DecreaseTokenRemainQuotaById(tokenId)
			if err != nil {
				common.SysError("Error decreasing token remain times: " + err.Error())
			}
		}
	}()
	isStream := resp.Header.Get("Content-Type") == "text/event-stream"
	if isStream {
		scanner := bufio.NewScanner(resp.Body)
		scanner.Split(func(data []byte, atEOF bool) (advance int, token []byte, err error) {
			if atEOF && len(data) == 0 {
				return 0, nil, nil
			}

			if i := strings.Index(string(data), "\n\n"); i >= 0 {
				return i + 2, data[0:i], nil
			}

			if atEOF {
				return len(data), data, nil
			}

			return 0, nil, nil
		})
		dataChan := make(chan string)
		stopChan := make(chan bool)
		go func() {
			for scanner.Scan() {
				data := scanner.Text()
				dataChan <- data
			}
			stopChan <- true
		}()
		c.Writer.Header().Set("Content-Type", "text/event-stream")
		c.Writer.Header().Set("Cache-Control", "no-cache")
		c.Writer.Header().Set("Connection", "keep-alive")
		c.Writer.Header().Set("Transfer-Encoding", "chunked")
		c.Stream(func(w io.Writer) bool {
			select {
			case data := <-dataChan:
				c.Render(-1, common.CustomEvent{Data: data})
				return true
			case <-stopChan:
				return false
			}
		})
		return
	} else {
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
}
