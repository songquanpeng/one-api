package meta

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/songquanpeng/one-api/common/ctxkey"
	"github.com/songquanpeng/one-api/model"
	"github.com/songquanpeng/one-api/relay/channeltype"
	"github.com/songquanpeng/one-api/relay/relaymode"
)

type Meta struct {
	Mode            int
	ChannelType     int
	ChannelId       int
	TokenId         int
	TokenName       string
	UserId          int
	Group           string
	ModelMapping    map[string]string
	BaseURL         string
	APIKey          string
	APIType         int
	Config          model.ChannelConfig
	IsStream        bool
	OriginModelName string
	ActualModelName string
	RequestURLPath  string
	PromptTokens    int // only for DoResponse
}

func (m *Meta) ToLogrusFields() map[string]interface{} {
	return map[string]interface{}{
		"mode":              m.Mode,
		"channel_type":      m.ChannelType,
		"channel_id":        m.ChannelId,
		"token_id":          m.TokenId,
		"token_name":        m.TokenName,
		"user_id":           m.UserId,
		"group":             m.Group,
		"model_mapping":     m.ModelMapping,
		"base_url":          m.BaseURL,
		"api_key":           m.APIKey,
		"api_type":          m.APIType,
		"config":            m.Config,
		"is_stream":         m.IsStream,
		"origin_model_name": m.OriginModelName,
		"actual_model_name": m.ActualModelName,
		"request_url_path":  m.RequestURLPath,
		"prompt_tokens":     m.PromptTokens,
	}

}

func GetByContext(c *gin.Context) *Meta {
	meta := Meta{
		Mode:            relaymode.GetByPath(c.Request.URL.Path),
		ChannelType:     c.GetInt(ctxkey.Channel),
		ChannelId:       c.GetInt(ctxkey.ChannelId),
		TokenId:         c.GetInt(ctxkey.TokenId),
		TokenName:       c.GetString(ctxkey.TokenName),
		UserId:          c.GetInt(ctxkey.Id),
		Group:           c.GetString(ctxkey.Group),
		ModelMapping:    c.GetStringMapString(ctxkey.ModelMapping),
		OriginModelName: c.GetString(ctxkey.RequestModel),
		BaseURL:         c.GetString(ctxkey.BaseURL),
		APIKey:          strings.TrimPrefix(c.Request.Header.Get("Authorization"), "Bearer "),
		RequestURLPath:  c.Request.URL.String(),
	}
	cfg, ok := c.Get(ctxkey.Config)
	if ok {
		meta.Config = cfg.(model.ChannelConfig)
	}
	if meta.BaseURL == "" {
		meta.BaseURL = channeltype.ChannelBaseURLs[meta.ChannelType]
	}
	meta.APIType = channeltype.ToAPIType(meta.ChannelType)
	return &meta
}
