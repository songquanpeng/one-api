package tencent

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"one-api/common/requester"
	"one-api/model"
	"one-api/providers/base"
	"one-api/types"
	"sort"
	"strconv"
	"strings"
)

type TencentProviderFactory struct{}

// 创建 TencentProvider
func (f TencentProviderFactory) Create(channel *model.Channel) base.ProviderInterface {
	return &TencentProvider{
		BaseProvider: base.BaseProvider{
			Config:    getConfig(),
			Channel:   channel,
			Requester: requester.NewHTTPRequester(channel.Proxy, requestErrorHandle),
		},
	}
}

type TencentProvider struct {
	base.BaseProvider
}

func getConfig() base.ProviderConfig {
	return base.ProviderConfig{
		BaseURL:         "https://hunyuan.cloud.tencent.com",
		ChatCompletions: "/hyllm/v1/chat/completions",
	}
}

// 请求错误处理
func requestErrorHandle(resp *http.Response) *types.OpenAIError {
	var tencentError *TencentResponseError
	err := json.NewDecoder(resp.Body).Decode(tencentError)
	if err != nil {
		return nil
	}

	return errorHandle(tencentError)
}

// 错误处理
func errorHandle(tencentError *TencentResponseError) *types.OpenAIError {
	if tencentError.Error.Code == 0 {
		return nil
	}
	return &types.OpenAIError{
		Message: tencentError.Error.Message,
		Type:    "tencent_error",
		Code:    tencentError.Error.Code,
	}
}

// 获取请求头
func (p *TencentProvider) GetRequestHeaders() (headers map[string]string) {
	headers = make(map[string]string)
	p.CommonRequestHeaders(headers)

	return headers
}

func (p *TencentProvider) parseTencentConfig(config string) (appId int64, secretId string, secretKey string, err error) {
	parts := strings.Split(config, "|")
	if len(parts) != 3 {
		err = errors.New("invalid tencent config")
		return
	}
	appId, err = strconv.ParseInt(parts[0], 10, 64)
	secretId = parts[1]
	secretKey = parts[2]
	return
}

func (p *TencentProvider) getTencentSign(req TencentChatRequest) string {
	apiKey := p.Channel.Key
	appId, secretId, secretKey, err := p.parseTencentConfig(apiKey)
	if err != nil {
		return ""
	}
	req.AppId = appId
	req.SecretId = secretId

	params := make([]string, 0)
	params = append(params, "app_id="+strconv.FormatInt(req.AppId, 10))
	params = append(params, "secret_id="+req.SecretId)
	params = append(params, "timestamp="+strconv.FormatInt(req.Timestamp, 10))
	params = append(params, "query_id="+req.QueryID)
	params = append(params, "temperature="+strconv.FormatFloat(req.Temperature, 'f', -1, 64))
	params = append(params, "top_p="+strconv.FormatFloat(req.TopP, 'f', -1, 64))
	params = append(params, "stream="+strconv.Itoa(req.Stream))
	params = append(params, "expired="+strconv.FormatInt(req.Expired, 10))

	var messageStr string
	for _, msg := range req.Messages {
		messageStr += fmt.Sprintf(`{"role":"%s","content":"%s"},`, msg.Role, msg.Content)
	}
	messageStr = strings.TrimSuffix(messageStr, ",")
	params = append(params, "messages=["+messageStr+"]")

	sort.Strings(params)
	url := "hunyuan.cloud.tencent.com/hyllm/v1/chat/completions?" + strings.Join(params, "&")
	mac := hmac.New(sha1.New, []byte(secretKey))
	signURL := url
	mac.Write([]byte(signURL))
	sign := mac.Sum([]byte(nil))
	return base64.StdEncoding.EncodeToString(sign)
}
