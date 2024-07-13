package cloudflare

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/songquanpeng/one-api/relay/adaptor"
	"github.com/songquanpeng/one-api/relay/meta"
	"github.com/songquanpeng/one-api/relay/model"
	"github.com/songquanpeng/one-api/relay/relaymode"
)

type Adaptor struct {
	meta *meta.Meta
}

// ConvertImageRequest implements adaptor.Adaptor.
func (*Adaptor) ConvertImageRequest(request *model.ImageRequest) (any, error) {
	return nil, errors.New("not implemented")
}

// ConvertImageRequest implements adaptor.Adaptor.

func (a *Adaptor) Init(meta *meta.Meta) {
	a.meta = meta
}

// WorkerAI cannot be used across accounts with AIGateWay
// https://developers.cloudflare.com/ai-gateway/providers/workersai/#openai-compatible-endpoints
// https://gateway.ai.cloudflare.com/v1/{account_id}/{gateway_id}/workers-ai
func (a *Adaptor) isAIGateWay(baseURL string) bool {
	return strings.HasPrefix(baseURL, "https://gateway.ai.cloudflare.com") && strings.HasSuffix(baseURL, "/workers-ai")
}

func (a *Adaptor) GetRequestURL(meta *meta.Meta) (string, error) {
	isAIGateWay := a.isAIGateWay(meta.BaseURL)
	var urlPrefix string
	if isAIGateWay {
		urlPrefix = meta.BaseURL
	} else {
		urlPrefix = fmt.Sprintf("%s/client/v4/accounts/%s/ai", meta.BaseURL, meta.Config.UserID)
	}

	switch meta.Mode {
	case relaymode.ChatCompletions:
		return fmt.Sprintf("%s/v1/chat/completions", urlPrefix), nil
	case relaymode.Embeddings:
		return fmt.Sprintf("%s/v1/embeddings", urlPrefix), nil
	default:
		if isAIGateWay {
			return fmt.Sprintf("%s/%s", urlPrefix, meta.ActualModelName), nil
		}
		return fmt.Sprintf("%s/run/%s", urlPrefix, meta.ActualModelName), nil
	}
}

func (a *Adaptor) SetupRequestHeader(c *gin.Context, req *http.Request, meta *meta.Meta) error {
	adaptor.SetupCommonRequestHeader(c, req, meta)
	req.Header.Set("Authorization", "Bearer "+meta.APIKey)
	return nil
}

func (a *Adaptor) ConvertRequest(c *gin.Context, relayMode int, request *model.GeneralOpenAIRequest) (any, error) {
	if request == nil {
		return nil, errors.New("request is nil")
	}
	switch relayMode {
	case relaymode.Completions:
		return ConvertCompletionsRequest(*request), nil
	case relaymode.ChatCompletions, relaymode.Embeddings:
		return request, nil
	default:
		return nil, errors.New("not implemented")
	}
}

func (a *Adaptor) DoRequest(c *gin.Context, meta *meta.Meta, requestBody io.Reader) (*http.Response, error) {
	return adaptor.DoRequestHelper(a, c, meta, requestBody)
}

func (a *Adaptor) DoResponse(c *gin.Context, resp *http.Response, meta *meta.Meta) (usage *model.Usage, err *model.ErrorWithStatusCode) {
	if meta.IsStream {
		err, usage = StreamHandler(c, resp, meta.PromptTokens, meta.ActualModelName)
	} else {
		err, usage = Handler(c, resp, meta.PromptTokens, meta.ActualModelName)
	}
	return
}

func (a *Adaptor) GetModelList() []string {
	return ModelList
}

func (a *Adaptor) GetChannelName() string {
	return "cloudflare"
}
