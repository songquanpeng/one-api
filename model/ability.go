package model

import (
	"context"
	"sort"
	"strings"

	"gorm.io/gorm"

	"github.com/songquanpeng/one-api/common"
	"github.com/songquanpeng/one-api/common/utils"
)

type Ability struct {
	Group     string `json:"group" gorm:"type:varchar(32);primaryKey;autoIncrement:false"`
	Model     string `json:"model" gorm:"primaryKey;autoIncrement:false"`
	ChannelId int    `json:"channel_id" gorm:"primaryKey;autoIncrement:false;index"`
	Enabled   bool   `json:"enabled"`
	Priority  *int64 `json:"priority" gorm:"bigint;default:0;index"`
}

func GetRandomSatisfiedChannel(group string, model string, ignoreFirstPriority bool) (*Channel, error) {
	ability := Ability{}
	groupCol := "`group`"
	trueVal := "1"
	if common.UsingPostgreSQL {
		groupCol = `"group"`
		trueVal = "true"
	}

	var err error = nil
	var channelQuery *gorm.DB
	if ignoreFirstPriority {
		channelQuery = DB.Where(groupCol+" = ? and model = ? and enabled = "+trueVal, group, model)
	} else {
		maxPrioritySubQuery := DB.Model(&Ability{}).Select("MAX(priority)").Where(groupCol+" = ? and model = ? and enabled = "+trueVal, group, model)
		channelQuery = DB.Where(groupCol+" = ? and model = ? and enabled = "+trueVal+" and priority = (?)", group, model, maxPrioritySubQuery)
	}
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

func (channel *Channel) AddAbilities() error {
	models_ := strings.Split(channel.Models, ",")
	models_ = utils.DeDuplication(models_)
	groups_ := strings.Split(channel.Group, ",")

	// Expand models to include aliases for standard names
	expandedModels := expandModelsWithAliases(models_, channel.Type)

	abilities := make([]Ability, 0, len(expandedModels))
	for _, model := range expandedModels {
		for _, group := range groups_ {
			ability := Ability{
				Group:     group,
				Model:     model,
				ChannelId: channel.Id,
				Enabled:   channel.Status == ChannelStatusEnabled,
				Priority:  channel.Priority,
			}
			abilities = append(abilities, ability)
		}
	}
	return DB.Create(&abilities).Error
}

// expandModelsWithAliases expands model list to include standard names for channel-specific models
func expandModelsWithAliases(models []string, channelType int) []string {
	expandedModels := make([]string, 0)
	modelSet := make(map[string]bool)

	for _, model := range models {
		model = strings.TrimSpace(model)
		if model == "" {
			continue
		}

		// Add original model
		if !modelSet[model] {
			expandedModels = append(expandedModels, model)
			modelSet[model] = true
		}

		// Try to find standard name for this channel-specific model
		standardName := getStandardModelNameForChannel(model, channelType)
		if standardName != model && !modelSet[standardName] {
			expandedModels = append(expandedModels, standardName)
			modelSet[standardName] = true
		}
	}

	return expandedModels
}

// getStandardModelNameForChannel returns the standard name for a channel-specific model
func getStandardModelNameForChannel(actualName string, channelType int) string {
	// Import alias mapping (we'll create a lightweight version here to avoid circular imports)
	aliasMap := getModelAliasesForChannelType(channelType)

	for standard, actual := range aliasMap {
		if actual == actualName {
			return standard
		}
	}

	return actualName
}

// getModelAliasesForChannelType returns model aliases for specific channel type
// This is a lightweight version to avoid importing the full alias module
func getModelAliasesForChannelType(channelType int) map[string]string {
	switch channelType {
	case 24: // OpenRouter
		return map[string]string{
			"gpt-4o":            "openai/gpt-4o",
			"gpt-4o-mini":       "openai/gpt-4o-mini",
			"gpt-4":             "openai/gpt-4",
			"gpt-4-turbo":       "openai/gpt-4-turbo",
			"gpt-3.5-turbo":     "openai/gpt-3.5-turbo",
			"o1":                "openai/o1",
			"o1-mini":           "openai/o1-mini",
			"o1-preview":        "openai/o1-preview",
			"claude-3-haiku":    "anthropic/claude-3-haiku",
			"claude-3-sonnet":   "anthropic/claude-3-sonnet",
			"claude-3-opus":     "anthropic/claude-3-opus",
			"claude-3.5-sonnet": "anthropic/claude-3.5-sonnet",
			"claude-3.5-haiku":  "anthropic/claude-3.5-haiku",
		}
	default:
		return map[string]string{}
	}
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

func GetGroupModels(ctx context.Context, group string) ([]string, error) {
	groupCol := "`group`"
	trueVal := "1"
	if common.UsingPostgreSQL {
		groupCol = `"group"`
		trueVal = "true"
	}
	var models []string
	err := DB.Model(&Ability{}).Distinct("model").Where(groupCol+" = ? and enabled = "+trueVal, group).Pluck("model", &models).Error
	if err != nil {
		return nil, err
	}
	sort.Strings(models)
	return models, err
}
