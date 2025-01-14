package vertexai

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"github.com/songquanpeng/one-api/common/ctxkey"
	"github.com/songquanpeng/one-api/relay/adaptor"
	"github.com/songquanpeng/one-api/relay/adaptor/anthropic"
	"github.com/songquanpeng/one-api/relay/billing/ratio"
	"github.com/songquanpeng/one-api/relay/meta"
	"github.com/songquanpeng/one-api/relay/model"
)

// https://cloud.google.com/vertex-ai/generative-ai/pricing?hl=zh-cn#claude-models
var RatioMap = map[string]ratio.Ratio{
	"claude-3-haiku@20240307":       {Input: 0.25 * ratio.MILLI_USD, Output: 1.25 * ratio.MILLI_USD},
	"claude-3-sonnet@20240229":      {Input: 3 * ratio.MILLI_USD, Output: 15 * ratio.MILLI_USD},
	"claude-3-opus@20240229":        {Input: 15 * ratio.MILLI_USD, Output: 75 * ratio.MILLI_USD},
	"claude-3-5-sonnet@20240620":    {Input: 3 * ratio.MILLI_USD, Output: 15 * ratio.MILLI_USD},
	"claude-3-5-sonnet-v2@20241022": {Input: 3 * ratio.MILLI_USD, Output: 15 * ratio.MILLI_USD},
	"claude-3-5-haiku@20241022":     {Input: 0.80 * ratio.MILLI_USD, Output: 4 * ratio.MILLI_USD},
}

const anthropicVersion = "vertex-2023-10-16"

type Adaptor struct {
}

func (a *Adaptor) ConvertRequest(c *gin.Context, relayMode int, request *model.GeneralOpenAIRequest) (any, error) {
	if request == nil {
		return nil, errors.New("request is nil")
	}

	claudeReq := anthropic.ConvertRequest(*request)
	req := Request{
		AnthropicVersion: anthropicVersion,
		// Model:            claudeReq.Model,
		Messages:    claudeReq.Messages,
		System:      claudeReq.System,
		MaxTokens:   claudeReq.MaxTokens,
		Temperature: claudeReq.Temperature,
		TopP:        claudeReq.TopP,
		TopK:        claudeReq.TopK,
		Stream:      claudeReq.Stream,
		Tools:       claudeReq.Tools,
	}

	c.Set(ctxkey.RequestModel, request.Model)
	c.Set(ctxkey.ConvertedRequest, req)
	return req, nil
}

func (a *Adaptor) DoResponse(c *gin.Context, resp *http.Response, meta *meta.Meta) (usage *model.Usage, err *model.ErrorWithStatusCode) {
	if meta.IsStream {
		err, usage = anthropic.StreamHandler(c, resp)
	} else {
		err, usage = anthropic.Handler(c, resp, meta.PromptTokens, meta.ActualModelName)
	}
	return
}

func (a *Adaptor) GetRatio(meta *meta.Meta) *ratio.Ratio {
	return adaptor.GetRatioHelper(meta, RatioMap)
}
