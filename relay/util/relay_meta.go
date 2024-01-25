package util

import (
	"github.com/gin-gonic/gin"
	"one-api/common"
	"strings"
)

type RelayMeta struct {
	ChannelType  int
	ChannelId    int
	TokenId      int
	TokenName    string
	UserId       int
	Group        string
	ModelMapping map[string]string
	BaseURL      string
	APIVersion   string
	APIKey       string
	Config       map[string]string
}

func GetRelayMeta(c *gin.Context) *RelayMeta {
	meta := RelayMeta{
		ChannelType:  c.GetInt("channel"),
		ChannelId:    c.GetInt("channel_id"),
		TokenId:      c.GetInt("token_id"),
		TokenName:    c.GetString("token_name"),
		UserId:       c.GetInt("id"),
		Group:        c.GetString("group"),
		ModelMapping: c.GetStringMapString("model_mapping"),
		BaseURL:      c.GetString("base_url"),
		APIVersion:   c.GetString("api_version"),
		APIKey:       strings.TrimPrefix(c.Request.Header.Get("Authorization"), "Bearer "),
		Config:       nil,
	}
	if meta.ChannelType == common.ChannelTypeAzure {
		meta.APIVersion = GetAzureAPIVersion(c)
	}
	if meta.BaseURL == "" {
		meta.BaseURL = common.ChannelBaseURLs[meta.ChannelType]
	}
	return &meta
}
