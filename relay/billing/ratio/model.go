package ratio

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/songquanpeng/one-api/common/logger"
)

const (
	USD2RMB = 7
	USD     = 500 // $0.002 = 1 -> $1 = 500
	RMB     = USD / USD2RMB
)

// ModelRatio
// https://platform.openai.com/docs/models/model-endpoint-compatibility
// https://cloud.baidu.com/doc/WENXINWORKSHOP/s/Blfmc9dlf
// https://openai.com/pricing
// 1 === $0.002 / 1K tokens
// 1 === ￥0.014 / 1k tokens
var ModelRatio = map[string]float64{
	// https://openai.com/pricing
	"gpt-4":                   15,
	"gpt-4-0314":              15,
	"gpt-4-0613":              15,
	"gpt-4-32k":               30,
	"gpt-4-32k-0314":          30,
	"gpt-4-32k-0613":          30,
	"gpt-4-1106-preview":      5,     // $0.01 / 1K tokens
	"gpt-4-0125-preview":      5,     // $0.01 / 1K tokens
	"gpt-4-turbo-preview":     5,     // $0.01 / 1K tokens
	"gpt-4-turbo":             5,     // $0.01 / 1K tokens
	"gpt-4-turbo-2024-04-09":  5,     // $0.01 / 1K tokens
	"gpt-4o":                  2.5,   // $0.005 / 1K tokens
	"gpt-4o-2024-05-13":       2.5,   // $0.005 / 1K tokens
	"gpt-4o-mini":             0.075, // $0.00015 / 1K tokens
	"gpt-4o-mini-2024-07-18":  0.075, // $0.00015 / 1K tokens
	"gpt-4-vision-preview":    5,     // $0.01 / 1K tokens
	"gpt-3.5-turbo":           0.25,  // $0.0005 / 1K tokens
	"gpt-3.5-turbo-0301":      0.75,
	"gpt-3.5-turbo-0613":      0.75,
	"gpt-3.5-turbo-16k":       1.5, // $0.003 / 1K tokens
	"gpt-3.5-turbo-16k-0613":  1.5,
	"gpt-3.5-turbo-instruct":  0.75, // $0.0015 / 1K tokens
	"gpt-3.5-turbo-1106":      0.5,  // $0.001 / 1K tokens
	"gpt-3.5-turbo-0125":      0.25, // $0.0005 / 1K tokens
	"davinci-002":             1,    // $0.002 / 1K tokens
	"babbage-002":             0.2,  // $0.0004 / 1K tokens
	"text-ada-001":            0.2,
	"text-babbage-001":        0.25,
	"text-curie-001":          1,
	"text-davinci-002":        10,
	"text-davinci-003":        10,
	"text-davinci-edit-001":   10,
	"code-davinci-edit-001":   10,
	"whisper-1":               15,  // $0.006 / minute -> $0.006 / 150 words -> $0.006 / 200 tokens -> $0.03 / 1k tokens
	"tts-1":                   7.5, // $0.015 / 1K characters
	"tts-1-1106":              7.5,
	"tts-1-hd":                15, // $0.030 / 1K characters
	"tts-1-hd-1106":           15,
	"davinci":                 10,
	"curie":                   10,
	"babbage":                 10,
	"ada":                     10,
	"text-embedding-ada-002":  0.05,
	"text-embedding-3-small":  0.01,
	"text-embedding-3-large":  0.065,
	"text-search-ada-doc-001": 10,
	"text-moderation-stable":  0.1,
	"text-moderation-latest":  0.1,
	"dall-e-2":                0.02 * USD, // $0.016 - $0.020 / image
	"dall-e-3":                0.04 * USD, // $0.040 - $0.120 / image
	// https://www.anthropic.com/api#pricing
	"claude-instant-1.2":         0.8 / 1000 * USD,
	"claude-2.0":                 8.0 / 1000 * USD,
	"claude-2.1":                 8.0 / 1000 * USD,
	"claude-3-haiku-20240307":    0.25 / 1000 * USD,
	"claude-3-sonnet-20240229":   3.0 / 1000 * USD,
	"claude-3-5-sonnet-20240620": 3.0 / 1000 * USD,
	"claude-3-opus-20240229":     15.0 / 1000 * USD,
	// https://cloud.baidu.com/doc/WENXINWORKSHOP/s/hlrk4akp7
	"ERNIE-4.0-8K":       0.120 * RMB,
	"ERNIE-3.5-8K":       0.012 * RMB,
	"ERNIE-3.5-8K-0205":  0.024 * RMB,
	"ERNIE-3.5-8K-1222":  0.012 * RMB,
	"ERNIE-Bot-8K":       0.024 * RMB,
	"ERNIE-3.5-4K-0205":  0.012 * RMB,
	"ERNIE-Speed-8K":     0.004 * RMB,
	"ERNIE-Speed-128K":   0.004 * RMB,
	"ERNIE-Lite-8K-0922": 0.008 * RMB,
	"ERNIE-Lite-8K-0308": 0.003 * RMB,
	"ERNIE-Tiny-8K":      0.001 * RMB,
	"BLOOMZ-7B":          0.004 * RMB,
	"Embedding-V1":       0.002 * RMB,
	"bge-large-zh":       0.002 * RMB,
	"bge-large-en":       0.002 * RMB,
	"tao-8k":             0.002 * RMB,
	// https://ai.google.dev/pricing
	"PaLM-2":                    1,
	"gemini-pro":                1, // $0.00025 / 1k characters -> $0.001 / 1k tokens
	"gemini-pro-vision":         1, // $0.00025 / 1k characters -> $0.001 / 1k tokens
	"gemini-1.0-pro-vision-001": 1,
	"gemini-1.0-pro-001":        1,
	"gemini-1.5-pro":            1,
	// https://open.bigmodel.cn/pricing
	"glm-4":         0.1 * RMB,
	"glm-4v":        0.1 * RMB,
	"glm-3-turbo":   0.005 * RMB,
	"embedding-2":   0.0005 * RMB,
	"chatglm_turbo": 0.3572, // ￥0.005 / 1k tokens
	"chatglm_pro":   0.7143, // ￥0.01 / 1k tokens
	"chatglm_std":   0.3572, // ￥0.005 / 1k tokens
	"chatglm_lite":  0.1429, // ￥0.002 / 1k tokens
	"cogview-3":     0.25 * RMB,
	// https://help.aliyun.com/zh/dashscope/developer-reference/tongyi-thousand-questions-metering-and-billing
	"qwen-turbo":                0.5715, // ￥0.008 / 1k tokens
	"qwen-plus":                 1.4286, // ￥0.02 / 1k tokens
	"qwen-max":                  1.4286, // ￥0.02 / 1k tokens
	"qwen-max-longcontext":      1.4286, // ￥0.02 / 1k tokens
	"text-embedding-v1":         0.05,   // ￥0.0007 / 1k tokens
	"ali-stable-diffusion-xl":   8,
	"ali-stable-diffusion-v1.5": 8,
	"wanx-v1":                   8,
	"SparkDesk":                 1.2858, // ￥0.018 / 1k tokens
	"SparkDesk-v1.1":            1.2858, // ￥0.018 / 1k tokens
	"SparkDesk-v2.1":            1.2858, // ￥0.018 / 1k tokens
	"SparkDesk-v3.1":            1.2858, // ￥0.018 / 1k tokens
	"SparkDesk-v3.5":            1.2858, // ￥0.018 / 1k tokens
	"SparkDesk-v4.0":            1.2858, // ￥0.018 / 1k tokens
	"360GPT_S2_V9":              0.8572, // ¥0.012 / 1k tokens
	"embedding-bert-512-v1":     0.0715, // ¥0.001 / 1k tokens
	"embedding_s1_v1":           0.0715, // ¥0.001 / 1k tokens
	"semantic_similarity_s1_v1": 0.0715, // ¥0.001 / 1k tokens
	"hunyuan":                   7.143,  // ¥0.1 / 1k tokens  // https://cloud.tencent.com/document/product/1729/97731#e0e6be58-60c8-469f-bdeb-6c264ce3b4d0
	"ChatStd":                   0.01 * RMB,
	"ChatPro":                   0.1 * RMB,
	// https://platform.moonshot.cn/pricing
	"moonshot-v1-8k":   0.012 * RMB,
	"moonshot-v1-32k":  0.024 * RMB,
	"moonshot-v1-128k": 0.06 * RMB,
	// https://platform.baichuan-ai.com/price
	"Baichuan2-Turbo":      0.008 * RMB,
	"Baichuan2-Turbo-192k": 0.016 * RMB,
	"Baichuan2-53B":        0.02 * RMB,
	// https://api.minimax.chat/document/price
	"abab6.5-chat":  0.03 * RMB,
	"abab6.5s-chat": 0.01 * RMB,
	"abab6-chat":    0.1 * RMB,
	"abab5.5-chat":  0.015 * RMB,
	"abab5.5s-chat": 0.005 * RMB,
	// https://docs.mistral.ai/platform/pricing/
	"open-mistral-7b":       0.25 / 1000 * USD,
	"open-mixtral-8x7b":     0.7 / 1000 * USD,
	"mistral-small-latest":  2.0 / 1000 * USD,
	"mistral-medium-latest": 2.7 / 1000 * USD,
	"mistral-large-latest":  8.0 / 1000 * USD,
	"mistral-embed":         0.1 / 1000 * USD,
	// https://wow.groq.com/#:~:text=inquiries%C2%A0here.-,Model,-Current%20Speed
	"llama3-70b-8192":    0.59 / 1000 * USD,
	"mixtral-8x7b-32768": 0.27 / 1000 * USD,
	"llama3-8b-8192":     0.05 / 1000 * USD,
	"gemma-7b-it":        0.1 / 1000 * USD,
	"llama2-70b-4096":    0.64 / 1000 * USD,
	"llama2-7b-2048":     0.1 / 1000 * USD,
	// https://platform.lingyiwanwu.com/docs#-计费单元
	"yi-34b-chat-0205": 2.5 / 1000 * RMB,
	"yi-34b-chat-200k": 12.0 / 1000 * RMB,
	"yi-vl-plus":       6.0 / 1000 * RMB,
	// stepfun todo
	"step-1v-32k": 0.024 * RMB,
	"step-1-32k":  0.024 * RMB,
	"step-1-200k": 0.15 * RMB,
	// aws llama3 https://aws.amazon.com/cn/bedrock/pricing/
	"llama3-8b-8192(33)":  0.0003 / 0.002,  // $0.0003 / 1K tokens
	"llama3-70b-8192(33)": 0.00265 / 0.002, // $0.00265 / 1K tokens
	// https://cohere.com/pricing
	"command":               0.5,
	"command-nightly":       0.5,
	"command-light":         0.5,
	"command-light-nightly": 0.5,
	"command-r":             0.5 / 1000 * USD,
	"command-r-plus":        3.0 / 1000 * USD,
	// https://platform.deepseek.com/api-docs/pricing/
	"deepseek-chat":  1.0 / 1000 * RMB,
	"deepseek-coder": 1.0 / 1000 * RMB,
	// https://www.deepl.com/pro?cta=header-prices
	"deepl-zh": 25.0 / 1000 * USD,
	"deepl-en": 25.0 / 1000 * USD,
	"deepl-ja": 25.0 / 1000 * USD,
}

