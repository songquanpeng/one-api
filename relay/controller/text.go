package controller

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/songquanpeng/one-api/common"
	"github.com/songquanpeng/one-api/common/logger"
	"github.com/songquanpeng/one-api/relay/channel/openai"
	"github.com/songquanpeng/one-api/relay/constant"
	"github.com/songquanpeng/one-api/relay/util"
	"net/http"
	"strings"
)

func RelayTextHelper(c *gin.Context) *openai.ErrorWithStatusCode {
	ctx := c.Request.Context()
	meta := util.GetRelayMeta(c)
	// get & validate textRequest
	textRequest, err := getAndValidateTextRequest(c, meta.Mode)
	if err != nil {
		logger.Errorf(ctx, "getAndValidateTextRequest failed: %s", err.Error())
		return openai.ErrorWrapper(err, "invalid_text_request", http.StatusBadRequest)
	}
	// map model name
	var isModelMapped bool
	textRequest.Model, isModelMapped = util.GetMappedModelName(textRequest.Model, meta.ModelMapping)
	// get model ratio & group ratio
	modelRatio := common.GetModelRatio(textRequest.Model)
	groupRatio := common.GetGroupRatio(meta.Group)
	ratio := modelRatio * groupRatio
	// pre-consume quota
	promptTokens := getPromptTokens(textRequest, meta.Mode)
	preConsumedQuota, bizErr := preConsumeQuota(ctx, textRequest, promptTokens, ratio, meta)
	if bizErr != nil {
		logger.Warnf(ctx, "preConsumeQuota failed: %+v", *bizErr)
		return bizErr
	}

	// get request body
	requestBody, err := GetRequestBody(c, *textRequest, isModelMapped, meta.APIType, meta.Mode)
	if err != nil {
		return openai.ErrorWrapper(err, "get_request_body_failed", http.StatusInternalServerError)
	}
	// do request
	var resp *http.Response
	isStream := textRequest.Stream
	if meta.APIType != constant.APITypeXunfei { // cause xunfei use websocket
		fullRequestURL, err := GetRequestURL(c.Request.URL.String(), meta, textRequest)
		if err != nil {
			logger.Error(ctx, fmt.Sprintf("util.GetRequestURL failed: %s", err.Error()))
			return openai.ErrorWrapper(fmt.Errorf("util.GetRequestURL failed"), "get_request_url_failed", http.StatusInternalServerError)
		}

		resp, err = doRequest(ctx, c, meta, isStream, fullRequestURL, requestBody)
		if err != nil {
			logger.Errorf(ctx, "doRequest failed: %s", err.Error())
			return openai.ErrorWrapper(err, "do_request_failed", http.StatusInternalServerError)
		}
		isStream = isStream || strings.HasPrefix(resp.Header.Get("Content-Type"), "text/event-stream")

		if resp.StatusCode != http.StatusOK {
			util.ReturnPreConsumedQuota(ctx, preConsumedQuota, meta.TokenId)
			return util.RelayErrorHandler(resp)
		}
	}
	// do response
	usage, respErr := DoResponse(c, textRequest, resp, meta.Mode, meta.APIType, isStream, promptTokens)
	if respErr != nil {
		logger.Errorf(ctx, "respErr is not nil: %+v", respErr)
		util.ReturnPreConsumedQuota(ctx, preConsumedQuota, meta.TokenId)
		return respErr
	}
	// post-consume quota
	go postConsumeQuota(ctx, usage, meta, textRequest, ratio, preConsumedQuota, modelRatio, groupRatio)
	return nil
}
