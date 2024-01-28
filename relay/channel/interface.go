package channel

import (
	"github.com/gin-gonic/gin"
	"github.com/songquanpeng/one-api/relay/channel/openai"
	"net/http"
)

type Adaptor interface {
	GetRequestURL() string
	Auth(c *gin.Context) error
	ConvertRequest(request *openai.GeneralOpenAIRequest) (any, error)
	DoRequest(request *openai.GeneralOpenAIRequest) error
	DoResponse(c *gin.Context, resp *http.Response) (*openai.ErrorWithStatusCode, *openai.Usage, error)
}
