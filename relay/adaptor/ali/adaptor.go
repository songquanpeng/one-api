package ali

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/songquanpeng/one-api/relay/adaptor"
	"github.com/songquanpeng/one-api/relay/meta"
	"github.com/songquanpeng/one-api/relay/model"
	"github.com/songquanpeng/one-api/relay/relaymode"
	"io"
	"net/http"
)

// https://help.aliyun.com/zh/dashscope/developer-reference/api-details

type Adaptor struct {
	meta *meta.Meta
}

func (a *Adaptor) Init(meta *meta.Meta) {
	a.meta = meta
}

func (a *Adaptor) GetRequestURL(meta *meta.Meta) (string, error) {
	fullRequestURL := ""
	switch meta.Mode {
	case relaymode.Embeddings:
		fullRequestURL = fmt.Sprintf("%s/api/v1/services/embeddings/text-embedding/text-embedding", meta.BaseURL)
	case relaymode.ImagesGenerations:
		fullRequestURL = fmt.Sprintf("%s/api/v1/services/aigc/text2image/image-synthesis", meta.BaseURL)
	default:
		fullRequestURL = fmt.Sprintf("%s/api/v1/services/aigc/text-generation/generation", meta.BaseURL)
	}

	return fullRequestURL, nil
}

func (a *Adaptor) SetupRequestHeader(c *gin.Context, req *http.Request, meta *meta.Meta) error {
	adaptor.SetupCommonRequestHeader(c, req, meta)
	if meta.IsStream {
		req.Header.Set("Accept", "text/event-stream")
		req.Header.Set("X-DashScope-SSE", "enable")
	}
	req.Header.Set("Authorization", "Bearer "+meta.APIKey)

	if meta.Mode == relaymode.ImagesGenerations {
		req.Header.Set("X-DashScope-Async", "enable")
	}
	if a.meta.Config.Plugin != "" {
		req.Header.Set("X-DashScope-Plugin", a.meta.Config.Plugin)
	}
	return nil
}

func (a *Adaptor) ConvertRequest(c *gin.Context, relayMode int, request *model.GeneralOpenAIRequest) (any, error) {
	if request == nil {
		return nil, errors.New("request is nil")
	}
	switch relayMode {
	case relaymode.Embeddings:
		aliEmbeddingRequest := ConvertEmbeddingRequest(*request)
		return aliEmbeddingRequest, nil
	default:
		aliRequest := ConvertRequest(*request)
		return aliRequest, nil
	}
}

func (a *Adaptor) ConvertImageRequest(request *model.ImageRequest) (any, error) {
	if request == nil {
		return nil, errors.New("request is nil")
	}

	aliRequest := ConvertImageRequest(*request)
	return aliRequest, nil
}

func (a *Adaptor) DoRequest(c *gin.Context, meta *meta.Meta, requestBody io.Reader) (*http.Response, error) {
	return adaptor.DoRequestHelper(a, c, meta, requestBody)
}

func (a *Adaptor) DoResponse(c *gin.Context, resp *http.Response, meta *meta.Meta) (usage *model.Usage, err *model.ErrorWithStatusCode) {
	if meta.IsStream {
		err, usage = StreamHandler(c, resp)
	} else {
		switch meta.Mode {
		case relaymode.Embeddings:
			err, usage = EmbeddingHandler(c, resp)
		case relaymode.ImagesGenerations:
			err, usage = ImageHandler(c, resp)
		default:
			err, usage = Handler(c, resp)
		}
	}
	return
}

func (a *Adaptor) GetModelList() []string {
	return ModelList
}

func (a *Adaptor) GetChannelName() string {
	return "ali"
}
