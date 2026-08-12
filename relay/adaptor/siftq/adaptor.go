package siftq

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/songquanpeng/one-api/common/client"
	"github.com/songquanpeng/one-api/relay/adaptor/openai"
	"github.com/songquanpeng/one-api/relay/meta"
	relaymodel "github.com/songquanpeng/one-api/relay/model"
)

type Adaptor struct{}

func (*Adaptor) Init(_ *meta.Meta) {}

func (*Adaptor) GetRequestURL(providerMeta *meta.Meta) (string, error) {
	if providerMeta == nil {
		return "", errors.New("relay metadata is required")
	}
	return BuildURL(providerMeta.BaseURL, QueryVideoPath)
}

func (*Adaptor) SetupRequestHeader(_ *gin.Context, request *http.Request, providerMeta *meta.Meta) error {
	if providerMeta == nil || strings.TrimSpace(providerMeta.APIKey) == "" {
		return errors.New("SiftQ API key is required")
	}
	request.Header.Set("Authorization", "Bearer "+providerMeta.APIKey)
	if request.Method == http.MethodPost {
		request.Header.Set("Content-Type", "application/json")
	}
	request.Header.Set("Accept", "application/json")
	return nil
}

func (*Adaptor) ConvertRequest(_ *gin.Context, _ int, _ *relaymodel.GeneralOpenAIRequest) (any, error) {
	return nil, errors.New("SiftQ video requests use the provider-neutral /v1/videos routes")
}

func (*Adaptor) ConvertImageRequest(_ *relaymodel.ImageRequest) (any, error) {
	return nil, errors.New("SiftQ does not implement OpenAI image generation")
}

func (a *Adaptor) DoRequest(c *gin.Context, providerMeta *meta.Meta, body io.Reader) (*http.Response, error) {
	fullURL, err := a.GetRequestURL(providerMeta)
	if err != nil {
		return nil, err
	}
	request, err := http.NewRequestWithContext(c.Request.Context(), c.Request.Method, fullURL, body)
	if err != nil {
		return nil, fmt.Errorf("create SiftQ request: %w", err)
	}
	if err := a.SetupRequestHeader(c, request, providerMeta); err != nil {
		return nil, err
	}
	return client.HTTPClient.Do(request)
}

func (*Adaptor) DoResponse(c *gin.Context, response *http.Response, _ *meta.Meta) (*relaymodel.Usage, *relaymodel.ErrorWithStatusCode) {
	if response == nil {
		return nil, openai.ErrorWrapper(errors.New("SiftQ response is nil"), "bad_response", http.StatusBadGateway)
	}
	defer response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return nil, openai.ErrorWrapper(fmt.Errorf("SiftQ returned HTTP %d", response.StatusCode), "siftq_api_error", response.StatusCode)
	}
	c.Header("Content-Type", "application/json")
	c.Status(response.StatusCode)
	if _, err := io.Copy(c.Writer, io.LimitReader(response.Body, MaxRequestBytes+1)); err != nil {
		return nil, openai.ErrorWrapper(err, "copy_response_failed", http.StatusBadGateway)
	}
	return &relaymodel.Usage{}, nil
}

func (*Adaptor) GetModelList() []string {
	return append([]string(nil), ModelList...)
}

func (*Adaptor) GetChannelName() string {
	return ChannelName
}

func BuildURL(baseURL, endpoint string) (string, error) {
	if strings.TrimSpace(baseURL) == "" {
		baseURL = DefaultBaseURL
	}
	parsed, err := url.Parse(baseURL)
	if err != nil || parsed.Host == "" || (parsed.Scheme != "http" && parsed.Scheme != "https") {
		return "", errors.New("SiftQ base URL must be an absolute HTTP or HTTPS URL")
	}
	if parsed.User != nil || parsed.RawQuery != "" || parsed.Fragment != "" {
		return "", errors.New("SiftQ base URL must not contain credentials, query parameters, or a fragment")
	}
	return strings.TrimRight(baseURL, "/") + "/" + strings.TrimLeft(endpoint, "/"), nil
}

type ListFilters struct {
	PageNum  int
	PageSize int
	Status   string
	TaskIDs  []string
	Model    string
	TaskType string
}

func BuildListURL(baseURL string, filters ListFilters) (string, error) {
	target, err := BuildURL(baseURL, QueryVideoPath)
	if err != nil {
		return "", err
	}
	parsed, err := url.Parse(target)
	if err != nil {
		return "", err
	}
	query := parsed.Query()
	if filters.PageNum > 0 {
		query.Set("page_num", fmt.Sprintf("%d", filters.PageNum))
	}
	if filters.PageSize > 0 {
		query.Set("page_size", fmt.Sprintf("%d", filters.PageSize))
	}
	if filters.Status != "" {
		query.Set("filter.status", filters.Status)
	}
	for _, taskID := range filters.TaskIDs {
		query.Add("filter.task_ids", taskID)
	}
	if filters.Model != "" {
		query.Set("filter.model", filters.Model)
	}
	if filters.TaskType != "" {
		query.Set("filter.task_type", filters.TaskType)
	}
	parsed.RawQuery = query.Encode()
	return parsed.String(), nil
}

func NewUpstreamRequest(c *gin.Context, channelMeta *meta.Meta, method, target string, body io.Reader) (*http.Request, error) {
	if channelMeta == nil || strings.TrimSpace(channelMeta.APIKey) == "" {
		return nil, errors.New("SiftQ API key is required")
	}
	request, err := http.NewRequestWithContext(c.Request.Context(), method, target, body)
	if err != nil {
		return nil, err
	}
	request.Header.Set("Authorization", "Bearer "+channelMeta.APIKey)
	request.Header.Set("Accept", "application/json")
	if method == http.MethodPost {
		request.Header.Set("Content-Type", "application/json")
	}
	return request, nil
}
