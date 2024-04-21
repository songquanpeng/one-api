package meta

import (
	"github.com/gin-gonic/gin"
	"github.com/songquanpeng/one-api/common/ctxkey"
	"github.com/songquanpeng/one-api/relay/adaptor/azure"
	"github.com/songquanpeng/one-api/relay/channeltype"
	"github.com/songquanpeng/one-api/relay/relaymode"
	"strings"
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
	APIVersion      string
	APIKey          string
	APIType         int
	Config          map[string]string
	IsStream        bool
	OriginModelName string
	ActualModelName string
	RequestURLPath  string
	PromptTokens    int // only for DoResponse
}

func GetByContext(c *gin.Context) *Meta {
	meta := Meta{
		Mode:           relaymode.GetByPath(c.Request.URL.Path),
		ChannelType:    c.GetInt(ctxkey.Channel),
		ChannelId:      c.GetInt(ctxkey.ChannelId),
		TokenId:        c.GetInt(ctxkey.TokenId),
		TokenName:      c.GetString(ctxkey.TokenName),
		UserId:         c.GetInt(ctxkey.Id),
		Group:          c.GetString(ctxkey.Group),
		ModelMapping:   c.GetStringMapString(ctxkey.ModelMapping),
		BaseURL:        c.GetString(ctxkey.BaseURL),
		APIVersion:     c.GetString(ctxkey.ConfigAPIVersion),
		APIKey:         strings.TrimPrefix(c.Request.Header.Get("Authorization"), "Bearer "),
		Config:         nil,
		RequestURLPath: c.Request.URL.String(),
	}
	if meta.ChannelType == channeltype.Azure {
		meta.APIVersion = azure.GetAPIVersion(c)
	}
	if meta.BaseURL == "" {
		meta.BaseURL = channeltype.ChannelBaseURLs[meta.ChannelType]
	}
	meta.APIType = channeltype.ToAPIType(meta.ChannelType)
	return &meta
}
