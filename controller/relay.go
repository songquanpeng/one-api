package controller

import (
	"fmt"
	"net/http"
	"one-api/common"
	"one-api/types"
	"strconv"

	"github.com/gin-gonic/gin"
)

func RelayNotImplemented(c *gin.Context) {
	err := types.OpenAIError{
		Message: "API not implemented",
		Type:    "one_api_error",
		Param:   "",
		Code:    "api_not_implemented",
	}
	c.JSON(http.StatusNotImplemented, gin.H{
		"error": err,
	})
}

func RelayNotFound(c *gin.Context) {
	err := types.OpenAIError{
		Message: fmt.Sprintf("Invalid URL (%s %s)", c.Request.Method, c.Request.URL.Path),
		Type:    "invalid_request_error",
		Param:   "",
		Code:    "",
	}
	c.JSON(http.StatusNotFound, gin.H{
		"error": err,
	})
}

func errorHelper(c *gin.Context, err *types.OpenAIErrorWithStatusCode) {
	requestId := c.GetString(common.RequestIdKey)
	retryTimesStr := c.Query("retry")
	retryTimes, _ := strconv.Atoi(retryTimesStr)
	if retryTimesStr == "" {
		retryTimes = common.RetryTimes
	}
	if retryTimes > 0 {
		c.Redirect(http.StatusTemporaryRedirect, fmt.Sprintf("%s?retry=%d", c.Request.URL.Path, retryTimes-1))
	} else {
		if err.StatusCode == http.StatusTooManyRequests {
			err.OpenAIError.Message = "当前分组上游负载已饱和，请稍后再试"
		}
		err.OpenAIError.Message = common.MessageWithRequestId(err.OpenAIError.Message, requestId)
		c.JSON(err.StatusCode, gin.H{
			"error": err.OpenAIError,
		})
	}
	channelId := c.GetInt("channel_id")
	common.LogError(c.Request.Context(), fmt.Sprintf("relay error (channel #%d): %s", channelId, err.Message))
	// https://platform.openai.com/docs/guides/error-codes/api-errors
	if shouldDisableChannel(&err.OpenAIError, err.StatusCode) {
		channelId := c.GetInt("channel_id")
		channelName := c.GetString("channel_name")
		disableChannel(channelId, channelName, err.Message)
	}
}
