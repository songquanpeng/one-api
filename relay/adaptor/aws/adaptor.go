package aws

import (
	"errors"
	"io"
	"net/http"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
	"github.com/gin-gonic/gin"
	"github.com/songquanpeng/one-api/relay/adaptor"
	"github.com/songquanpeng/one-api/relay/adaptor/aws/utils"
	"github.com/songquanpeng/one-api/relay/meta"
	"github.com/songquanpeng/one-api/relay/model"
)

var _ adaptor.Adaptor = new(Adaptor)

type Adaptor struct {
	awsAdapter utils.AwsAdapter

	Meta      *meta.Meta
	AwsClient *bedrockruntime.Client
}

func (a *Adaptor) Init(meta *meta.Meta) {
	a.Meta = meta
	a.AwsClient = bedrockruntime.New(bedrockruntime.Options{
		Region:      meta.Config.Region,
		Credentials: aws.NewCredentialsCache(credentials.NewStaticCredentialsProvider(meta.Config.AK, meta.Config.SK, "")),
	})
}

func (a *Adaptor) ConvertRequest(c *gin.Context, relayMode int, request *model.GeneralOpenAIRequest) (any, error) {
	if request == nil {
		return nil, errors.New("request is nil")
	}

	adaptor := GetAdaptor(request.Model)
	if adaptor == nil {
		return nil, errors.New("adaptor not found")
	}

	a.awsAdapter = adaptor
	return adaptor.ConvertRequest(c, relayMode, request)
}

func (a *Adaptor) DoResponse(c *gin.Context, resp *http.Response, meta *meta.Meta) (usage *model.Usage, err *model.ErrorWithStatusCode) {
	if a.awsAdapter == nil {
		return nil, utils.WrapErr(errors.New("awsAdapter is nil"))
	}
	return a.awsAdapter.DoResponse(c, a.AwsClient, meta)
}

func (a *Adaptor) GetModelList() (models []string) {
	for model := range adaptors {
		models = append(models, model)
	}
	return
}

func (a *Adaptor) GetChannelName() string {
	return "aws"
}

func (a *Adaptor) GetRequestURL(meta *meta.Meta) (string, error) {
	return "", nil
}

func (a *Adaptor) SetupRequestHeader(c *gin.Context, req *http.Request, meta *meta.Meta) error {
	return nil
}

func (a *Adaptor) ConvertImageRequest(request *model.ImageRequest) (any, error) {
	if request == nil {
		return nil, errors.New("request is nil")
	}
	return request, nil
}

func (a *Adaptor) DoRequest(c *gin.Context, meta *meta.Meta, requestBody io.Reader) (*http.Response, error) {
	return nil, nil
}
