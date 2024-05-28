package zhipu

import (
	"encoding/json"
	"fmt"
	"net/http"
	"one-api/common/logger"
	"one-api/common/requester"
	"one-api/model"
	"one-api/providers/base"
	"one-api/types"
	"strings"
	"sync"
	"time"

	"github.com/golang-jwt/jwt"
)

var zhipuTokens sync.Map
var expSeconds int64 = 24 * 3600

type ZhipuProviderFactory struct{}

// 创建 ZhipuProvider
func (f ZhipuProviderFactory) Create(channel *model.Channel) base.ProviderInterface {
	return &ZhipuProvider{
		BaseProvider: base.BaseProvider{
			Config:    getConfig(),
			Channel:   channel,
			Requester: requester.NewHTTPRequester(*channel.Proxy, requestErrorHandle),
		},
	}
}

type ZhipuProvider struct {
	base.BaseProvider
}

func getConfig() base.ProviderConfig {
	return base.ProviderConfig{
		BaseURL:           "https://open.bigmodel.cn/api/paas/v4",
		ChatCompletions:   "/chat/completions",
		Embeddings:        "/embeddings",
		ImagesGenerations: "/images/generations",
	}
}

// 请求错误处理
func requestErrorHandle(resp *http.Response) *types.OpenAIError {
	zhipuError := &ZhipuResponseError{}
	err := json.NewDecoder(resp.Body).Decode(zhipuError)
	if err != nil {
		return nil
	}

	return errorHandle(&zhipuError.Error)
}

// 错误处理
func errorHandle(zhipuError *ZhipuError) *types.OpenAIError {
	if zhipuError.Message == "" {
		return nil
	}
	return &types.OpenAIError{
		Message: zhipuError.Message,
		Type:    "zhipu_error",
		Code:    zhipuError.Code,
	}
}

// 获取请求头
func (p *ZhipuProvider) GetRequestHeaders() (headers map[string]string) {
	headers = make(map[string]string)
	p.CommonRequestHeaders(headers)
	headers["Authorization"] = p.getZhipuToken()
	return headers
}

// 获取完整请求 URL
func (p *ZhipuProvider) GetFullRequestURL(requestURL string) string {
	baseURL := strings.TrimSuffix(p.GetBaseURL(), "/")

	return fmt.Sprintf("%s%s", baseURL, requestURL)
}

func (p *ZhipuProvider) getZhipuToken() string {
	apikey := p.Channel.Key
	data, ok := zhipuTokens.Load(apikey)
	if ok {
		tokenData := data.(zhipuTokenData)
		if time.Now().Before(tokenData.ExpiryTime) {
			return tokenData.Token
		}
	}

	split := strings.Split(apikey, ".")
	if len(split) != 2 {
		logger.SysError("invalid zhipu key: " + apikey)
		return ""
	}

	id := split[0]
	secret := split[1]

	expMillis := time.Now().Add(time.Duration(expSeconds)*time.Second).UnixNano() / 1e6
	expiryTime := time.Now().Add(time.Duration(expSeconds) * time.Second)

	timestamp := time.Now().UnixNano() / 1e6

	payload := jwt.MapClaims{
		"api_key":   id,
		"exp":       expMillis,
		"timestamp": timestamp,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, payload)

	token.Header["alg"] = "HS256"
	token.Header["sign_type"] = "SIGN"

	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		return ""
	}

	zhipuTokens.Store(apikey, zhipuTokenData{
		Token:      tokenString,
		ExpiryTime: expiryTime,
	})

	return tokenString
}

func convertRole(roleName string) string {
	switch roleName {
	case types.ChatMessageRoleFunction:
		return types.ChatMessageRoleTool
	case types.ChatMessageRoleTool, types.ChatMessageRoleSystem, types.ChatMessageRoleAssistant:
		return roleName
	default:
		return types.ChatMessageRoleUser
	}
}

func convertTopP(topP float64) float64 {
	// 检测 topP 是否在 0-1 之间 如果等于0 设为0.1 如果大于等于1 设为0.9
	if topP <= 0 {
		return 0.1
	} else if topP >= 1 {
		return 0.9
	}
	return topP
}
