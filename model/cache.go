package model

import (
	"encoding/json"
	"errors"
	"fmt"
	"one-api/common"
	"strconv"
	"sync"
	"time"

	"github.com/mroth/weightedrand/v2"
)

var (
	TokenCacheSeconds         = common.SyncFrequency
	UserId2GroupCacheSeconds  = common.SyncFrequency
	UserId2QuotaCacheSeconds  = common.SyncFrequency
	UserId2StatusCacheSeconds = common.SyncFrequency
)

func CacheGetTokenByKey(key string) (*Token, error) {
	keyCol := "`key`"
	if common.UsingPostgreSQL {
		keyCol = `"key"`
	}
	var token Token
	if !common.RedisEnabled {
		err := DB.Where(keyCol+" = ?", key).First(&token).Error
		return &token, err
	}
	tokenObjectString, err := common.RedisGet(fmt.Sprintf("token:%s", key))
	if err != nil {
		err := DB.Where(keyCol+" = ?", key).First(&token).Error
		if err != nil {
			return nil, err
		}
		jsonBytes, err := json.Marshal(token)
		if err != nil {
			return nil, err
		}
		err = common.RedisSet(fmt.Sprintf("token:%s", key), string(jsonBytes), time.Duration(TokenCacheSeconds)*time.Second)
		if err != nil {
			common.SysError("Redis set token error: " + err.Error())
		}
		return &token, nil
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
		err = common.RedisSet(fmt.Sprintf("user_group:%d", id), group, time.Duration(UserId2GroupCacheSeconds)*time.Second)
		if err != nil {
			common.SysError("Redis set user group error: " + err.Error())
		}
	}
	return group, err
}

func CacheGetUserQuota(id int) (quota int, err error) {
	if !common.RedisEnabled {
		return GetUserQuota(id)
	}
	quotaString, err := common.RedisGet(fmt.Sprintf("user_quota:%d", id))
	if err != nil {
		quota, err = GetUserQuota(id)
		if err != nil {
			return 0, err
		}
		err = common.RedisSet(fmt.Sprintf("user_quota:%d", id), fmt.Sprintf("%d", quota), time.Duration(UserId2QuotaCacheSeconds)*time.Second)
		if err != nil {
			common.SysError("Redis set user quota error: " + err.Error())
		}
		return quota, err
	}
	quota, err = strconv.Atoi(quotaString)
	return quota, err
}

func CacheUpdateUserQuota(id int) error {
	if !common.RedisEnabled {
		return nil
	}
	quota, err := GetUserQuota(id)
	if err != nil {
		return err
	}
	err = common.RedisSet(fmt.Sprintf("user_quota:%d", id), fmt.Sprintf("%d", quota), time.Duration(UserId2QuotaCacheSeconds)*time.Second)
	return err
}

func CacheDecreaseUserQuota(id int, quota int) error {
	if !common.RedisEnabled {
		return nil
	}
	err := common.RedisDecrease(fmt.Sprintf("user_quota:%d", id), int64(quota))
	return err
}

func CacheIsUserEnabled(userId int) (bool, error) {
	if !common.RedisEnabled {
		return IsUserEnabled(userId)
	}
	enabled, err := common.RedisGet(fmt.Sprintf("user_enabled:%d", userId))
	if err == nil {
		return enabled == "1", nil
	}

	userEnabled, err := IsUserEnabled(userId)
	if err != nil {
		return false, err
	}
	enabled = "0"
	if userEnabled {
		enabled = "1"
	}
	err = common.RedisSet(fmt.Sprintf("user_enabled:%d", userId), enabled, time.Duration(UserId2StatusCacheSeconds)*time.Second)
	if err != nil {
		common.SysError("Redis set user enabled error: " + err.Error())
	}
	return userEnabled, err
}

var group2model2channels map[string]map[string]*weightedrand.Chooser[*Channel, int]
var channelSyncLock sync.RWMutex

func InitChannelCache() {
	newChannelId2channel := make(map[int]*Channel)
	var channels []*Channel
	DB.Where("status = ?", common.ChannelStatusEnabled).Find(&channels)
	for _, channel := range channels {
		newChannelId2channel[channel.Id] = channel
	}
	var abilities []*Ability
	DB.Find(&abilities)
	groups := make(map[string]bool)
	for _, ability := range abilities {
		groups[ability.Group] = true
	}
	newGroup2model2channels := make(map[string]map[string][]weightedrand.Choice[*Channel, int])
	for group := range groups {
		newGroup2model2channels[group] = make(map[string][]weightedrand.Choice[*Channel, int])
	}
	for _, channel := range channels {
		groups := channel.GetGroups()
		for _, group := range groups {
			models := channel.GetModels()
			weightMapping := channel.GetWeightMapping()
			for _, model := range models {
				if _, ok := newGroup2model2channels[group][model]; !ok {
					newGroup2model2channels[group][model] = make([]weightedrand.Choice[*Channel, int], 0)
				}
				weight, ok := weightMapping[model]
				if weight < 0 || !ok {
					// use default value if:
					// weight < 0: invalid
					// !ok: weight not set
					weight = common.DefaultWeight
				}
				newGroup2model2channels[group][model] = append(newGroup2model2channels[group][model], weightedrand.NewChoice(channel, weight))
			}
		}
	}

	// sort by priority
	m := make(map[string]map[string]*weightedrand.Chooser[*Channel, int])
	for group, model2channels := range newGroup2model2channels {
		m[group] = make(map[string]*weightedrand.Chooser[*Channel, int])
		for model, channels := range model2channels {
			if len(channels) == 0 {
				common.SysError(fmt.Sprintf("no channel found for group %s model %s", group, model))
				continue
			}
			c, err := weightedrand.NewChooser(channels...)
			if err != nil {
				common.SysError(fmt.Sprintf("failed to create chooser: %s", err.Error()))
				continue
			}
			m[group][model] = c
		}
	}

	channelSyncLock.Lock()
	group2model2channels = m
	channelSyncLock.Unlock()
	common.SysLog("channels synced from database")
}

func SyncChannelCache(frequency int) {
	for {
		time.Sleep(time.Duration(frequency) * time.Second)
		common.SysLog("syncing channels from database")
		InitChannelCache()
	}
}

func CacheGetRandomSatisfiedChannel(group string, model string) (*Channel, error) {
	if !common.MemoryCacheEnabled {
		return GetRandomSatisfiedChannel(group, model)
	}
	channelSyncLock.RLock()
	defer channelSyncLock.RUnlock()
	channels := group2model2channels[group][model]
	if channels == nil {
		return nil, errors.New("channel not found")
	}
	return channels.Pick(), nil
}
