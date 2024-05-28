// Author: Calcium-Ion
// GitHub: https://github.com/Calcium-Ion/new-api
// Path: controller/relay.go
package midjourney

import (
	"fmt"
	"net/http"
	"one-api/common/logger"
	provider "one-api/providers/midjourney"
	"strings"

	"github.com/gin-gonic/gin"
)

func RelayMidjourney(c *gin.Context) {
	relayMode := Path2RelayModeMidjourney(c.Request.URL.Path)
	var err *provider.MidjourneyResponse
	switch relayMode {
	case provider.RelayModeMidjourneyNotify:
		err = RelayMidjourneyNotify(c)
	case provider.RelayModeMidjourneyTaskFetch, provider.RelayModeMidjourneyTaskFetchByCondition:
		err = RelayMidjourneyTask(c, relayMode)
	case provider.RelayModeMidjourneyTaskImageSeed:
		err = RelayMidjourneyTaskImageSeed(c)
	case provider.RelayModeMidjourneySwapFace:
		err = RelaySwapFace(c)
	default:
		err = RelayMidjourneySubmit(c, relayMode)
	}

	if err != nil {
		statusCode := http.StatusBadRequest
		if err.Code == 30 {
			err.Result = "当前分组负载已饱和，请稍后再试，或升级账户以提升服务质量。"
			statusCode = http.StatusTooManyRequests
		}

		typeMsg := "upstream_error"
		if err.Type != "" {
			typeMsg = err.Type
		}
		c.JSON(statusCode, gin.H{
			"description": fmt.Sprintf("%s %s", err.Description, err.Result),
			"type":        typeMsg,
			"code":        err.Code,
		})
		channelId := c.GetInt("channel_id")
		logger.SysError(fmt.Sprintf("relay error (channel #%d): %s", channelId, fmt.Sprintf("%s %s", err.Description, err.Result)))
	}
}

func MidjourneyErrorFromInternal(code int, description string) *provider.MidjourneyResponse {
	return &provider.MidjourneyResponse{
		Code:        code,
		Description: description,
		Type:        "internal_error",
	}
}

func Path2RelayModeMidjourney(path string) int {
	relayMode := provider.RelayModeUnknown
	if strings.HasSuffix(path, "/mj/submit/action") {
		// midjourney plus
		relayMode = provider.RelayModeMidjourneyAction
	} else if strings.HasSuffix(path, "/mj/submit/modal") {
		// midjourney plus
		relayMode = provider.RelayModeMidjourneyModal
	} else if strings.HasSuffix(path, "/mj/submit/shorten") {
		// midjourney plus
		relayMode = provider.RelayModeMidjourneyShorten
	} else if strings.HasSuffix(path, "/mj/insight-face/swap") {
		// midjourney plus
		relayMode = provider.RelayModeMidjourneySwapFace
	} else if strings.HasSuffix(path, "/mj/submit/imagine") {
		relayMode = provider.RelayModeMidjourneyImagine
	} else if strings.HasSuffix(path, "/mj/submit/blend") {
		relayMode = provider.RelayModeMidjourneyBlend
	} else if strings.HasSuffix(path, "/mj/submit/describe") {
		relayMode = provider.RelayModeMidjourneyDescribe
	} else if strings.HasSuffix(path, "/mj/notify") {
		relayMode = provider.RelayModeMidjourneyNotify
	} else if strings.HasSuffix(path, "/mj/submit/change") {
		relayMode = provider.RelayModeMidjourneyChange
	} else if strings.HasSuffix(path, "/mj/submit/simple-change") {
		relayMode = provider.RelayModeMidjourneyChange
	} else if strings.HasSuffix(path, "/fetch") {
		relayMode = provider.RelayModeMidjourneyTaskFetch
	} else if strings.HasSuffix(path, "/image-seed") {
		relayMode = provider.RelayModeMidjourneyTaskImageSeed
	} else if strings.HasSuffix(path, "/list-by-condition") {
		relayMode = provider.RelayModeMidjourneyTaskFetchByCondition
	}
	return relayMode
}
