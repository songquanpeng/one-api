package channel

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"one-api/relay/channel/openai"
)

type Adaptor interface {
	GetRequestURL() string
	Auth(c *gin.Context) error
	ConvertRequest(request *openai.GeneralOpenAIRequest) (any, error)
	DoRequest(request *openai.GeneralOpenAIRequest) error
	DoResponse(c *gin.Context, resp *http.Response) (*openai.ErrorWithStatusCode, *openai.Usage, error)
}
