package adaptor

import (
	"github.com/gin-gonic/gin"
	"github.com/songquanpeng/one-api/relay/meta"
	"github.com/songquanpeng/one-api/relay/model"
	"io"
	"net/http"
)

type Adaptor interface {
	Init(meta *meta.Meta)
	GetRequestURL(meta *meta.Meta) (string, error)
	SetupRequestHeader(c *gin.Context, req *http.Request, meta *meta.Meta) error
	ConvertRequest(c *gin.Context, relayMode int, request *model.GeneralOpenAIRequest) (any, error)
	ConvertImageRequest(request *model.ImageRequest) (any, error)
	DoRequest(c *gin.Context, meta *meta.Meta, requestBody io.Reader) (*http.Response, error)
	DoResponse(c *gin.Context, resp *http.Response, meta *meta.Meta) (usage *model.Usage, err *model.ErrorWithStatusCode)
	GetModelList() []string
	GetChannelName() string
}