var CompletionRatio = map[string]float64{
	// aws llama3
	"llama3-8b-8192(33)":  0.0006 / 0.0003,
	"llama3-70b-8192(33)": 0.0035 / 0.00265,
}

var DefaultModelRatio map[string]float64
var DefaultCompletionRatio map[string]float64

func init() {
	DefaultModelRatio = make(map[string]float64)
	for k, v := range ModelRatio {
		DefaultModelRatio[k] = v
	}
	DefaultCompletionRatio = make(map[string]float64)
	for k, v := range CompletionRatio {
		DefaultCompletionRatio[k] = v
	}
}

func AddNewMissingRatio(oldRatio string) string {
	newRatio := make(map[string]float64)
	err := json.Unmarshal([]byte(oldRatio), &newRatio)
	if err != nil {
		logger.SysError("error unmarshalling old ratio: " + err.Error())
		return oldRatio
	}
	for k, v := range DefaultModelRatio {
		if _, ok := newRatio[k]; !ok {
			newRatio[k] = v
		}
	}
	jsonBytes, err := json.Marshal(newRatio)
	if err != nil {
		logger.SysError("error marshalling new ratio: " + err.Error())
		return oldRatio
	}
	return string(jsonBytes)
}

func ModelRatio2JSONString() string {
	jsonBytes, err := json.Marshal(ModelRatio)
	if err != nil {
		logger.SysError("error marshalling model ratio: " + err.Error())
	}
	return string(jsonBytes)
}

