package controller

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/songquanpeng/one-api/common/audit"
	"github.com/songquanpeng/one-api/common/config"
	"github.com/songquanpeng/one-api/common/ctxkey"
	"github.com/songquanpeng/one-api/common/helper"
	"github.com/songquanpeng/one-api/common/logger"
	"github.com/songquanpeng/one-api/relay"
	"github.com/songquanpeng/one-api/relay/adaptor/openai"
	"github.com/songquanpeng/one-api/relay/apitype"
	"github.com/songquanpeng/one-api/relay/billing"
	"github.com/songquanpeng/one-api/relay/channeltype"
	"github.com/songquanpeng/one-api/relay/meta"
	"github.com/songquanpeng/one-api/relay/model"
)

func (rl *defaultRelay) RelayTextHelper(c *gin.Context) *model.ErrorWithStatusCode {
	meta := meta.GetByContext(c)
	// get & validate textRequest
	textRequest, err := getAndValidateTextRequest(c, meta.Mode)
	if err != nil {
		logger.Errorf(c, "getAndValidateTextRequest failed: %s", err.Error())
		return openai.ErrorWrapper(err, "invalid_text_request", http.StatusBadRequest)
	}
	meta.IsStream = textRequest.Stream

	// map model name
	var isModelMapped bool
	meta.OriginModelName = textRequest.Model
	textRequest.Model, isModelMapped = getMappedModelName(textRequest.Model, meta.ModelMapping)
	meta.ActualModelName = textRequest.Model
	// get model ratio & group ratio
	var (
		preConsumedQuota int64
		modelRatio       float64
		groupRatio       float64
		ratio            float64
	)
	if rl.Bookkeeper != nil {
		modelRatio = rl.ModelRatio(textRequest.Model)
		groupRatio = rl.GroupRation(meta.Group)
		ratio = modelRatio * groupRatio
		// pre-consume quota
		meta.PromptTokens = getPromptTokens(textRequest, meta.Mode)
		preConsumeQuota := getPreConsumedQuota(textRequest, meta.PromptTokens, ratio)
		consumedQuota, bizErr := rl.PreConsumeQuota(c, preConsumeQuota, meta.UserId, meta.TokenId)
		if bizErr != nil {
			logger.Warnf(c, "preConsumeQuota failed: %+v", *bizErr)
			return bizErr
		}
		preConsumedQuota = consumedQuota
	}
	adaptor := relay.GetAdaptor(meta.APIType)
	if adaptor == nil {
		return openai.ErrorWrapper(fmt.Errorf("invalid api type: %d", meta.APIType), "invalid_api_type", http.StatusBadRequest)
	}
	adaptor.Init(meta)

	// get request body
	var requestBody io.Reader
	if meta.APIType == apitype.OpenAI {
		// no need to convert request for openai
		shouldResetRequestBody := isModelMapped || meta.ChannelType == channeltype.Baichuan // frequency_penalty 0 is not acceptable for baichuan
		if shouldResetRequestBody {
			jsonStr, err := json.Marshal(textRequest)
			if err != nil {
				return openai.ErrorWrapper(err, "json_marshal_failed", http.StatusInternalServerError)
			}
			requestBody = bytes.NewBuffer(jsonStr)
		} else {
			requestBody = c.Request.Body
		}
	} else {
		convertedRequest, err := adaptor.ConvertRequest(c, meta.Mode, textRequest)
		if err != nil {
			return openai.ErrorWrapper(err, "convert_request_failed", http.StatusInternalServerError)
		}
		jsonData, err := json.Marshal(convertedRequest)
		if err != nil {
			return openai.ErrorWrapper(err, "json_marshal_failed", http.StatusInternalServerError)
		}
		logger.Debugf(c, "converted request: \n%s", string(jsonData))
		requestBody = bytes.NewBuffer(jsonData)
	}

	if config.UpstreamAuditEnabled {
		buf := bytes.Buffer{}
		requestBody = io.TeeReader(requestBody, &buf)
		defer func() {
			audit.Logger().
				WithField("stage", "upstream request").
				WithField("raw", audit.B64encode(buf.Bytes())).
				WithField("requestid", c.GetString(helper.RequestIdKey)).
				WithFields(meta.ToLogrusFields()).
				Info("upstream request")
		}()
	}

	// do request
	resp, err := adaptor.DoRequest(c, meta, requestBody)
	if err != nil {
		logger.Errorf(c, "DoRequest failed: %s", err.Error())
		return openai.ErrorWrapper(err, "do_request_failed", http.StatusInternalServerError)
	}
	if config.UpstreamAuditEnabled {
		buf := audit.CaptureHTTPResponseBody(resp)
		defer func() {
			audit.Logger().
				WithField("stage", "upstream response").
				WithField("raw", audit.B64encode(buf.Bytes())).
				WithField("requestid", c.GetString(helper.RequestIdKey)).
				WithFields(meta.ToLogrusFields()).
				Info("upstream response")
		}()
	}
	refund := func() {
		if rl.Bookkeeper != nil && preConsumedQuota > 0 {
			rl.RefundQuota(c, preConsumedQuota, meta.TokenId)
		}
	}
	if isErrorHappened(meta, resp) {
		refund()
		return RelayErrorHandler(resp)
	}

	// do response
	usage, respErr := adaptor.DoResponse(c, resp, meta)
	if respErr != nil {
		logger.Errorf(c, "respErr is not nil: %+v", respErr)
		refund()
		return respErr
	}
	// post-consume quota
	if rl.Bookkeeper != nil {
		// go postConsumeQuota(c, usage, meta, textRequest, ratio, preConsumedQuota, modelRatio, groupRatio)
		completionRatio := rl.ModelCompletionRatio(textRequest.Model)
		logContent := fmt.Sprintf("模型倍率 %.2f，分组倍率 %.2f，补全倍率 %.2f", modelRatio, groupRatio, completionRatio)
		consumeLog := &billing.ConsumeLog{
			UserId:           meta.UserId,
			ChannelId:        meta.ChannelId,
			ModelName:        textRequest.Model,
			TokenName:        c.GetString(ctxkey.TokenName),
			TokenId:          meta.TokenId,
			Quota:            usage.Quota(completionRatio, ratio),
			Content:          logContent,
			PromptTokens:     usage.PromptTokens,
			CompletionTokens: usage.CompletionTokens,
			PreConsumedQuota: preConsumedQuota,
		}
		rl.Bookkeeper.Consume(c, consumeLog)
	}
	return nil
}
