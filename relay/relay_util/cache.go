package relay_util

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"github.com/gin-gonic/gin"
	"one-api/common"
	"one-api/common/config"
	"one-api/common/utils"
)

type ChatCacheProps struct {
	UserId           int    `json:"user_id"`
	TokenId          int    `json:"token_id"`
	ChannelID        int    `json:"channel_id"`
	PromptTokens     int    `json:"prompt_tokens"`
	CompletionTokens int    `json:"completion_tokens"`
	ModelName        string `json:"model_name"`
	Response         string `json:"response"`

	Hash   string      `json:"-"`
	Cache  bool        `json:"-"`
	Driver CacheDriver `json:"-"`
}

type CacheDriver interface {
	Get(hash string, userId int) *ChatCacheProps
	Set(hash string, props *ChatCacheProps, expire int64) error
}

func NewChatCacheProps(c *gin.Context, allow bool) *ChatCacheProps {
	props := &ChatCacheProps{
		Cache: false,
	}

	if !allow {
		return props
	}

	if config.ChatCacheEnabled && c.GetBool("chat_cache") {
		props.Cache = true
	}

	if common.RedisEnabled {
		props.Driver = &ChatCacheRedis{}
	} else {
		props.Driver = &ChatCacheDB{}
	}

	props.UserId = c.GetInt("id")
	props.TokenId = c.GetInt("token_id")

	return props
}

func (p *ChatCacheProps) SetHash(request any) {
	if !p.needCache() || request == nil {
		return
	}

	p.hash(utils.Marshal(request))
}

func (p *ChatCacheProps) SetResponse(response any) {
	if !p.needCache() || response == nil {
		return
	}

	if str, ok := response.(string); ok {
		p.Response += str
		return
	}

	responseStr := utils.Marshal(response)
	if responseStr == "" {
		return
	}

	p.Response = responseStr
}

func (p *ChatCacheProps) NoCache() {
	p.Cache = false
}

func (p *ChatCacheProps) StoreCache(channelId, promptTokens, completionTokens int, modelName string) error {
	if !p.needCache() || p.Response == "" {
		return nil
	}

	p.ChannelID = channelId
	p.PromptTokens = promptTokens
	p.CompletionTokens = completionTokens
	p.ModelName = modelName

	return p.Driver.Set(p.getHash(), p, int64(config.ChatCacheExpireMinute))
}

func (p *ChatCacheProps) GetCache() *ChatCacheProps {
	if !p.needCache() {
		return nil
	}

	return p.Driver.Get(p.getHash(), p.UserId)
}

func (p *ChatCacheProps) needCache() bool {
	return config.ChatCacheEnabled && p.Cache
}

func (p *ChatCacheProps) getHash() string {
	return p.Hash
}

func (p *ChatCacheProps) hash(request string) {
	hash := md5.Sum([]byte(fmt.Sprintf("%d-%d-%s", p.UserId, p.TokenId, request)))
	p.Hash = hex.EncodeToString(hash[:])
}