func UpdateModelRatioByJSONString(jsonStr string) error {
	ModelRatio = make(map[string]float64)
	return json.Unmarshal([]byte(jsonStr), &ModelRatio)
}

func GetModelRatio(name string, channelType int) float64 {
	if strings.HasPrefix(name, "qwen-") && strings.HasSuffix(name, "-internet") {
		name = strings.TrimSuffix(name, "-internet")
	}
	if strings.HasPrefix(name, "command-") && strings.HasSuffix(name, "-internet") {
		name = strings.TrimSuffix(name, "-internet")
	}
	model := fmt.Sprintf("%s(%d)", name, channelType)
	if ratio, ok := ModelRatio[model]; ok {
		return ratio
	}
	if ratio, ok := DefaultModelRatio[model]; ok {
		return ratio
	}
	if ratio, ok := ModelRatio[name]; ok {
		return ratio
	}
	if ratio, ok := DefaultModelRatio[name]; ok {
		return ratio
	}
	logger.SysError("model ratio not found: " + name)
	return 30
}

func CompletionRatio2JSONString() string {
	jsonBytes, err := json.Marshal(CompletionRatio)
	if err != nil {
		logger.SysError("error marshalling completion ratio: " + err.Error())
	}
	return string(jsonBytes)
}

func UpdateCompletionRatioByJSONString(jsonStr string) error {
	CompletionRatio = make(map[string]float64)
	return json.Unmarshal([]byte(jsonStr), &CompletionRatio)
}

