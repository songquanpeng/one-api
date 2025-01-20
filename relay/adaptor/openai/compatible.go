package openai

import (
	"github.com/songquanpeng/one-api/relay/adaptor/ai360"
	"github.com/songquanpeng/one-api/relay/adaptor/baichuan"
	"github.com/songquanpeng/one-api/relay/adaptor/deepseek"
	"github.com/songquanpeng/one-api/relay/adaptor/doubao"
	"github.com/songquanpeng/one-api/relay/adaptor/groq"
	"github.com/songquanpeng/one-api/relay/adaptor/lingyiwanwu"
	"github.com/songquanpeng/one-api/relay/adaptor/minimax"
	"github.com/songquanpeng/one-api/relay/adaptor/mistral"
	"github.com/songquanpeng/one-api/relay/adaptor/moonshot"
	"github.com/songquanpeng/one-api/relay/adaptor/novita"
	"github.com/songquanpeng/one-api/relay/adaptor/siliconflow"
	"github.com/songquanpeng/one-api/relay/adaptor/stepfun"
	"github.com/songquanpeng/one-api/relay/adaptor/togetherai"
	"github.com/songquanpeng/one-api/relay/adaptor/xai"
	"github.com/songquanpeng/one-api/relay/billing/ratio"
	"github.com/songquanpeng/one-api/relay/channeltype"
)

var CompatibleChannels = []int{
	channeltype.Azure,
	channeltype.AI360,
	channeltype.Moonshot,
	channeltype.Baichuan,
	channeltype.Minimax,
	channeltype.Doubao,
	channeltype.Mistral,
	channeltype.Groq,
	channeltype.LingYiWanWu,
	channeltype.StepFun,
	channeltype.DeepSeek,
	channeltype.TogetherAI,
	channeltype.Novita,
	channeltype.SiliconFlow,
	channeltype.XAI,
}

func GetCompatibleChannelMeta(channelType int) (string, map[string]ratio.Ratio) {
	switch channelType {
	case channeltype.Azure:
		return "azure", RatioMap
	case channeltype.AI360:
		return "360", ai360.RatioMap
	case channeltype.Moonshot:
		return "moonshot", moonshot.RatioMap
	case channeltype.Baichuan:
		return "baichuan", baichuan.RatioMap
	case channeltype.Minimax:
		return "minimax", minimax.RatioMap
	case channeltype.Mistral:
		return "mistralai", mistral.RatioMap
	case channeltype.Groq:
		return "groq", groq.RatioMap
	case channeltype.LingYiWanWu:
		return "lingyiwanwu", lingyiwanwu.RatioMap
	case channeltype.StepFun:
		return "stepfun", stepfun.RatioMap
	case channeltype.DeepSeek:
		return "deepseek", deepseek.RatioMap
	case channeltype.TogetherAI:
		return "together.ai", togetherai.RatioMap
	case channeltype.Doubao:
		return "doubao", doubao.RatioMap
	case channeltype.Novita:
		return "novita", novita.RatioMap
	case channeltype.SiliconFlow:
		return "siliconflow", siliconflow.RatioMap
	case channeltype.XAI:
		return "xai", xai.RatioMap
	default:
		return "openai", RatioMap
	}
}
