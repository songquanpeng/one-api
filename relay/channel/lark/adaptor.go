package lark

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/songquanpeng/one-api/relay/channel"
	"github.com/songquanpeng/one-api/relay/constant"
	"github.com/songquanpeng/one-api/relay/model"
	"github.com/songquanpeng/one-api/relay/util"
	base "github.com/volcengine/volc-sdk-golang/base"
	"io"
	"io/ioutil"
	"net/http"
)

type Adaptor struct {
}

func (a *Adaptor) Init(meta *util.RelayMeta) {

}

func (a *Adaptor) GetRequestURL(meta *util.RelayMeta) (string, error) {
	fullRequestURL := fmt.Sprintf("%s/api/v1/chat", meta.BaseURL)
	if meta.Mode == constant.RelayModeEmbeddings {
		fullRequestURL = fmt.Sprintf("%s/api/v1/embeddings", meta.BaseURL)
	}
	return fullRequestURL, nil
}

type RequestBody struct {
	Model struct {
		Name string `json:"name"`
	} `json:"model"`
	Messages []struct {
		Role    string `json:"role"`
		Content string `json:"content"`
	} `json:"messages"`
	Parameters struct {
		Temperature  float64 `json:"temperature"`
		MaxNewTokens int     `json:"max_new_tokens"`
	} `json:"parameters"`
	Stream bool `json:"stream,omitempty"` // 使用omitempty，以便在stream不需要时不包含这个字段
}

func (a *Adaptor) SetupRequestHeader(c *gin.Context, req *http.Request, meta *util.RelayMeta) error {
	channel.SetupCommonRequestHeader(c, req, meta)
	credentials := base.Credentials{
		AccessKeyID:     meta.AK,
		SecretAccessKey: meta.SK,
		Service:         "ml_maas",
		Region:          "cn-beijing",
	}

	// 读取请求体
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		return err
	}
	// 确保在此函数结束时关闭原始的req.Body
	defer req.Body.Close()

	// 解析请求体
	var requestBody RequestBody
	err = json.Unmarshal(body, &requestBody)
	if err != nil {
		return err
	}

	// 如果是stream模式，添加stream字段
	if meta.IsStream {
		requestBody.Stream = true
	}

	// 重新编码为JSON
	newBody, err := json.Marshal(requestBody)
	if err != nil {
		return err
	}

	// 更新请求体
	req.Body = ioutil.NopCloser(bytes.NewBuffer(newBody))
	req.ContentLength = int64(len(newBody)) // 也许需要更新ContentLength
	req.Header.Set("Content-Type", "application/json")

	// 签名请求
	credentials.Sign(req)

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
	return "lark"
}
