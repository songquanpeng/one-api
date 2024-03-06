package model

import (
	"one-api/common"
	"strings"
)

type Ability struct {
	Group     string `json:"group" gorm:"type:varchar(32);primaryKey;autoIncrement:false"`
	Model     string `json:"model" gorm:"primaryKey;autoIncrement:false"`
	ChannelId int    `json:"channel_id" gorm:"primaryKey;autoIncrement:false;index"`
	Enabled   bool   `json:"enabled"`
	Priority  *int64 `json:"priority" gorm:"bigint;default:0;index"`
	Weight    *uint  `json:"weight" gorm:"default:1"`
}

func GetRandomSatisfiedChannel(group string, model string) (*Channel, error) {
	ability := Ability{}
	groupCol := "`group`"
	trueVal := "1"
	if common.UsingPostgreSQL {
		groupCol = `"group"`
		trueVal = "true"
	}

	var err error = nil
	maxPrioritySubQuery := DB.Model(&Ability{}).Select("MAX(priority)").Where(groupCol+" = ? and model = ? and enabled = "+trueVal, group, model)
	channelQuery := DB.Where(groupCol+" = ? and model = ? and enabled = "+trueVal+" and priority = (?)", group, model, maxPrioritySubQuery)
	if common.UsingSQLite || common.UsingPostgreSQL {
		err = channelQuery.Order("RANDOM()").First(&ability).Error
	} else {
		err = channelQuery.Order("RAND()").First(&ability).Error
	}
	if err != nil {
		return nil, err
	}
	channel := Channel{}
	channel.Id = ability.ChannelId
	err = DB.First(&channel, "id = ?", ability.ChannelId).Error
	return &channel, err
}

func GetGroupModels(group string) ([]string, error) {
	var models []string
	groupCol := "`group`"
	trueVal := "1"
	if common.UsingPostgreSQL {
		groupCol = `"group"`
		trueVal = "true"
	}

	err := DB.Model(&Ability{}).Where(groupCol+" = ? and enabled = ? ", group, trueVal).Distinct("model").Pluck("model", &models).Error
	if err != nil {
		return nil, err
	}
	return models, nil
}

func (channel *Channel) AddAbilities() error {
	models_ := strings.Split(channel.Models, ",")
	groups_ := strings.Split(channel.Group, ",")
	abilities := make([]Ability, 0, len(models_))
	for _, model := range models_ {
		for _, group := range groups_ {
			ability := Ability{
				Group:     group,
				Model:     model,
				ChannelId: channel.Id,
				Enabled:   channel.Status == common.ChannelStatusEnabled,
				Priority:  channel.Priority,
				Weight:    channel.Weight,
			}
			abilities = append(abilities, ability)
		}
	}
	return DB.Create(&abilities).Error
}

func (channel *Channel) DeleteAbilities() error {
	return DB.Where("channel_id = ?", channel.Id).Delete(&Ability{}).Error
}

// UpdateAbilities updates abilities of this channel.
// Make sure the channel is completed before calling this function.
func (channel *Channel) UpdateAbilities() error {
	// A quick and dirty way to update abilities
	// First delete all abilities of this channel
	err := channel.DeleteAbilities()
	if err != nil {
		return err
	}
	// Then add new abilities
	err = channel.AddAbilities()
	if err != nil {
		return err
	}
	return nil
}

func UpdateAbilityStatus(channelId int, status bool) error {
	return DB.Model(&Ability{}).Where("channel_id = ?", channelId).Select("enabled").Update("enabled", status).Error
}

func GetEnabledAbility() ([]*Ability, error) {
	trueVal := "1"
	if common.UsingPostgreSQL {
		trueVal = "true"
	}

	var abilities []*Ability
	err := DB.Where("enabled = ?", trueVal).Order("priority desc, weight desc").Find(&abilities).Error
	return abilities, err
}

type AbilityChannelGroup struct {
	Group      string `json:"group"`
	Model      string `json:"model"`
	Priority   int    `json:"priority"`
	ChannelIds string `json:"channel_ids"`
}

func GetAbilityChannelGroup() ([]*AbilityChannelGroup, error) {
	var abilities []*AbilityChannelGroup

	var channelSql string
	if common.UsingPostgreSQL {
		channelSql = `string_agg("channel_id"::text, ',')`
	} else if common.UsingSQLite {
		channelSql = `group_concat("channel_id", ',')`
	} else {
		channelSql = "GROUP_CONCAT(`channel_id` SEPARATOR ',')"
	}

	trueVal := "1"
	if common.UsingPostgreSQL {
		trueVal = "true"
	}

	err := DB.Raw(`
	SELECT `+quotePostgresField("group")+`, model, priority, `+channelSql+` as channel_ids
	FROM abilities
	WHERE enabled = ?
	GROUP BY `+quotePostgresField("group")+`, model, priority
	ORDER BY priority DESC
	`, trueVal).Scan(&abilities).Error

	return abilities, err
}
