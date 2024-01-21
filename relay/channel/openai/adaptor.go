package openai

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

type Adaptor struct {
}

func (a *Adaptor) Auth(c *gin.Context) error {
	return nil
}

func (a *Adaptor) ConvertRequest(request *GeneralOpenAIRequest) (any, error) {
	return nil, nil
}

func (a *Adaptor) DoResponse(c *gin.Context, resp *http.Response) (*ErrorWithStatusCode, *Usage, error) {
	return nil, nil, nil
}
