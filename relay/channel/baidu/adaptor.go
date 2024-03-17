package baidu

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/songquanpeng/one-api/relay/channel"
	"github.com/songquanpeng/one-api/relay/constant"
	"github.com/songquanpeng/one-api/relay/model"
	"github.com/songquanpeng/one-api/relay/util"
)

type Adaptor struct {
}

func (a *Adaptor) Init(meta *util.RelayMeta) {

}

func (a *Adaptor) GetRequestURL(meta *util.RelayMeta) (string, error) {
	// https://cloud.baidu.com/doc/WENXINWORKSHOP/s/clntwmv7t
	suffix := "chat/"
	if strings.HasPrefix(meta.ActualModelName, "Embedding") {
		suffix = "embeddings/"
	}
	if strings.HasPrefix(meta.ActualModelName, "bge-large") {
		suffix = "embeddings/"
	}
	if strings.HasPrefix(meta.ActualModelName, "tao-8k") {
		suffix = "embeddings/"
	}
	switch meta.ActualModelName {
	case "ERNIE-4.0":
		suffix += "completions_pro"
	case "ERNIE-Bot-4":
		suffix += "completions_pro"
	case "ERNIE-3.5-8K":
		suffix += "completions"
	case "ERNIE-Bot-8K":
		suffix += "ernie_bot_8k"
	case "ERNIE-Bot":
		suffix += "completions"
	case "ERNIE-Speed":
		suffix += "ernie_speed"
	case "ERNIE-Bot-turbo":
		suffix += "eb-instant"
	case "BLOOMZ-7B":
		suffix += "bloomz_7b1"
	case "Embedding-V1":
		suffix += "embedding-v1"
	case "bge-large-zh":
		suffix += "bge_large_zh"
	case "bge-large-en":
		suffix += "bge_large_en"
	case "tao-8k":
		suffix += "tao_8k"
	default:
		suffix += meta.ActualModelName
	}
	fullRequestURL := fmt.Sprintf("%s/rpc/2.0/ai_custom/v1/wenxinworkshop/%s", meta.BaseURL, suffix)
	var accessToken string
	var err error
	if accessToken, err = GetAccessToken(meta.APIKey); err != nil {
		return "", err
	}
	fullRequestURL += "?access_token=" + accessToken
	return fullRequestURL, nil
}

func (a *Adaptor) SetupRequestHeader(c *gin.Context, req *http.Request, meta *util.RelayMeta) error {
	channel.SetupCommonRequestHeader(c, req, meta)
	req.Header.Set("Authorization", "Bearer "+meta.APIKey)
	return nil
}

func (a *Adaptor) ConvertRequest(c *gin.Context, relayMode int, request *model.GeneralOpenAIRequest) (any, error) {
	if request == nil {
		return nil, errors.New("request is nil")
	}
	switch relayMode {
	case constant.RelayModeEmbeddings:
		baiduEmbeddingRequest := ConvertEmbeddingRequest(*request)
		return baiduEmbeddingRequest, nil
	default:
		baiduRequest := ConvertRequest(*request)
		return baiduRequest, nil
	}
}

func (a *Adaptor) DoRequest(c *gin.Context, meta *util.RelayMeta, requestBody io.Reader) (*http.Response, error) {
	return channel.DoRequestHelper(a, c, meta, requestBody)
}

func (a *Adaptor) DoResponse(c *gin.Context, resp *http.Response, meta *util.RelayMeta) (usage *model.Usage, err *model.ErrorWithStatusCode) {
	if meta.IsStream {
		err, usage = StreamHandler(c, resp)
	} else {
		switch meta.Mode {
		case constant.RelayModeEmbeddings:
			err, usage = EmbeddingHandler(c, resp)
		default:
			err, usage = Handler(c, resp)
		}
	}
	return
}

func (a *Adaptor) GetModelList() []string {
	return ModelList
}

func (a *Adaptor) GetChannelName() string {
	return "baidu"
}
