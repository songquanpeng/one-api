package anthropic

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

const (
	// NativeAnthropicEndpoint is the endpoint for native Anthropic API
	NativeAnthropicEndpoint = "/v1/messages"
	// ThirdPartyAnthropicEndpoint is the endpoint for third-party providers supporting Anthropic protocol
	ThirdPartyAnthropicEndpoint = "/anthropic/v1/messages"
)

type Adaptor struct {
}

func (a *Adaptor) Init(meta *meta.Meta) {

}

func (a *Adaptor) GetRequestURL(meta *meta.Meta) (string, error) {
	// For native Anthropic API
	if strings.Contains(meta.BaseURL, "api.anthropic.com") {
		return fmt.Sprintf("%s%s", meta.BaseURL, NativeAnthropicEndpoint), nil
	}

	// For third-party providers supporting Anthropic protocol
	// Common scenario: BaseURL ends with /v1 (e.g., https://api.deepseek.com/v1)
	// ThirdPartyAnthropicEndpoint is /anthropic/v1/messages
	// We need to avoid: /v1/anthropic/v1/messages (double /v1)
	baseURL := strings.TrimSuffix(meta.BaseURL, "/")

	// Smart handling: if ThirdPartyAnthropicEndpoint already contains /v1/,
	// and baseURL ends with /v1, remove it from baseURL to prevent duplication
	if strings.HasPrefix(ThirdPartyAnthropicEndpoint, "/") &&
		strings.Contains(ThirdPartyAnthropicEndpoint, "/v1/") &&
		strings.HasSuffix(baseURL, "/v1") {
		baseURL = strings.TrimSuffix(baseURL, "/v1")
	}

	return fmt.Sprintf("%s%s", baseURL, ThirdPartyAnthropicEndpoint), nil
}

func (a *Adaptor) SetupRequestHeader(c *gin.Context, req *http.Request, meta *meta.Meta) error {
	adaptor.SetupCommonRequestHeader(c, req, meta)
	req.Header.Set("x-api-key", meta.APIKey)
	anthropicVersion := c.Request.Header.Get("anthropic-version")
	if anthropicVersion == "" {
		anthropicVersion = "2023-06-01"
	}
	req.Header.Set("anthropic-version", anthropicVersion)
	req.Header.Set("anthropic-beta", "messages-2023-12-15")

	// https://x.com/alexalbert__/status/1812921642143900036
	// claude-3-5-sonnet can support 8k context
	if strings.HasPrefix(meta.ActualModelName, "claude-3-5-sonnet") {
		req.Header.Set("anthropic-beta", "max-tokens-3-5-sonnet-2024-07-15")
	}

	return nil
}

func (a *Adaptor) ConvertRequest(c *gin.Context, relayMode int, request *model.GeneralOpenAIRequest) (any, error) {
	if request == nil {
		return nil, errors.New("request is nil")
	}

	// For native Anthropic protocol requests, return the request as-is (no conversion needed)
	if relayMode == relaymode.AnthropicMessages {
		// The request should already be in Anthropic format, so we pass it through
		// This will be handled by the caller which already has the anthropic request
		return request, nil
	}

	// For OpenAI to Anthropic conversion (existing functionality)
	return ConvertRequest(*request), nil
}

func (a *Adaptor) ConvertImageRequest(request *model.ImageRequest) (any, error) {
	if request == nil {
		return nil, errors.New("request is nil")
	}
	return request, nil
}

func (a *Adaptor) DoRequest(c *gin.Context, meta *meta.Meta, requestBody io.Reader) (*http.Response, error) {
	return adaptor.DoRequestHelper(a, c, meta, requestBody)
}

func (a *Adaptor) DoResponse(c *gin.Context, resp *http.Response, meta *meta.Meta) (usage *model.Usage, err *model.ErrorWithStatusCode) {
	// For native Anthropic protocol requests, handle response directly without conversion
	if meta.Mode == relaymode.AnthropicMessages {
		if meta.IsStream {
			err, usage = DirectStreamHandler(c, resp)
		} else {
			err, usage = DirectHandler(c, resp, meta.PromptTokens, meta.ActualModelName)
		}
		return
	}

	// For OpenAI to Anthropic conversion (existing functionality)
	if meta.IsStream {
		err, usage = StreamHandler(c, resp)
	} else {
		err, usage = Handler(c, resp, meta.PromptTokens, meta.ActualModelName)
	}
	return
}

func (a *Adaptor) GetModelList() []string {
	return ModelList
}

func (a *Adaptor) GetChannelName() string {
	return "anthropic"
}
