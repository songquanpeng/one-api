package controller

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/songquanpeng/one-api/common"
	"github.com/songquanpeng/one-api/common/audit"
	"github.com/songquanpeng/one-api/common/config"
	"github.com/songquanpeng/one-api/common/ctxkey"
	"github.com/songquanpeng/one-api/common/helper"
	"github.com/songquanpeng/one-api/common/logger"
	"github.com/songquanpeng/one-api/middleware"
	dbmodel "github.com/songquanpeng/one-api/model"
	"github.com/songquanpeng/one-api/monitor"
	"github.com/songquanpeng/one-api/relay/controller"
	"github.com/songquanpeng/one-api/relay/meta"
	"github.com/songquanpeng/one-api/relay/model"
	"github.com/songquanpeng/one-api/relay/relaymode"
)

// https://platform.openai.com/docs/api-reference/chat

type Options struct {
	Debug         bool
	EnableMonitor bool
	EnableBilling bool
}

type RelayController struct {
	opts Options
	controller.RelayInstance
	monitor.MonitorInstance
}

func NewRelayController(opts Options) *RelayController {
	ctrl := &RelayController{
		opts: opts,
	}
	ctrl.RelayInstance = controller.NewRelayInstance(controller.Options{
		EnableBilling: opts.EnableBilling,
	})
	if opts.EnableMonitor {
		ctrl.MonitorInstance = monitor.NewMonitorInstance()
	}
	return ctrl
}

func (ctrl *RelayController) relayHelper(c *gin.Context, relayMode int) *model.ErrorWithStatusCode {
	if config.ClientAuditEnabled {
		buf := audit.CaptureResponseBody(c)
		m := meta.GetByContext(c)
		defer func() {
			audit.Logger().
				WithField("raw", audit.B64encode(buf.Bytes())).
				WithField("parsed", audit.ParseOPENAIStreamResponse(buf)).
				WithField("requestid", c.GetString(helper.RequestIdKey)).
				WithFields(m.ToLogrusFields()).
				Info("client response")
		}()
	}
	var err *model.ErrorWithStatusCode
	switch relayMode {
	case relaymode.ImagesGenerations:
		err = ctrl.RelayImageHelper(c, relayMode)
	case relaymode.AudioSpeech:
		fallthrough
	case relaymode.AudioTranslation:
		fallthrough
	case relaymode.AudioTranscription:
		err = ctrl.RelayAudioHelper(c, relayMode)
	default:
		err = ctrl.RelayTextHelper(c)
	}
	return err
}

func (ctrl *RelayController) Relay(c *gin.Context) {
	ctx := c.Request.Context()
	relayMode := relaymode.GetByPath(c.Request.URL.Path)
	if config.DebugEnabled {
		requestBody, _ := common.GetRequestBody(c)
		logger.Debugf(ctx, "request body: %s", string(requestBody))
	}
	if config.ClientAuditEnabled {
		requestBody, _ := common.GetRequestBody(c)
		m := meta.GetByContext(c)
		audit.Logger().
			WithField("raw", audit.B64encode(requestBody)).
			WithField("requestid", c.GetString(helper.RequestIdKey)).
			WithFields(m.ToLogrusFields()).
			Info("client request")
	}
	channelId := c.GetInt(ctxkey.ChannelId)
	bizErr := ctrl.relayHelper(c, relayMode)
	if bizErr == nil {
		if ctrl.MonitorInstance != nil {
			ctrl.Emit(channelId, true)
		}
		return
	}
	lastFailedChannelId := channelId
	channelName := c.GetString(ctxkey.ChannelName)
	group := c.GetString(ctxkey.Group)
	originalModel := c.GetString(ctxkey.OriginalModel)
	userId := c.GetInt(ctxkey.Id)
	go ctrl.processChannelRelayError(ctx, userId, channelId, channelName, bizErr)
	requestId := c.GetString(helper.RequestIdKey)
	retryTimes := config.RetryTimes
	if !shouldRetry(c, bizErr.StatusCode) {
		logger.Errorf(ctx, "relay error happen, status code is %d, won't retry in this case", bizErr.StatusCode)
		retryTimes = 0
	}
	for i := retryTimes; i > 0; i-- {
		channel, err := dbmodel.CacheGetRandomSatisfiedChannel(group, originalModel, i != retryTimes)
		if err != nil {
			logger.Errorf(ctx, "CacheGetRandomSatisfiedChannel failed: %+v", err)
			break
		}
		logger.Infof(ctx, "using channel #%d to retry (remain times %d)", channel.Id, i)
		if channel.Id == lastFailedChannelId {
			continue
		}
		middleware.SetupContextForSelectedChannel(c, channel, originalModel)
		requestBody, err := common.GetRequestBody(c)
		if err != nil {
			logger.Errorf(ctx, "GetRequestBody failed: %+v", err)
			break
		}
		c.Request.Body = io.NopCloser(bytes.NewBuffer(requestBody))
		bizErr = ctrl.relayHelper(c, relayMode)
		if bizErr == nil {
			return
		}
		channelId := c.GetInt(ctxkey.ChannelId)
		lastFailedChannelId = channelId
		channelName := c.GetString(ctxkey.ChannelName)
		go ctrl.processChannelRelayError(ctx, userId, channelId, channelName, bizErr)
	}
	if bizErr != nil {
		if bizErr.StatusCode == http.StatusTooManyRequests {
			bizErr.Error.Message = "当前分组上游负载已饱和，请稍后再试"
		}
		bizErr.Error.Message = helper.MessageWithRequestId(bizErr.Error.Message, requestId)
		c.JSON(bizErr.StatusCode, gin.H{
			"error": bizErr.Error,
		})
	}
}

func shouldRetry(c *gin.Context, statusCode int) bool {
	if _, ok := c.Get(ctxkey.SpecificChannelId); ok {
		return false
	}
	if statusCode == http.StatusTooManyRequests {
		return true
	}
	if statusCode/100 == 5 {
		return true
	}
	if statusCode == http.StatusBadRequest {
		return false
	}
	if statusCode/100 == 2 {
		return false
	}
	return true
}

func (ctrl *RelayController) processChannelRelayError(ctx context.Context, userId int, channelId int, channelName string, err *model.ErrorWithStatusCode) {
	if ctrl.MonitorInstance == nil {
		return
	}
	logger.Errorf(ctx, "relay error (channel id %d, user id: %d): %s", channelId, userId, err.Message)
	// https://platform.openai.com/docs/guides/error-codes/api-errors
	if ctrl.ShouldDisableChannel(&err.Error, err.StatusCode) {
		ctrl.DisableChannel(channelId, channelName, err.Message)
	} else {
		ctrl.Emit(channelId, false)
	}
}

func RelayNotImplemented(c *gin.Context) {
	err := model.Error{
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
	err := model.Error{
		Message: fmt.Sprintf("Invalid URL (%s %s)", c.Request.Method, c.Request.URL.Path),
		Type:    "invalid_request_error",
		Param:   "",
		Code:    "",
	}
	c.JSON(http.StatusNotFound, gin.H{
		"error": err,
	})
}
