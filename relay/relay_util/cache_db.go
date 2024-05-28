package relay_util

import (
	"errors"
	"one-api/common/utils"
	"one-api/model"
	"time"
)

type ChatCacheDB struct{}

func (db *ChatCacheDB) Get(hash string, userId int) *ChatCacheProps {
	cache, _ := model.GetChatCache(hash, userId)
	if cache == nil {
		return nil
	}

	props, err := utils.UnmarshalString[ChatCacheProps](cache.Data)
	if err != nil {
		return nil
	}

	return &props
}

func (db *ChatCacheDB) Set(hash string, props *ChatCacheProps, expire int64) error {
	return SetCacheDB(hash, props, expire)
}

func SetCacheDB(hash string, props *ChatCacheProps, expire int64) error {
	data := utils.Marshal(props)
	if data == "" {
		return errors.New("marshal error")
	}

	expire = expire * 60
	expire += time.Now().Unix()

	cache := &model.ChatCache{
		Hash:       hash,
		UserId:     props.UserId,
		Data:       data,
		Expiration: expire,
	}

	return cache.Insert()
}
