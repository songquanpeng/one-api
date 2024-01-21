package controller

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"one-api/common/config"
	"one-api/common/helper"
	"one-api/common/logger"
	"one-api/relay/channel/openai"
	"one-api/relay/constant"
	"one-api/relay/controller"
	"one-api/relay/util"
	"strconv"
)

// https://platform.openai.com/docs/api-reference/chat

func Relay(c *gin.Context) {
	relayMode := constant.Path2RelayMode(c.Request.URL.Path)
	var err *openai.ErrorWithStatusCode
	switch relayMode {
	case constant.RelayModeImagesGenerations:
		err = controller.RelayImageHelper(c, relayMode)
	case constant.RelayModeAudioSpeech:
		fallthrough
	case constant.RelayModeAudioTranslation:
		fallthrough
	case constant.RelayModeAudioTranscription:
		err = controller.RelayAudioHelper(c, relayMode)
	default:
		err = controller.RelayTextHelper(c, relayMode)
	}
	if err != nil {
		requestId := c.GetString(logger.RequestIdKey)
		retryTimesStr := c.Query("retry")
		retryTimes, _ := strconv.Atoi(retryTimesStr)
		if retryTimesStr == "" {
			retryTimes = config.RetryTimes
		}
		if retryTimes > 0 {
			c.Redirect(http.StatusTemporaryRedirect, fmt.Sprintf("%s?retry=%d", c.Request.URL.Path, retryTimes-1))
		} else {
			if err.StatusCode == http.StatusTooManyRequests {
				err.Error.Message = "当前分组上游负载已饱和，请稍后再试"
			}
			err.Error.Message = helper.MessageWithRequestId(err.Error.Message, requestId)
			c.JSON(err.StatusCode, gin.H{
				"error": err.Error,
			})
		}
		channelId := c.GetInt("channel_id")
		logger.Error(c.Request.Context(), fmt.Sprintf("relay error (channel #%d): %s", channelId, err.Message))
		// https://platform.openai.com/docs/guides/error-codes/api-errors
		if util.ShouldDisableChannel(&err.Error, err.StatusCode) {
			channelId := c.GetInt("channel_id")
			channelName := c.GetString("channel_name")
			disableChannel(channelId, channelName, err.Message)
		}
	}
}

func RelayNotImplemented(c *gin.Context) {
	err := openai.Error{
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
	err := openai.Error{
		Message: fmt.Sprintf("Invalid URL (%s %s)", c.Request.Method, c.Request.URL.Path),
		Type:    "invalid_request_error",
		Param:   "",
		Code:    "",
	}
	c.JSON(http.StatusNotFound, gin.H{
		"error": err,
	})
}
