package constant

import (
	"one-api/common"
)

const (
	APITypeOpenAI = iota
	APITypeClaude
	APITypePaLM
	APITypeBaidu
	APITypeZhipu
	APITypeAli
	APITypeXunfei
	APITypeAIProxyLibrary
	APITypeTencent
	APITypeGemini
	APITypeZhipu_v4
)

func ChannelType2APIType(channelType int) int {
	apiType := APITypeOpenAI
	switch channelType {
	case common.ChannelTypeAnthropic:
		apiType = APITypeClaude
	case common.ChannelTypeBaidu:
		apiType = APITypeBaidu
	case common.ChannelTypePaLM:
		apiType = APITypePaLM
	case common.ChannelTypeZhipu:
		apiType = APITypeZhipu
	case common.ChannelTypeAli:
		apiType = APITypeAli
	case common.ChannelTypeXunfei:
		apiType = APITypeXunfei
	case common.ChannelTypeAIProxyLibrary:
		apiType = APITypeAIProxyLibrary
	case common.ChannelTypeTencent:
		apiType = APITypeTencent
	case common.ChannelTypeGemini:
		apiType = APITypeGemini
	case common.ChannelTypeZhipu_v4:
		apiType = APITypeZhipu_v4
	}
	return apiType
}

//func GetAdaptor(apiType int) channel.Adaptor {
//	switch apiType {
//	case APITypeOpenAI:
//		return &openai.Adaptor{}
//	case APITypeClaude:
//		return &anthropic.Adaptor{}
//	case APITypePaLM:
//		return &google.Adaptor{}
//	case APITypeZhipu:
//		return &baidu.Adaptor{}
//	case APITypeBaidu:
//		return &baidu.Adaptor{}
//	case APITypeAli:
//		return &ali.Adaptor{}
//	case APITypeXunfei:
//		return &xunfei.Adaptor{}
//	case APITypeAIProxyLibrary:
//		return &aiproxy.Adaptor{}
//	case APITypeTencent:
//		return &tencent.Adaptor{}
//	case APITypeGemini:
//		return &google.Adaptor{}
//	}
//	return nil
//}
