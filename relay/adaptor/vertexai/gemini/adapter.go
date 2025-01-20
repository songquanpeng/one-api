package vertexai

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"github.com/songquanpeng/one-api/common/ctxkey"
	"github.com/songquanpeng/one-api/relay/adaptor"
	"github.com/songquanpeng/one-api/relay/adaptor/gemini"
	"github.com/songquanpeng/one-api/relay/adaptor/openai"
	"github.com/songquanpeng/one-api/relay/billing/ratio"
	"github.com/songquanpeng/one-api/relay/meta"
	"github.com/songquanpeng/one-api/relay/model"
	"github.com/songquanpeng/one-api/relay/relaymode"
)

var RatioMap = gemini.RatioMap

type Adaptor struct {
}

func (a *Adaptor) ConvertRequest(c *gin.Context, relayMode int, request *model.GeneralOpenAIRequest) (any, error) {
	if request == nil {
		return nil, errors.New("request is nil")
	}

	geminiRequest := gemini.ConvertRequest(*request)
	c.Set(ctxkey.RequestModel, request.Model)
	c.Set(ctxkey.ConvertedRequest, geminiRequest)
	return geminiRequest, nil
}

func (a *Adaptor) DoResponse(c *gin.Context, resp *http.Response, meta *meta.Meta) (usage *model.Usage, err *model.ErrorWithStatusCode) {
	if meta.IsStream {
		var responseText string
		err, responseText = gemini.StreamHandler(c, resp)
		usage = openai.ResponseText2Usage(responseText, meta.ActualModelName, meta.PromptTokens)
	} else {
		switch meta.Mode {
		case relaymode.Embeddings:
			err, usage = gemini.EmbeddingHandler(c, resp)
		default:
			err, usage = gemini.Handler(c, resp, meta.PromptTokens, meta.ActualModelName)
		}
	}
	return
}

func (a *Adaptor) GetRatio(meta *meta.Meta) *ratio.Ratio {
	return adaptor.GetRatioHelper(meta, RatioMap)
}
