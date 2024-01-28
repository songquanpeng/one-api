package aiproxy

import (
	"github.com/gin-gonic/gin"
	"github.com/songquanpeng/one-api/relay/channel/openai"
	"net/http"
)

type Adaptor struct {
}

func (a *Adaptor) Auth(c *gin.Context) error {
	return nil
}

func (a *Adaptor) ConvertRequest(request *openai.GeneralOpenAIRequest) (any, error) {
	return nil, nil
}

func (a *Adaptor) DoResponse(c *gin.Context, resp *http.Response) (*openai.ErrorWithStatusCode, *openai.Usage, error) {
	return nil, nil, nil
}
