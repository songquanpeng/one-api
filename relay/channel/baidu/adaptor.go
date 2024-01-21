package baidu

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"one-api/relay/channel/openai"
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