func GetCompletionRatio(name string, channelType int) float64 {
	if strings.HasPrefix(name, "qwen-") && strings.HasSuffix(name, "-internet") {
		name = strings.TrimSuffix(name, "-internet")
	}
	model := fmt.Sprintf("%s(%d)", name, channelType)
	if ratio, ok := CompletionRatio[model]; ok {
		return ratio
	}
	if ratio, ok := DefaultCompletionRatio[model]; ok {
		return ratio
	}
	if ratio, ok := CompletionRatio[name]; ok {
		return ratio
	}
	if ratio, ok := DefaultCompletionRatio[name]; ok {
		return ratio
	}
	if strings.HasPrefix(name, "gpt-3.5") {
		if name == "gpt-3.5-turbo" || strings.HasSuffix(name, "0125") {
			// https://openai.com/blog/new-embedding-models-and-api-updates
			// Updated GPT-3.5 Turbo model and lower pricing
			return 3
		}
		if strings.HasSuffix(name, "1106") {
			return 2
		}
		return 4.0 / 3.0
	}
	if strings.HasPrefix(name, "gpt-4") {
		if strings.HasPrefix(name, "gpt-4o-mini") {
			return 4
		}
		if strings.HasPrefix(name, "gpt-4-turbo") ||
			strings.HasPrefix(name, "gpt-4o") ||
			strings.HasSuffix(name, "preview") {
			return 3
		}
		return 2
	}
	if strings.HasPrefix(name, "claude-3") {
		return 5
	}
	if strings.HasPrefix(name, "claude-") {
		return 3
	}
	if strings.HasPrefix(name, "mistral-") {
		return 3
	}
	if strings.HasPrefix(name, "gemini-") {
		return 3
	}
	if strings.HasPrefix(name, "deepseek-") {
		return 2
	}
	switch name {
	case "llama2-70b-4096":
		return 0.8 / 0.64
	case "llama3-8b-8192":
		return 2
	case "llama3-70b-8192":
		return 0.79 / 0.59
	case "command", "command-light", "command-nightly", "command-light-nightly":
		return 2
	case "command-r":
		return 3
	case "command-r-plus":
		return 5
	}
	return 1
}
