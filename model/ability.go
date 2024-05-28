package model

import (
	"one-api/common"
	"one-api/common/config"
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
				Enabled:   channel.Status == config.ChannelStatusEnabled,
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
