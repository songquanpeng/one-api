package xiaomimimo

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/songquanpeng/one-api/relay/adaptor"
	"github.com/songquanpeng/one-api/relay/adaptor/openai"
	"github.com/songquanpeng/one-api/relay/meta"
)

// Adaptor 小米 MiMo 适配器。
// MiMo 开放平台为 OpenAI 兼容协议（https://api.xiaomimimo.com/v1），
// 除标准的 Authorization: Bearer 之外，平台同时接受 api-key 头认证，
// 此处双写以兼容两种客户端配置方式。
type Adaptor struct {
	openai.Adaptor
}

func (a *Adaptor) SetupRequestHeader(c *gin.Context, req *http.Request, meta *meta.Meta) error {
	adaptor.SetupCommonRequestHeader(c, req, meta)
	req.Header.Set("Authorization", "Bearer "+meta.APIKey)
	req.Header.Set("api-key", meta.APIKey)
	return nil
}

func (a *Adaptor) GetModelList() []string {
	return ModelList
}

func (a *Adaptor) GetChannelName() string {
	return "xiaomi-mimo"
}
