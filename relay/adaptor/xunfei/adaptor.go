package xunfei

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/songquanpeng/one-api/relay/adaptor"
	"github.com/songquanpeng/one-api/relay/adaptor/openai"
	"github.com/songquanpeng/one-api/relay/meta"
	"github.com/songquanpeng/one-api/relay/model"
	"io"
	"net/http"
	"strings"
)

type Adaptor struct {
	request *model.GeneralOpenAIRequest
	meta    *meta.Meta
}

func (a *Adaptor) Init(meta *meta.Meta) {
	a.meta = meta
}

func (a *Adaptor) GetRequestURL(meta *meta.Meta) (string, error) {
	return "", nil
}

func (a *Adaptor) SetupRequestHeader(c *gin.Context, req *http.Request, meta *meta.Meta) error {
	adaptor.SetupCommonRequestHeader(c, req, meta)
	// check DoResponse for auth part
	return nil
}

func (a *Adaptor) ConvertRequest(c *gin.Context, relayMode int, request *model.GeneralOpenAIRequest) (any, error) {
	if request == nil {
		return nil, errors.New("request is nil")
	}
	a.request = request
	return nil, nil
}

func (a *Adaptor) ConvertImageRequest(request *model.ImageRequest) (any, error) {
	if request == nil {
		return nil, errors.New("request is nil")
	}
	return request, nil
}

func (a *Adaptor) DoRequest(c *gin.Context, meta *meta.Meta, requestBody io.Reader) (*http.Response, error) {
	// xunfei's request is not http request, so we don't need to do anything here
	dummyResp := &http.Response{}
	dummyResp.StatusCode = http.StatusOK
	return dummyResp, nil
}

func (a *Adaptor) DoResponse(c *gin.Context, resp *http.Response, meta *meta.Meta) (usage *model.Usage, err *model.ErrorWithStatusCode) {
	splits := strings.Split(meta.APIKey, "|")
	if len(splits) != 3 {
		return nil, openai.ErrorWrapper(errors.New("invalid auth"), "invalid_auth", http.StatusBadRequest)
	}
	if a.request == nil {
		return nil, openai.ErrorWrapper(errors.New("request is nil"), "request_is_nil", http.StatusBadRequest)
	}
	version := parseAPIVersionByModelName(meta.ActualModelName)
	if version == "" {
		version = a.meta.Config.APIVersion
	}
	if version == "" {
		version = "v1.1"
	}
	a.meta.Config.APIVersion = version
	if meta.IsStream {
		err, usage = StreamHandler(c, meta, *a.request, splits[0], splits[1], splits[2])
	} else {
		err, usage = Handler(c, meta, *a.request, splits[0], splits[1], splits[2])
	}
	return
}

func (a *Adaptor) GetModelList() []string {
	return ModelList
}

func (a *Adaptor) GetChannelName() string {
	return "xunfei"
}
