package controller

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/songquanpeng/one-api/common/ctxkey"
	"github.com/songquanpeng/one-api/common/logger"
	"github.com/songquanpeng/one-api/relay"
	"github.com/songquanpeng/one-api/relay/adaptor/openai"
	"github.com/songquanpeng/one-api/relay/billing"
	billingratio "github.com/songquanpeng/one-api/relay/billing/ratio"
	"github.com/songquanpeng/one-api/relay/channeltype"
	"github.com/songquanpeng/one-api/relay/meta"
	relaymodel "github.com/songquanpeng/one-api/relay/model"
)

func isWithinRange(element string, value int) bool {
	if _, ok := billingratio.ImageGenerationAmounts[element]; !ok {
		return false
	}
	min := billingratio.ImageGenerationAmounts[element][0]
	max := billingratio.ImageGenerationAmounts[element][1]
	return value >= min && value <= max
}

func (rl *defaultRelay) RelayImageHelper(c *gin.Context, relayMode int) *relaymodel.ErrorWithStatusCode {
	meta := meta.GetByContext(c)
	imageRequest, err := getImageRequest(c, meta.Mode)
	if err != nil {
		logger.Errorf(c, "getImageRequest failed: %s", err.Error())
		return openai.ErrorWrapper(err, "invalid_image_request", http.StatusBadRequest)
	}

	// map model name
	var (
		isModelMapped    bool
		preConsumeQuota  int64
		preConsumedQuota int64
		imageCostRatio   float64
		bizErr           *relaymodel.ErrorWithStatusCode
	)
	meta.OriginModelName = imageRequest.Model
	imageRequest.Model, isModelMapped = getMappedModelName(imageRequest.Model, meta.ModelMapping)
	meta.ActualModelName = imageRequest.Model

	// model validation
	bizErr = validateImageRequest(imageRequest, meta)
	if bizErr != nil {
		return bizErr
	}

	if rl.Bookkeeper != nil {
		imageCostRatio, err = getImageCostRatio(imageRequest)
		if err != nil {
			return openai.ErrorWrapper(err, "get_image_cost_ratio_failed", http.StatusInternalServerError)
		}
	}

	originModel := imageRequest.Model
	// Convert the original image model
	imageRequest.Model, _ = getMappedModelName(imageRequest.Model, billingratio.ImageOriginModelName)
	c.Set("response_format", imageRequest.ResponseFormat)

	var requestBody io.Reader
	if isModelMapped || meta.ChannelType == channeltype.Azure { // make Azure channel request body
		jsonStr, err := json.Marshal(imageRequest)
		if err != nil {
			return openai.ErrorWrapper(err, "marshal_image_request_failed", http.StatusInternalServerError)
		}
		requestBody = bytes.NewBuffer(jsonStr)
	} else {
		requestBody = c.Request.Body
	}

	adaptor := relay.GetAdaptor(meta.APIType)
	if adaptor == nil {
		return openai.ErrorWrapper(fmt.Errorf("invalid api type: %d", meta.APIType), "invalid_api_type", http.StatusBadRequest)
	}
	adaptor.Init(meta)

	switch meta.ChannelType {
	case channeltype.Ali:
		fallthrough
	case channeltype.Baidu:
		fallthrough
	case channeltype.Zhipu:
		finalRequest, err := adaptor.ConvertImageRequest(imageRequest)
		if err != nil {
			return openai.ErrorWrapper(err, "convert_image_request_failed", http.StatusInternalServerError)
		}
		jsonStr, err := json.Marshal(finalRequest)
		if err != nil {
			return openai.ErrorWrapper(err, "marshal_image_request_failed", http.StatusInternalServerError)
		}
		requestBody = bytes.NewBuffer(jsonStr)
	}

	if rl.Bookkeeper != nil {
		modelRatio := rl.ModelRatio(originModel)
		groupRatio := rl.GroupRation(meta.Group)
		ratio := modelRatio * groupRatio
		preConsumeQuota = int64(ratio*imageCostRatio*1000) * int64(imageRequest.N)
		preConsumedQuota, bizErr = rl.PreConsumeQuota(c, preConsumeQuota, meta.UserId, meta.TokenId)
		if bizErr != nil {
			logger.Warnf(c, "preConsumeQuota failed: %+v", *bizErr)
			return bizErr
		}
	}

	refund := func() {
		if rl.Bookkeeper != nil && preConsumedQuota > 0 {
			rl.RefundQuota(c, preConsumedQuota, meta.TokenId)
		}
	}
	// do request
	resp, err := adaptor.DoRequest(c, meta, requestBody)
	if err != nil {
		logger.Errorf(c, "DoRequest failed: %s", err.Error())
		refund()
		return openai.ErrorWrapper(err, "do_request_failed", http.StatusInternalServerError)
	}

	defer func(ctx context.Context) {
		if resp != nil && resp.StatusCode != http.StatusOK {
			return
		}
		if rl.Bookkeeper == nil {
			return
		}
		modelRatio := rl.ModelRatio(originModel)
		groupRatio := rl.GroupRation(meta.Group)
		ratio := modelRatio * groupRatio
		consumedQuota := int64(ratio*imageCostRatio*1000) * int64(imageRequest.N)

		if consumedQuota != 0 {
			tokenName := c.GetString(ctxkey.TokenName)
			logContent := fmt.Sprintf("模型倍率 %.2f，分组倍率 %.2f", modelRatio, groupRatio)
			consumeLog := &billing.ConsumeLog{
				UserId:           meta.UserId,
				ChannelId:        meta.ChannelId,
				ModelName:        imageRequest.Model,
				TokenName:        tokenName,
				TokenId:          meta.TokenId,
				Quota:            consumedQuota,
				Content:          logContent,
				PromptTokens:     0,
				CompletionTokens: 0,
				PreConsumedQuota: preConsumedQuota,
			}
			rl.Bookkeeper.Consume(c, consumeLog)
		}
	}(c)

	// do response
	_, respErr := adaptor.DoResponse(c, resp, meta)
	if respErr != nil {
		logger.Errorf(c, "respErr is not nil: %+v", respErr)
		return respErr
	}

	return nil
}
