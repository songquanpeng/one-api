package model

import (
	"encoding/json"
	"fmt"
	"one-api/common"
	"sync"
	"time"
)

const (
	TokenCacheSeconds        = 60 * 60
	UserId2GroupCacheSeconds = 60 * 60
)

func CacheGetTokenByKey(key string) (*Token, error) {
	var token Token
	if !common.RedisEnabled {
		err := DB.Where("`key` = ?", key).First(&token).Error
		return &token, err
	}
	tokenObjectString, err := common.RedisGet(fmt.Sprintf("token:%s", key))
	if err != nil {
		err := DB.Where("`key` = ?", key).First(&token).Error
		if err != nil {
			return nil, err
		}
		jsonBytes, err := json.Marshal(token)
		if err != nil {
			return nil, err
		}
		err = common.RedisSet(fmt.Sprintf("token:%s", key), string(jsonBytes), TokenCacheSeconds*time.Second)
		if err != nil {
			common.SysError("Redis set token error: " + err.Error())
		}
	}
	err = json.Unmarshal([]byte(tokenObjectString), &token)
	return &token, err
}

func CacheGetUserGroup(id int) (group string, err error) {
	if !common.RedisEnabled {
		return GetUserGroup(id)
	}
	group, err = common.RedisGet(fmt.Sprintf("user_group:%d", id))
	if err != nil {
		group, err = GetUserGroup(id)
		if err != nil {
			return "", err
		}
		err = common.RedisSet(fmt.Sprintf("user_group:%d", id), group, UserId2GroupCacheSeconds*time.Second)
		if err != nil {
			common.SysError("Redis set user group error: " + err.Error())
		}
	}
	return group, err
}

var channelId2channel map[int]*Channel
var channelSyncLock sync.RWMutex
var group2model2channels map[string]map[string][]*Channel

func InitChannelCache() {
	channelSyncLock.Lock()
	defer channelSyncLock.Unlock()
	channelId2channel = make(map[int]*Channel)
	var channels []*Channel
	DB.Find(&channels)
	for _, channel := range channels {
		channelId2channel[channel.Id] = channel
	}
	var abilities []*Ability
	DB.Find(&abilities)
	groups := make(map[string]bool)
	for _, ability := range abilities {
		groups[ability.Group] = true
	}
	group2model2channels = make(map[string]map[string][]*Channel)
	for group := range groups {
		group2model2channels[group] = make(map[string][]*Channel)
		// TODO: implement this
	}
}

func SyncChannelCache(frequency int) {
	for {
		time.Sleep(time.Duration(frequency) * time.Second)
		common.SysLog("Syncing channels from database")
		InitChannelCache()
	}
}

func CacheGetRandomSatisfiedChannel(group string, model string) (*Channel, error) {
	if !common.RedisEnabled {
		return GetRandomSatisfiedChannel(group, model)
	}
	// TODO: implement this
	return nil, nil
}
