package tencent

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"errors"
	"fmt"
	"one-api/providers/base"
	"sort"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

type TencentProviderFactory struct{}

// 创建 TencentProvider
func (f TencentProviderFactory) Create(c *gin.Context) base.ProviderInterface {
	return &TencentProvider{
		BaseProvider: base.BaseProvider{
			BaseURL:         "https://hunyuan.cloud.tencent.com",
			ChatCompletions: "/hyllm/v1/chat/completions",
			Context:         c,
		},
	}
}

type TencentProvider struct {
	base.BaseProvider
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

	sort.Sort(sort.StringSlice(params))
	url := "hunyuan.cloud.tencent.com/hyllm/v1/chat/completions?" + strings.Join(params, "&")
	mac := hmac.New(sha1.New, []byte(secretKey))
	signURL := url
	mac.Write([]byte(signURL))
	sign := mac.Sum([]byte(nil))
	return base64.StdEncoding.EncodeToString(sign)
}
