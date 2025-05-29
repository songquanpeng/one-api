package vertexai

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/songquanpeng/one-api/relay/adaptor"
	channelhelper "github.com/songquanpeng/one-api/relay/adaptor"
	"github.com/songquanpeng/one-api/relay/meta"
	"github.com/songquanpeng/one-api/relay/model"
	relaymodel "github.com/songquanpeng/one-api/relay/model"
)

var _ adaptor.Adaptor = new(Adaptor)

const channelName = "vertexai"

type Adaptor struct{}

func (a *Adaptor) Init(meta *meta.Meta) {
}

func (a *Adaptor) ConvertRequest(c *gin.Context, relayMode int, request *model.GeneralOpenAIRequest) (any, error) {
	if request == nil {
		return nil, errors.New("request is nil")
	}

	adaptor := GetAdaptor(request.Model)
	if adaptor == nil {
		return nil, errors.New("adaptor not found")
	}

	return adaptor.ConvertRequest(c, relayMode, request)
}

func (a *Adaptor) DoResponse(c *gin.Context, resp *http.Response, meta *meta.Meta) (usage *model.Usage, err *model.ErrorWithStatusCode) {
	adaptor := GetAdaptor(meta.ActualModelName)
	if adaptor == nil {
		return nil, &relaymodel.ErrorWithStatusCode{
			StatusCode: http.StatusInternalServerError,
			Error: relaymodel.Error{
				Message: "adaptor not found",
			},
		}
	}
	return adaptor.DoResponse(c, resp, meta)
}

func (a *Adaptor) GetModelList() (models []string) {
	models = modelList
	return
}

func (a *Adaptor) GetChannelName() string {
	return channelName
}

func (a *Adaptor) GetRequestURL(meta *meta.Meta) (string, error) {
	suffix := ""
	modelType := PredictModelType(meta.ActualModelName)
	if modelType == VertexAIGemini {
		if meta.IsStream {
			suffix = "streamGenerateContent?alt=sse"
		} else {
			suffix = "generateContent"
		}
	} else if modelType == VertexAIEmbedding {
		suffix = "predict"
	} else {
		if meta.IsStream {
			suffix = "streamRawPredict?alt=sse"
		} else {
			suffix = "rawPredict"
		}
	}

	model := meta.ActualModelName
	if strings.Contains(model, "?") {
		// TODO: Maybe fix meta.ActualModelName?
		model = strings.Split(model, "?")[0]
	}

	if meta.BaseURL != "" {
		return fmt.Sprintf(
			"%s/v1/projects/%s/locations/%s/publishers/google/models/%s:%s",
			meta.BaseURL,
			meta.Config.VertexAIProjectID,
			meta.Config.Region,
			model,
			suffix,
		), nil
	}
	return fmt.Sprintf(
		"https://%s-aiplatform.googleapis.com/v1/projects/%s/locations/%s/publishers/google/models/%s:%s",
		meta.Config.Region,
		meta.Config.VertexAIProjectID,
		meta.Config.Region,
		model,
		suffix,
	), nil
}

func (a *Adaptor) SetupRequestHeader(c *gin.Context, req *http.Request, meta *meta.Meta) error {
	adaptor.SetupCommonRequestHeader(c, req, meta)
	token, err := getToken(c, meta.ChannelId, meta.Config.VertexAIADC)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	return nil
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

func PredictModelType(model string) VertexAIModelType {
	if strings.HasPrefix(model, "gemini-") {
		return VertexAIGemini
	}
	if strings.HasPrefix(model, "text-embedding") || strings.HasPrefix(model, "text-multilingual-embedding") {
		return VertexAIEmbedding
	}
	return VertexAIClaude
}
