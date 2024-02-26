package gemini

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/songquanpeng/one-api/common/helper"
	channelhelper "github.com/songquanpeng/one-api/relay/channel"
	"github.com/songquanpeng/one-api/relay/channel/openai"
	relaymode "github.com/songquanpeng/one-api/relay/constant"
	"github.com/songquanpeng/one-api/relay/model"
	"github.com/songquanpeng/one-api/relay/util"
	"io"
	"net/http"
	"net/url"
)

type Adaptor struct {
}

func (a *Adaptor) Init(meta *util.RelayMeta) {}

func (a *Adaptor) GetRequestURL(meta *util.RelayMeta) (string, error) {
	version := helper.AssignOrDefault(meta.APIVersion, "v1")
	var action string

	switch meta.Mode {
	case relaymode.RelayModeEmbeddings:
		action = "batchEmbedContents"
	default:
		if meta.IsStream {
			action = "streamGenerateContent"
		} else {
			action = "generateContent"
		}
	}

	return fmt.Sprintf("%s/%s/models/%s:%s", meta.BaseURL, version, meta.ActualModelName, action), nil
}

func (a *Adaptor) SetupRequestHeader(c *gin.Context, req *http.Request, meta *util.RelayMeta) error {
	channelhelper.SetupCommonRequestHeader(c, req, meta)

	req.URL.RawQuery = url.Values{
		"key": {meta.APIKey},
	}.Encode()

	return nil
}

func (a *Adaptor) ConvertRequest(c *gin.Context, relayMode int, request *model.GeneralOpenAIRequest) (any, error) {
	if request == nil {
		return nil, errors.New("request is nil")
	}

	if relaymode.RelayModeEmbeddings == relayMode {
		return ConvertEmbeddingRequest(*request), nil
	} else {
		return ConvertRequest(*request), nil
	}
}

func (a *Adaptor) DoRequest(c *gin.Context, meta *util.RelayMeta, requestBody io.Reader) (*http.Response, error) {
	return channelhelper.DoRequestHelper(a, c, meta, requestBody)
}

func (a *Adaptor) DoResponse(c *gin.Context, resp *http.Response, meta *util.RelayMeta) (usage *model.Usage, err *model.ErrorWithStatusCode) {
	if relaymode.RelayModeEmbeddings == meta.Mode {
		err, usage = EmbeddingHandler(c, resp, meta.PromptTokens, meta.ActualModelName)
	} else if meta.IsStream {
		var responseText string
		err, responseText = StreamHandler(c, resp)
		usage = openai.ResponseText2Usage(responseText, meta.ActualModelName, meta.PromptTokens)
	} else {
		err, usage = Handler(c, resp, meta.PromptTokens, meta.ActualModelName)
	}
	return
}

func (a *Adaptor) GetModelList() []string {
	return ModelList
}

func (a *Adaptor) GetChannelName() string {
	return "google gemini"
}
