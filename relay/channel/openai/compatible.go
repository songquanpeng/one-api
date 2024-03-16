package openai

import (
	"github.com/songquanpeng/one-api/common"
	"github.com/songquanpeng/one-api/relay/channel/ai360"
	"github.com/songquanpeng/one-api/relay/channel/baichuan"
	"github.com/songquanpeng/one-api/relay/channel/groq"
	"github.com/songquanpeng/one-api/relay/channel/lingyiwanwu"
	"github.com/songquanpeng/one-api/relay/channel/minimax"
	"github.com/songquanpeng/one-api/relay/channel/mistral"
	"github.com/songquanpeng/one-api/relay/channel/moonshot"
)

var CompatibleChannels = []int{
	common.ChannelTypeAzure,
	common.ChannelType360,
	common.ChannelTypeMoonshot,
	common.ChannelTypeBaichuan,
	common.ChannelTypeMinimax,
	common.ChannelTypeMistral,
	common.ChannelTypeGroq,
	common.ChannelTypeLingYiWanWu,
}

func GetCompatibleChannelMeta(channelType int) (string, []string) {
	switch channelType {
	case common.ChannelTypeAzure:
		return "azure", ModelList
	case common.ChannelType360:
		return "360", ai360.ModelList
	case common.ChannelTypeMoonshot:
		return "moonshot", moonshot.ModelList
	case common.ChannelTypeBaichuan:
		return "baichuan", baichuan.ModelList
	case common.ChannelTypeMinimax:
		return "minimax", minimax.ModelList
	case common.ChannelTypeMistral:
		return "mistralai", mistral.ModelList
	case common.ChannelTypeGroq:
		return "groq", groq.ModelList
	case common.ChannelTypeLingYiWanWu:
		return "lingyiwanwu", lingyiwanwu.ModelList
	default:
		return "openai", ModelList
	}
}
