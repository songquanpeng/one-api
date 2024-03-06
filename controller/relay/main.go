package relay

import (
	"fmt"
	"net/http"
	"one-api/common"
	"one-api/model"
	"one-api/types"

	"github.com/gin-gonic/gin"
)

func Relay(c *gin.Context) {
	relay := Path2Relay(c, c.Request.URL.Path)
	if relay == nil {
		common.AbortWithMessage(c, http.StatusNotFound, "Not Found")
		return
	}

	if err := relay.setRequest(); err != nil {
		common.AbortWithMessage(c, http.StatusBadRequest, err.Error())
		return
	}

	if err := relay.setProvider(relay.getOriginalModel()); err != nil {
		common.AbortWithMessage(c, http.StatusServiceUnavailable, err.Error())
		return
	}

	apiErr, done := RelayHandler(relay)
	if apiErr == nil {
		return
	}

	channel := relay.getProvider().GetChannel()
	go processChannelRelayError(c.Request.Context(), channel.Id, channel.Name, apiErr)

	retryTimes := common.RetryTimes
	if done || !shouldRetry(c, apiErr.StatusCode) {
		common.LogError(c.Request.Context(), fmt.Sprintf("relay error happen, status code is %d, won't retry in this case", apiErr.StatusCode))
		retryTimes = 0
	}

	for i := retryTimes; i > 0; i-- {
		// 冻结通道
		model.ChannelGroup.Cooldowns(channel.Id)
		if err := relay.setProvider(relay.getOriginalModel()); err != nil {
			continue
		}

		channel = relay.getProvider().GetChannel()
		common.LogError(c.Request.Context(), fmt.Sprintf("using channel #%d(%s) to retry (remain times %d)", channel.Id, channel.Name, i))
		apiErr, done = RelayHandler(relay)
		if apiErr == nil {
			return
		}
		go processChannelRelayError(c.Request.Context(), channel.Id, channel.Name, apiErr)
		if done || !shouldRetry(c, apiErr.StatusCode) {
			break
		}
	}

	if apiErr != nil {
		requestId := c.GetString(common.RequestIdKey)
		if apiErr.StatusCode == http.StatusTooManyRequests {
			apiErr.OpenAIError.Message = "当前分组上游负载已饱和，请稍后再试"
		}
		apiErr.OpenAIError.Message = common.MessageWithRequestId(apiErr.OpenAIError.Message, requestId)
		c.JSON(apiErr.StatusCode, gin.H{
			"error": apiErr.OpenAIError,
		})

	}
}

func RelayHandler(relay RelayBaseInterface) (err *types.OpenAIErrorWithStatusCode, done bool) {
	promptTokens, tonkeErr := relay.getPromptTokens()
	if tonkeErr != nil {
		err = common.ErrorWrapper(tonkeErr, "token_error", http.StatusBadRequest)
		done = true
		return
	}

	usage := &types.Usage{
		PromptTokens: promptTokens,
	}

	relay.getProvider().SetUsage(usage)

	var quotaInfo *QuotaInfo
	quotaInfo, err = generateQuotaInfo(relay.getContext(), relay.getModelName(), promptTokens)
	if err != nil {
		done = true
		return
	}

	err, done = relay.send()

	if err != nil {
		quotaInfo.undo(relay.getContext())
		return
	}

	quotaInfo.consume(relay.getContext(), usage)
	return
}
