package middleware

import (
	"bytes"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"io"
	"net/http"
	"one-api/common"
	"strconv"
	"time"
)

type OpenAIErrorWithStatusCode struct {
	OpenAIError
	StatusCode int `json:"status_code"`
}

type OpenAIError struct {
	Message string `json:"message"`
	Type    string `json:"type"`
	Param   string `json:"param"`
	Code    any    `json:"code"`
}

func RetryHandler(group *gin.RouterGroup) gin.HandlerFunc {
	var retryHandler gin.HandlerFunc
	// 获取RetryHandler在当前HandlersChain的位置
	index := len(group.Handlers) + 1
	retryHandler = func(c *gin.Context) {
		// Backup request
		hasBody := c.Request.ContentLength > 0
		backupHeader := c.Request.Header.Clone()
		var backupBody []byte
		var err error
		if hasBody {
			backupBody, err = io.ReadAll(c.Request.Body)
			if err != nil {
				abortWithMessage(c, http.StatusBadRequest, "Invalid request")
				return
			}
			_ = c.Request.Body.Close()
			c.Request.Body = io.NopCloser(bytes.NewBuffer(backupBody))
		}

		// 获取 retryHandler 后续的中间件
		// Get next handlers
		nextHandlers := group.Handlers[index:]

		// 加入Relay处理函数 c.Handler() => c.handlers.Last() => controller.Relay
		// Add Relay handler
		nextHandlers = append(nextHandlers, c.Handler())

		// Retry
		maxRetryStr := c.Query("retry")
		maxRetry, err := strconv.Atoi(maxRetryStr)
		if err != nil || maxRetryStr == "" || maxRetry < 1 || maxRetry > common.RetryTimes {
			maxRetry = common.Max(common.RetryTimes+1, 1)
		}
		retryDelay := time.Duration(common.Max(common.RetryInterval, 0)) * time.Millisecond
		for i := 0; i < maxRetry; i++ {
			if i == 0 {
				// 第一次请求, 直接执行使用c.Next()调用后续中间件, 防止直接使用handler 内部调用c.Next() 导致重复执行
				// First request, execute next middleware
				c.Next()
			} else {
				// Clear errors to avoid confusion in next middleware
				c.Errors = c.Errors[:0]
				// 重试, 恢复请求头和请求体, 并执行后续中间件
				// Retry, restore request and execute next middleware
				c.Request.Header = backupHeader.Clone()
				if hasBody {
					c.Request.Body = io.NopCloser(bytes.NewBuffer(backupBody))
				}
				for _, handler := range nextHandlers {
					handler(c)
				}
			}

			// If no errors, return
			if len(c.Errors) == 0 {
				return
			}
			// c.index 指向 AbortIndex 可以防止出错时重复执行后续中间件
			c.Abort()
			// If errors, retry after delay
			time.Sleep(retryDelay)
		}
		var openaiErr *OpenAIErrorWithStatusCode
		err = json.Unmarshal([]byte(c.Errors.Last().Error()), &openaiErr)
		if err != nil {
			abortWithMessage(c, http.StatusInternalServerError, c.Errors.Last().Error())
			return
		}
		c.JSON(openaiErr.StatusCode, gin.H{
			"error": openaiErr.OpenAIError,
		})
	}
	group.Handlers = append(group.Handlers, retryHandler)
	return retryHandler
}
