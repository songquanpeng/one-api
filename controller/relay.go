package controller

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/gin-gonic/gin"
	"io"
	"net/http"
	"one-api/common"
	"strings"
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
	defer resp.Body.Close()
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
	for k, v := range resp.Header {
		c.Writer.Header().Set(k, v[0])
	}
	c.Stream(func(w io.Writer) bool {
		select {
		case data := <-dataChan:
			//fmt.Println(data)
			//c.Data(http.StatusOK, "text/event-stream", []byte(data))
			//c.Render(-1, common.Event{Data: data})
			//c.SSEvent("", data)
			//w.Write([]byte(data))
			//w.(http.Flusher).Flush()
			//c.Writer.Write(append([]byte(data), []byte("\n\n")...))
			outputBytes := bytes.NewBufferString(data)
			w.Write(outputBytes.Bytes())
			if strings.HasPrefix(data, "data: ") {
				w.Write([]byte("\n\n"))
			}
			//w.Write(append(outputBytes.Bytes(), []byte("\n\n")...))
			w.(http.Flusher).Flush()
			//fmt.Println(data)
			return true
		case <-stopChan:
			return false
		}
	})
	return
}
