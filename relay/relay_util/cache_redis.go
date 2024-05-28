package relay_util

import (
	"errors"
	"fmt"
	"one-api/common"
	"one-api/common/utils"
	"time"
)

type ChatCacheRedis struct{}

var chatCacheKey = "chat_cache"

func (r *ChatCacheRedis) Get(hash string, userId int) *ChatCacheProps {
	cache, err := common.RedisGet(r.getKey(hash, userId))
	if err != nil {
		return nil
	}

	props, err := utils.UnmarshalString[ChatCacheProps](cache)
	if err != nil {
		return nil
	}

	return &props
}

func (r *ChatCacheRedis) Set(hash string, props *ChatCacheProps, expire int64) error {

	if !props.Cache {
		return nil
	}

	data := utils.Marshal(&props)
	if data == "" {
		return errors.New("marshal error")
	}

	return common.RedisSet(r.getKey(hash, props.UserId), data, time.Duration(expire)*time.Minute)
}

func (r *ChatCacheRedis) getKey(hash string, userId int) string {
	return fmt.Sprintf("%s:%d:%s", chatCacheKey, userId, hash)
}
