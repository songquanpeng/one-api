package aws

import (
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
	"github.com/songquanpeng/one-api/common/ctxkey"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"github.com/songquanpeng/one-api/relay/adaptor"
	"github.com/songquanpeng/one-api/relay/adaptor/aws/utils"
	"github.com/songquanpeng/one-api/relay/billing/ratio"
	"github.com/songquanpeng/one-api/relay/meta"
	"github.com/songquanpeng/one-api/relay/model"
)

var _ utils.AwsAdapter = new(Adaptor)

type Adaptor struct {
}

func (a *Adaptor) ConvertRequest(c *gin.Context, relayMode int, request *model.GeneralOpenAIRequest) (any, error) {
	if request == nil {
		return nil, errors.New("request is nil")
	}

	llamaReq := ConvertRequest(*request)
	c.Set(ctxkey.RequestModel, request.Model)
	c.Set(ctxkey.ConvertedRequest, llamaReq)
	return llamaReq, nil
}

func (a *Adaptor) DoResponse(c *gin.Context, awsCli *bedrockruntime.Client, meta *meta.Meta) (usage *model.Usage, err *model.ErrorWithStatusCode) {
	if meta.IsStream {
		err, usage = StreamHandler(c, awsCli)
	} else {
		err, usage = Handler(c, awsCli, meta.ActualModelName)
	}
	return
}

func (a *Adaptor) GetRatio(meta *meta.Meta) *ratio.Ratio {
	return adaptor.GetRatioHelper(meta, RatioMap)
}
