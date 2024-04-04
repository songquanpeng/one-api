package model

import (
	"errors"
	"math/rand"
	"one-api/common"
	"strings"
	"sync"
	"time"
)

type ChannelChoice struct {
	Channel       *Channel
	CooldownsTime int64
}

type ChannelsChooser struct {
	sync.RWMutex
	Channels map[int]*ChannelChoice
	Rule     map[string]map[string][][]int // group -> model -> priority -> channelIds
	Match    []string
}

func (cc *ChannelsChooser) Cooldowns(channelId int) bool {
	if common.RetryCooldownSeconds == 0 {
		return false
	}
	cc.Lock()
	defer cc.Unlock()
	if _, ok := cc.Channels[channelId]; !ok {
		return false
	}

	cc.Channels[channelId].CooldownsTime = time.Now().Unix() + int64(common.RetryCooldownSeconds)
	return true
}

func (cc *ChannelsChooser) balancer(channelIds []int) *Channel {
	nowTime := time.Now().Unix()
	totalWeight := 0

	validChannels := make([]*ChannelChoice, 0, len(channelIds))
	for _, channelId := range channelIds {
		if choice, ok := cc.Channels[channelId]; ok && choice.CooldownsTime < nowTime {
			weight := int(*choice.Channel.Weight)
			totalWeight += weight
			validChannels = append(validChannels, choice)
		}
	}

	if len(validChannels) == 0 {
		return nil
	}

	if len(validChannels) == 1 {
		return validChannels[0].Channel
	}

	choiceWeight := rand.Intn(totalWeight)
	for _, choice := range validChannels {
		weight := int(*choice.Channel.Weight)
		choiceWeight -= weight
		if choiceWeight < 0 {
			return choice.Channel
		}
	}

	return nil
}

func (cc *ChannelsChooser) Next(group, modelName string) (*Channel, error) {
	cc.RLock()
	defer cc.RUnlock()
	if _, ok := cc.Rule[group]; !ok {
		return nil, errors.New("group not found")
	}

	channelsPriority, ok := cc.Rule[group][modelName]
	if !ok {
		matchModel := common.GetModelsWithMatch(&cc.Match, modelName)
		channelsPriority, ok = cc.Rule[group][matchModel]
		if !ok {
			return nil, errors.New("model not found")
		}
	}

	if len(channelsPriority) == 0 {
		return nil, errors.New("channel not found")
	}

	for _, priority := range channelsPriority {
		channel := cc.balancer(priority)
		if channel != nil {
			return channel, nil
		}
	}

	return nil, errors.New("channel not found")
}

func (cc *ChannelsChooser) GetGroupModels(group string) ([]string, error) {
	cc.RLock()
	defer cc.RUnlock()

	if _, ok := cc.Rule[group]; !ok {
		return nil, errors.New("group not found")
	}

	models := make([]string, 0, len(cc.Rule[group]))
	for model := range cc.Rule[group] {
		models = append(models, model)
	}

	return models, nil
}

func (cc *ChannelsChooser) GetChannel(channelId int) *Channel {
	cc.RLock()
	defer cc.RUnlock()

	if choice, ok := cc.Channels[channelId]; ok {
		return choice.Channel
	}

	return nil
}

var ChannelGroup = ChannelsChooser{}

func (cc *ChannelsChooser) Load() {
	var channels []*Channel
	DB.Where("status = ?", common.ChannelStatusEnabled).Find(&channels)

	abilities, err := GetAbilityChannelGroup()
	if err != nil {
		common.SysLog("get enabled abilities failed: " + err.Error())
		return
	}

	newGroup := make(map[string]map[string][][]int)
	newChannels := make(map[int]*ChannelChoice)
	newMatch := make(map[string]bool)

	for _, channel := range channels {
		if *channel.Weight == 0 {
			channel.Weight = &common.DefaultChannelWeight
		}
		newChannels[channel.Id] = &ChannelChoice{
			Channel:       channel,
			CooldownsTime: 0,
		}
	}

	for _, ability := range abilities {
		if _, ok := newGroup[ability.Group]; !ok {
			newGroup[ability.Group] = make(map[string][][]int)
		}

		if _, ok := newGroup[ability.Group][ability.Model]; !ok {
			newGroup[ability.Group][ability.Model] = make([][]int, 0)
		}

		// 如果是以 *结尾的 model名称
		if strings.HasSuffix(ability.Model, "*") {
			if _, ok := newMatch[ability.Model]; !ok {
				newMatch[ability.Model] = true
			}
		}

		var priorityIds []int
		// 逗号分割 ability.ChannelId
		channelIds := strings.Split(ability.ChannelIds, ",")
		for _, channelId := range channelIds {
			priorityIds = append(priorityIds, common.String2Int(channelId))
		}

		newGroup[ability.Group][ability.Model] = append(newGroup[ability.Group][ability.Model], priorityIds)
	}

	newMatchList := make([]string, 0, len(newMatch))
	for match := range newMatch {
		newMatchList = append(newMatchList, match)
	}

	cc.Lock()
	cc.Rule = newGroup
	cc.Channels = newChannels
	cc.Match = newMatchList
	cc.Unlock()
	common.SysLog("channels Load success")
}
