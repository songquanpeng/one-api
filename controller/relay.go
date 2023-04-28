package controller

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"io"
	"net/http"
	"one-api/common"
	"one-api/model"
	"strings"
)

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type TextRequest struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
	Prompt   string    `json:"prompt"`
	//Stream   bool      `json:"stream"`
}

type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

type TextResponse struct {
	Usage `json:"usage"`
}

type StreamResponse struct {
	Choices []struct {
		Delta struct {
			Content string `json:"content"`
		} `json:"delta"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
}

func Relay(c *gin.Context) {
	err := relayHelper(c)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"error": gin.H{
				"message": err.Error(),
				"type":    "one_api_error",
			},
		})
	}
}

func relayHelper(c *gin.Context) error {
	channelType := c.GetInt("channel")
	tokenId := c.GetInt("token_id")
	consumeQuota := c.GetBool("consume_quota")
	baseURL := common.ChannelBaseURLs[channelType]
	if channelType == common.ChannelTypeCustom {
		baseURL = c.GetString("base_url")
	}
	if consumeQuota {
		requestBody, err := io.ReadAll(c.Request.Body)
		if err != nil {
			return err
		}
		err = c.Request.Body.Close()
		if err != nil {
			return err
		}
		var textRequest TextRequest
		err = json.Unmarshal(requestBody, &textRequest)
		if err != nil {
			return err
		}
		// Reset request body
		c.Request.Body = io.NopCloser(bytes.NewBuffer(requestBody))
	}
	requestURL := c.Request.URL.String()
	req, err := http.NewRequest(c.Request.Method, fmt.Sprintf("%s%s", baseURL, requestURL), c.Request.Body)
	if err != nil {
		return err
	}
	err = c.Request.Body.Close()
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", c.Request.Header.Get("Authorization"))
	req.Header.Set("Content-Type", c.Request.Header.Get("Content-Type"))
	req.Header.Set("Accept", c.Request.Header.Get("Accept"))
	req.Header.Set("Connection", c.Request.Header.Get("Connection"))
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	err = req.Body.Close()
	if err != nil {
		return err
	}

	var textResponse TextResponse
	isStream := resp.Header.Get("Content-Type") == "text/event-stream"
	var streamResponseText string

	defer func() {
		if consumeQuota {
			quota := 0
			if isStream {
				quota = int(float64(len(streamResponseText)) * common.BytesNumber2Quota)
			} else {
				quota = textResponse.Usage.TotalTokens
			}
			err := model.ConsumeTokenQuota(tokenId, quota)
			if err != nil {
				common.SysError("Error consuming token remain quota: " + err.Error())
			}
		}
	}()

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
				data = data[6:]
				if data != "[DONE]" {
					var streamResponse StreamResponse
					err = json.Unmarshal([]byte(data), &streamResponse)
					if err != nil {
						common.SysError("Error unmarshalling stream response: " + err.Error())
						return
					}
					for _, choice := range streamResponse.Choices {
						streamResponseText += choice.Delta.Content
					}
				}
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
		err = resp.Body.Close()
		if err != nil {
			return err
		}
		return nil
	} else {
		for k, v := range resp.Header {
			c.Writer.Header().Set(k, v[0])
		}
		if consumeQuota {
			responseBody, err := io.ReadAll(resp.Body)
			if err != nil {
				return err
			}
			err = resp.Body.Close()
			if err != nil {
				return err
			}
			err = json.Unmarshal(responseBody, &textResponse)
			if err != nil {
				return err
			}
			// Reset response body
			resp.Body = io.NopCloser(bytes.NewBuffer(responseBody))
		}
		_, err = io.Copy(c.Writer, resp.Body)
		if err != nil {
			return err
		}
		err = resp.Body.Close()
		if err != nil {
			return err
		}
		return nil
	}
}

func RelayNotImplemented(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"error": gin.H{
			"message": "Not Implemented",
			"type":    "one_api_error",
		},
	})
}
