package xunfei

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"net/url"
	"one-api/common/logger"
	"one-api/common/requester"
	"one-api/model"
	"one-api/providers/base"
	"one-api/types"
	"strings"
	"time"
)

type XunfeiProviderFactory struct{}

// 创建 XunfeiProvider
func (f XunfeiProviderFactory) Create(channel *model.Channel) base.ProviderInterface {
	return &XunfeiProvider{
		BaseProvider: base.BaseProvider{
			Config:    getConfig(),
			Channel:   channel,
			Requester: requester.NewHTTPRequester(*channel.Proxy, nil),
		},
		wsRequester: requester.NewWSRequester(*channel.Proxy),
	}
}

// https://www.xfyun.cn/doc/spark/Web.html
type XunfeiProvider struct {
	base.BaseProvider
	domain      string
	apiId       string
	wsRequester *requester.WSRequester
}

func getConfig() base.ProviderConfig {
	return base.ProviderConfig{
		BaseURL:         "wss://spark-api.xf-yun.com",
		ChatCompletions: "/",
	}
}

// 错误处理
func errorHandle(xunfeiError *XunfeiChatResponse) *types.OpenAIError {
	if xunfeiError.Header.Code == 0 {
		return nil
	}
	return &types.OpenAIError{
		Message: xunfeiError.Header.Message,
		Type:    "xunfei_error",
		Code:    xunfeiError.Header.Code,
	}
}

// 获取请求头
func (p *XunfeiProvider) GetRequestHeaders() (headers map[string]string) {
	headers = make(map[string]string)
	return headers
}

// 获取完整请求 URL
func (p *XunfeiProvider) GetFullRequestURL(modelName string) string {
	splits := strings.Split(p.Channel.Key, "|")
	if len(splits) != 3 {
		return ""
	}
	domain, authUrl := p.getXunfeiAuthUrl(splits[2], splits[1], modelName)

	p.domain = domain
	p.apiId = splits[0]

	return authUrl
}

func (p *XunfeiProvider) getAPIVersion(modelName string) string {
	query := p.Context.Request.URL.Query()
	apiVersion := query.Get("api-version")
	if apiVersion != "" {
		return apiVersion
	}
	parts := strings.Split(modelName, "-")
	if len(parts) == 2 {
		apiVersion = parts[1]
		return apiVersion
	}

	apiVersion = p.Channel.Other
	if apiVersion != "" {
		return apiVersion
	}
	apiVersion = "v1.1"

	logger.SysLog("api_version not found, use default: " + apiVersion)
	return apiVersion
}

// https://www.xfyun.cn/doc/spark/Web.html#_1-%E6%8E%A5%E5%8F%A3%E8%AF%B4%E6%98%8E
func apiVersion2domain(apiVersion string) string {
	switch apiVersion {
	case "v1.1":
		return "general"
	case "v2.1":
		return "generalv2"
	case "v3.1":
		return "generalv3"
	case "v3.5":
		return "generalv3.5"
	}
	return "general" + apiVersion
}

func (p *XunfeiProvider) getXunfeiAuthUrl(apiKey string, apiSecret string, modelName string) (string, string) {
	apiVersion := p.getAPIVersion(modelName)
	domain := apiVersion2domain(apiVersion)

	authUrl := p.buildXunfeiAuthUrl(fmt.Sprintf("%s/%s/chat", p.Config.BaseURL, apiVersion), apiKey, apiSecret)
	return domain, authUrl
}

func (p *XunfeiProvider) buildXunfeiAuthUrl(hostUrl string, apiKey, apiSecret string) string {
	HmacWithShaToBase64 := func(algorithm, data, key string) string {
		mac := hmac.New(sha256.New, []byte(key))
		mac.Write([]byte(data))
		encodeData := mac.Sum(nil)
		return base64.StdEncoding.EncodeToString(encodeData)
	}
	ul, err := url.Parse(hostUrl)
	if err != nil {
		logger.SysError("url parse error: " + err.Error())
		return ""
	}
	date := time.Now().UTC().Format(time.RFC1123)
	signString := []string{"host: " + ul.Host, "date: " + date, "GET " + ul.Path + " HTTP/1.1"}
	sign := strings.Join(signString, "\n")
	sha := HmacWithShaToBase64("hmac-sha256", sign, apiSecret)
	authUrl := fmt.Sprintf("hmac username=\"%s\", algorithm=\"%s\", headers=\"%s\", signature=\"%s\"", apiKey,
		"hmac-sha256", "host date request-line", sha)
	authorization := base64.StdEncoding.EncodeToString([]byte(authUrl))
	v := url.Values{}
	v.Add("host", ul.Host)
	v.Add("date", date)
	v.Add("authorization", authorization)
	callUrl := hostUrl + "?" + v.Encode()
	return callUrl
}
