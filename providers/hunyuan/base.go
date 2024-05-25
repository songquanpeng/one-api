package hunyuan

import (
	"encoding/json"
	"errors"
	"net/http"
	"one-api/common/requester"
	"one-api/model"
	"one-api/providers/base"
	"one-api/types"
	"strings"
)

type HunyuanProviderFactory struct{}

// 创建 HunyuanProvider
func (f HunyuanProviderFactory) Create(channel *model.Channel) base.ProviderInterface {
	return &HunyuanProvider{
		BaseProvider: base.BaseProvider{
			Config:    getConfig(),
			Channel:   channel,
			Requester: requester.NewHTTPRequester(*channel.Proxy, requestErrorHandle),
		},
	}
}

type HunyuanProvider struct {
	base.BaseProvider
}

func getConfig() base.ProviderConfig {
	return base.ProviderConfig{
		BaseURL:         "https://hunyuan.tencentcloudapi.com",
		ChatCompletions: "ChatCompletions",
	}
}

// 请求错误处理
func requestErrorHandle(resp *http.Response) *types.OpenAIError {
	tunyuanError := &HunyuanResponseError{}
	err := json.NewDecoder(resp.Body).Decode(tunyuanError)
	if err != nil {
		return nil
	}

	return errorHandle(tunyuanError)
}

// 错误处理
func errorHandle(tunyuanError *HunyuanResponseError) *types.OpenAIError {
	if tunyuanError.Error == nil {
		return nil
	}
	return &types.OpenAIError{
		Message: tunyuanError.Error.Message,
		Type:    "tunyuan_error",
		Code:    tunyuanError.Error.Code,
	}
}

// 获取请求头
func (p *HunyuanProvider) GetRequestHeaders() (headers map[string]string) {
	headers = make(map[string]string)
	p.CommonRequestHeaders(headers)

	return headers
}

func (p *HunyuanProvider) parseHunyuanConfig(config string) (secretId string, secretKey string, err error) {
	parts := strings.Split(config, "|")
	if len(parts) != 2 {
		err = errors.New("invalid tunyuan config")
		return
	}

	secretId = parts[0]
	secretKey = parts[1]
	return
}
