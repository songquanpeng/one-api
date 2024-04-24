package vertex

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	channelhelper "github.com/songquanpeng/one-api/relay/adaptor"
	"github.com/songquanpeng/one-api/relay/meta"
	"github.com/songquanpeng/one-api/relay/model"
	"io"
	"net/http"
)

type Adaptor struct {
}

func (a *Adaptor) Init(meta *meta.Meta) {

}

// https://$LOCATION-aiplatform.googleapis.com/v1/projects/$PROJECT_ID/locations/$LOCATION/publishers/anthropic/models/$MODEL:streamRawPredict
func (a *Adaptor) GetRequestURL(meta *meta.Meta) (string, error) {
	_ = meta
	location := "us-east5"
	projectId := "monica-dev-392609"
	models := "claude-3-opus@20240229"

	return fmt.Sprintf("https://%s-aiplatform.googleapis.com/v1/projects/%s/locations/%s/publishers/anthropic/models/%s:streamRawPredict",
		location, projectId, location, models), nil
}

func (a *Adaptor) SetupRequestHeader(c *gin.Context, req *http.Request, meta *meta.Meta) error {
	channelhelper.SetupCommonRequestHeader(c, req, meta)
	token, err := getToken(c, meta)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	return nil
}

func (a *Adaptor) ConvertRequest(c *gin.Context, relayMode int, request *model.GeneralOpenAIRequest) (any, error) {
	_, _ = c, relayMode
	if request == nil {
		return nil, errors.New("request is nil")
	}
	return ConvertRequest(*request), nil
}

func (a *Adaptor) ConvertImageRequest(request *model.ImageRequest) (any, error) {
	if request == nil {
		return nil, errors.New("request is nil")
	}
	return request, nil
}

func (a *Adaptor) DoRequest(c *gin.Context, meta *meta.Meta, requestBody io.Reader) (*http.Response, error) {
	return channelhelper.DoRequestHelper(a, c, meta, requestBody)
}

func (a *Adaptor) DoResponse(c *gin.Context, resp *http.Response, meta *meta.Meta) (usage *model.Usage, err *model.ErrorWithStatusCode) {
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
	return "vertex"
}
