package zhipu

import (
	"fmt"
	"one-api/common"
	"one-api/providers/base"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
)

var zhipuTokens sync.Map
var expSeconds int64 = 24 * 3600

type ZhipuProviderFactory struct{}

// 创建 ZhipuProvider
func (f ZhipuProviderFactory) Create(c *gin.Context) base.ProviderInterface {
	return &ZhipuProvider{
		BaseProvider: base.BaseProvider{
			BaseURL:         "https://open.bigmodel.cn",
			ChatCompletions: "/api/paas/v3/model-api",
			Context:         c,
		},
	}
}

type ZhipuProvider struct {
	base.BaseProvider
}

// 获取请求头
func (p *ZhipuProvider) GetRequestHeaders() (headers map[string]string) {
	headers = make(map[string]string)
	p.CommonRequestHeaders(headers)
	headers["Authorization"] = p.getZhipuToken()

	return headers
}

// 获取完整请求 URL
func (p *ZhipuProvider) GetFullRequestURL(requestURL string, modelName string) string {
	baseURL := strings.TrimSuffix(p.GetBaseURL(), "/")

	return fmt.Sprintf("%s%s/%s", baseURL, requestURL, modelName)
}

func (p *ZhipuProvider) getZhipuToken() string {
	apikey := p.Context.GetString("api_key")
	data, ok := zhipuTokens.Load(apikey)
	if ok {
		tokenData := data.(zhipuTokenData)
		if time.Now().Before(tokenData.ExpiryTime) {
			return tokenData.Token
		}
	}

	split := strings.Split(apikey, ".")
	if len(split) != 2 {
		common.SysError("invalid zhipu key: " + apikey)
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
