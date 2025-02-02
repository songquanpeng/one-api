package ratio

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/songquanpeng/one-api/common/logger"
)

const (
	USD2RMB   = 7
	USD       = 500 // $0.002 = 1 -> $1 = 500
	MILLI_USD = 1.0 / 1000 * USD
	RMB       = USD / USD2RMB
)

// ModelRatio
// https://platform.openai.com/docs/models/model-endpoint-compatibility
// https://cloud.baidu.com/doc/WENXINWORKSHOP/s/Blfmc9dlf
// https://openai.com/pricing
// 1 === $0.002 / 1K tokens
// 1 === ￥0.014 / 1k tokens
var ModelRatio = map[string]float64{
	// https://openai.com/pricing
	"gpt-4":                  15,
	"gpt-4-0314":             15,
	"gpt-4-0613":             15,
	"gpt-4-32k":              30,
	"gpt-4-32k-0314":         30,
	"gpt-4-32k-0613":         30,
	"gpt-4-1106-preview":     5,     // $0.01 / 1K tokens
	"gpt-4-0125-preview":     5,     // $0.01 / 1K tokens
	"gpt-4-turbo-preview":    5,     // $0.01 / 1K tokens
	"gpt-4-turbo":            5,     // $0.01 / 1K tokens
	"gpt-4-turbo-2024-04-09": 5,     // $0.01 / 1K tokens
	"gpt-4o":                 2.5,   // $0.005 / 1K tokens
	"chatgpt-4o-latest":      2.5,   // $0.005 / 1K tokens
	"gpt-4o-2024-05-13":      2.5,   // $0.005 / 1K tokens
	"gpt-4o-2024-08-06":      1.25,  // $0.0025 / 1K tokens
	"gpt-4o-2024-11-20":      1.25,  // $0.0025 / 1K tokens
	"gpt-4o-mini":            0.075, // $0.00015 / 1K tokens
	"gpt-4o-mini-2024-07-18": 0.075, // $0.00015 / 1K tokens
	"gpt-4-vision-preview":   5,     // $0.01 / 1K tokens
	// Audio billing will mix text and audio tokens, the unit price is different.
	// Here records the cost of text, the cost multiplier of audio
	// relative to text is in AudioRatio
	"gpt-4o-audio-preview":                 1.25,             // $0.0025 / 1K tokens
	"gpt-4o-audio-preview-2024-12-17":      1.25,             // $0.0025 / 1K tokens
	"gpt-4o-audio-preview-2024-10-01":      1.25,             // $0.0025 / 1K tokens
	"gpt-4o-mini-audio-preview":            0.15 * MILLI_USD, // $0.15/1M tokens
	"gpt-4o-mini-audio-preview-2024-12-17": 0.15 * MILLI_USD, // $0.15/1M tokens
	"gpt-3.5-turbo":                        0.25,             // $0.0005 / 1K tokens
	"gpt-3.5-turbo-0301":                   0.75,
	"gpt-3.5-turbo-0613":                   0.75,
	"gpt-3.5-turbo-16k":                    1.5, // $0.003 / 1K tokens
	"gpt-3.5-turbo-16k-0613":               1.5,
	"gpt-3.5-turbo-instruct":               0.75, // $0.0015 / 1K tokens
	"gpt-3.5-turbo-1106":                   0.5,  // $0.001 / 1K tokens
	"gpt-3.5-turbo-0125":                   0.25, // $0.0005 / 1K tokens
	"o1":                                   7.5,  // $15.00 / 1M input tokens
	"o1-2024-12-17":                        7.5,
	"o1-preview":                           7.5, // $15.00 / 1M input tokens
	"o1-preview-2024-09-12":                7.5,
	"o1-mini":                              1.5, // $3.00 / 1M input tokens
	"o1-mini-2024-09-12":                   1.5,
	"o3-mini":                              1.1 * MILLI_USD,
	"o3-mini-2025-01-31":                   1.1 * MILLI_USD,
	"davinci-002":                          1,   // $0.002 / 1K tokens
	"babbage-002":                          0.2, // $0.0004 / 1K tokens
	"text-ada-001":                         0.2,
	"text-babbage-001":                     0.25,
	"text-curie-001":                       1,
	"text-davinci-002":                     10,
	"text-davinci-003":                     10,
	"text-davinci-edit-001":                10,
	"code-davinci-edit-001":                10,
	"whisper-1":                            15,
	"tts-1":                                7.5, // $0.015 / 1K characters
	"tts-1-1106":                           7.5,
	"tts-1-hd":                             15, // $0.030 / 1K characters
	"tts-1-hd-1106":                        15,
	"davinci":                              10,
	"curie":                                10,
	"babbage":                              10,
	"ada":                                  10,
	"text-embedding-ada-002":               0.05,
	"text-embedding-3-small":               0.01,
	"text-embedding-3-large":               0.065,
	"text-search-ada-doc-001":              10,
	"text-moderation-stable":               0.1,
	"text-moderation-latest":               0.1,
	"dall-e-2":                             0.02 * USD, // $0.016 - $0.020 / image
	"dall-e-3":                             0.04 * USD, // $0.040 - $0.120 / image
	// https://www.anthropic.com/api#pricing
	"claude-instant-1.2":         0.8 / 1000 * USD,
	"claude-2.0":                 8.0 / 1000 * USD,
	"claude-2.1":                 8.0 / 1000 * USD,
	"claude-3-haiku-20240307":    0.25 / 1000 * USD,
	"claude-3-5-haiku-20241022":  1.0 / 1000 * USD,
	"claude-3-sonnet-20240229":   3.0 / 1000 * USD,
	"claude-3-5-sonnet-20240620": 3.0 / 1000 * USD,
	"claude-3-5-sonnet-20241022": 3.0 / 1000 * USD,
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
	"gemini-pro":                          1, // $0.00025 / 1k characters -> $0.001 / 1k tokens
	"gemini-1.0-pro":                      1,
	"gemini-1.5-pro":                      1,
	"gemini-1.5-pro-001":                  1,
	"gemini-1.5-flash":                    1,
	"gemini-1.5-flash-001":                1,
	"gemini-2.0-flash-exp":                1,
	"gemini-2.0-flash-thinking-exp":       1,
	"gemini-2.0-flash-thinking-exp-01-21": 1,
	"aqa":                                 1,
	// https://open.bigmodel.cn/pricing
	"glm-zero-preview": 0.01 * RMB,
	"glm-4-plus":       0.05 * RMB,
	"glm-4-0520":       0.1 * RMB,
	"glm-4-airx":       0.01 * RMB,
	"glm-4-air":        0.0005 * RMB,
	"glm-4-long":       0.001 * RMB,
	"glm-4-flashx":     0.0001 * RMB,
	"glm-4-flash":      0,
	"glm-4":            0.1 * RMB,   // deprecated model, available until 2025/06
	"glm-3-turbo":      0.001 * RMB, // deprecated model, available until 2025/06
	"glm-4v-plus":      0.004 * RMB,
	"glm-4v":           0.05 * RMB,
	"glm-4v-flash":     0,
	"cogview-3-plus":   0.06 * RMB,
	"cogview-3":        0.1 * RMB,
	"cogview-3-flash":  0,
	"cogviewx":         0.5 * RMB,
	"cogviewx-flash":   0,
	"charglm-4":        0.001 * RMB,
	"emohaa":           0.015 * RMB,
	"codegeex-4":       0.0001 * RMB,
	"embedding-2":      0.0005 * RMB,
	"embedding-3":      0.0005 * RMB,
	// https://help.aliyun.com/zh/dashscope/developer-reference/tongyi-thousand-questions-metering-and-billing
	"qwen-turbo":                  1.4286, // ￥0.02 / 1k tokens
	"qwen-turbo-latest":           1.4286,
	"qwen-plus":                   1.4286,
	"qwen-plus-latest":            1.4286,
	"qwen-max":                    1.4286,
	"qwen-max-latest":             1.4286,
	"qwen-max-longcontext":        1.4286,
	"qwen-vl-max":                 1.4286,
	"qwen-vl-max-latest":          1.4286,
	"qwen-vl-plus":                1.4286,
	"qwen-vl-plus-latest":         1.4286,
	"qwen-vl-ocr":                 1.4286,
	"qwen-vl-ocr-latest":          1.4286,
	"qwen-audio-turbo":            1.4286,
	"qwen-math-plus":              1.4286,
	"qwen-math-plus-latest":       1.4286,
	"qwen-math-turbo":             1.4286,
	"qwen-math-turbo-latest":      1.4286,
	"qwen-coder-plus":             1.4286,
	"qwen-coder-plus-latest":      1.4286,
	"qwen-coder-turbo":            1.4286,
	"qwen-coder-turbo-latest":     1.4286,
	"qwq-32b-preview":             1.4286,
	"qwen2.5-72b-instruct":        1.4286,
	"qwen2.5-32b-instruct":        1.4286,
	"qwen2.5-14b-instruct":        1.4286,
	"qwen2.5-7b-instruct":         1.4286,
	"qwen2.5-3b-instruct":         1.4286,
	"qwen2.5-1.5b-instruct":       1.4286,
	"qwen2.5-0.5b-instruct":       1.4286,
	"qwen2-72b-instruct":          1.4286,
	"qwen2-57b-a14b-instruct":     1.4286,
	"qwen2-7b-instruct":           1.4286,
	"qwen2-1.5b-instruct":         1.4286,
	"qwen2-0.5b-instruct":         1.4286,
	"qwen1.5-110b-chat":           1.4286,
	"qwen1.5-72b-chat":            1.4286,
	"qwen1.5-32b-chat":            1.4286,
	"qwen1.5-14b-chat":            1.4286,
	"qwen1.5-7b-chat":             1.4286,
	"qwen1.5-1.8b-chat":           1.4286,
	"qwen1.5-0.5b-chat":           1.4286,
	"qwen-72b-chat":               1.4286,
	"qwen-14b-chat":               1.4286,
	"qwen-7b-chat":                1.4286,
	"qwen-1.8b-chat":              1.4286,
	"qwen-1.8b-longcontext-chat":  1.4286,
	"qwen2-vl-7b-instruct":        1.4286,
	"qwen2-vl-2b-instruct":        1.4286,
	"qwen-vl-v1":                  1.4286,
	"qwen-vl-chat-v1":             1.4286,
	"qwen2-audio-instruct":        1.4286,
	"qwen-audio-chat":             1.4286,
	"qwen2.5-math-72b-instruct":   1.4286,
	"qwen2.5-math-7b-instruct":    1.4286,
	"qwen2.5-math-1.5b-instruct":  1.4286,
	"qwen2-math-72b-instruct":     1.4286,
	"qwen2-math-7b-instruct":      1.4286,
	"qwen2-math-1.5b-instruct":    1.4286,
	"qwen2.5-coder-32b-instruct":  1.4286,
	"qwen2.5-coder-14b-instruct":  1.4286,
	"qwen2.5-coder-7b-instruct":   1.4286,
	"qwen2.5-coder-3b-instruct":   1.4286,
	"qwen2.5-coder-1.5b-instruct": 1.4286,
	"qwen2.5-coder-0.5b-instruct": 1.4286,
	"text-embedding-v1":           0.05, // ￥0.0007 / 1k tokens
	"text-embedding-v3":           0.05,
	"text-embedding-v2":           0.05,
	"text-embedding-async-v2":     0.05,
	"text-embedding-async-v1":     0.05,
	"ali-stable-diffusion-xl":     8.00,
	"ali-stable-diffusion-v1.5":   8.00,
	"wanx-v1":                     8.00,
	"SparkDesk":                   1.2858, // ￥0.018 / 1k tokens
	"SparkDesk-v1.1":              1.2858, // ￥0.018 / 1k tokens
	"SparkDesk-v2.1":              1.2858, // ￥0.018 / 1k tokens
	"SparkDesk-v3.1":              1.2858, // ￥0.018 / 1k tokens
	"SparkDesk-v3.1-128K":         1.2858, // ￥0.018 / 1k tokens
	"SparkDesk-v3.5":              1.2858, // ￥0.018 / 1k tokens
	"SparkDesk-v3.5-32K":          1.2858, // ￥0.018 / 1k tokens
	"SparkDesk-v4.0":              1.2858, // ￥0.018 / 1k tokens
	"360GPT_S2_V9":                0.8572, // ¥0.012 / 1k tokens
	"embedding-bert-512-v1":       0.0715, // ¥0.001 / 1k tokens
	"embedding_s1_v1":             0.0715, // ¥0.001 / 1k tokens
	"semantic_similarity_s1_v1":   0.0715, // ¥0.001 / 1k tokens
	// https://cloud.tencent.com/document/product/1729/97731#e0e6be58-60c8-469f-bdeb-6c264ce3b4d0
	"hunyuan-turbo":             0.015 * RMB,
	"hunyuan-large":             0.004 * RMB,
	"hunyuan-large-longcontext": 0.006 * RMB,
	"hunyuan-standard":          0.0008 * RMB,
	"hunyuan-standard-256K":     0.0005 * RMB,
	"hunyuan-translation-lite":  0.005 * RMB,
	"hunyuan-role":              0.004 * RMB,
	"hunyuan-functioncall":      0.004 * RMB,
	"hunyuan-code":              0.004 * RMB,
	"hunyuan-turbo-vision":      0.08 * RMB,
	"hunyuan-vision":            0.018 * RMB,
	"hunyuan-embedding":         0.0007 * RMB,
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
	"gemma-7b-it":                           0.07 / 1000000 * USD,
	"gemma2-9b-it":                          0.20 / 1000000 * USD,
	"llama-3.1-70b-versatile":               0.59 / 1000000 * USD,
	"llama-3.1-8b-instant":                  0.05 / 1000000 * USD,
	"llama-3.2-11b-text-preview":            0.05 / 1000000 * USD,
	"llama-3.2-11b-vision-preview":          0.05 / 1000000 * USD,
	"llama-3.2-1b-preview":                  0.05 / 1000000 * USD,
	"llama-3.2-3b-preview":                  0.05 / 1000000 * USD,
	"llama-3.2-90b-text-preview":            0.59 / 1000000 * USD,
	"llama-guard-3-8b":                      0.05 / 1000000 * USD,
	"llama3-70b-8192":                       0.59 / 1000000 * USD,
	"llama3-8b-8192":                        0.05 / 1000000 * USD,
	"llama3-groq-70b-8192-tool-use-preview": 0.89 / 1000000 * USD,
	"llama3-groq-8b-8192-tool-use-preview":  0.19 / 1000000 * USD,
	"mixtral-8x7b-32768":                    0.24 / 1000000 * USD,

	// https://platform.lingyiwanwu.com/docs#-计费单元
	"yi-34b-chat-0205": 2.5 / 1000 * RMB,
	"yi-34b-chat-200k": 12.0 / 1000 * RMB,
	"yi-vl-plus":       6.0 / 1000 * RMB,
	// https://platform.stepfun.com/docs/pricing/details
	"step-1-8k":    0.005 / 1000 * RMB,
	"step-1-32k":   0.015 / 1000 * RMB,
	"step-1-128k":  0.040 / 1000 * RMB,
	"step-1-256k":  0.095 / 1000 * RMB,
	"step-1-flash": 0.001 / 1000 * RMB,
	"step-2-16k":   0.038 / 1000 * RMB,
	"step-1v-8k":   0.005 / 1000 * RMB,
	"step-1v-32k":  0.015 / 1000 * RMB,
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
	"deepseek-chat":     0.14 * MILLI_USD,
	"deepseek-reasoner": 0.55 * MILLI_USD,
	// https://www.deepl.com/pro?cta=header-prices
	"deepl-zh": 25.0 / 1000 * USD,
	"deepl-en": 25.0 / 1000 * USD,
	"deepl-ja": 25.0 / 1000 * USD,
	// https://console.x.ai/
	"grok-beta": 5.0 / 1000 * USD,
	// replicate charges based on the number of generated images
	// https://replicate.com/pricing
	"black-forest-labs/flux-1.1-pro":                0.04 * USD,
	"black-forest-labs/flux-1.1-pro-ultra":          0.06 * USD,
	"black-forest-labs/flux-canny-dev":              0.025 * USD,
	"black-forest-labs/flux-canny-pro":              0.05 * USD,
	"black-forest-labs/flux-depth-dev":              0.025 * USD,
	"black-forest-labs/flux-depth-pro":              0.05 * USD,
	"black-forest-labs/flux-dev":                    0.025 * USD,
	"black-forest-labs/flux-dev-lora":               0.032 * USD,
	"black-forest-labs/flux-fill-dev":               0.04 * USD,
	"black-forest-labs/flux-fill-pro":               0.05 * USD,
	"black-forest-labs/flux-pro":                    0.055 * USD,
	"black-forest-labs/flux-redux-dev":              0.025 * USD,
	"black-forest-labs/flux-redux-schnell":          0.003 * USD,
	"black-forest-labs/flux-schnell":                0.003 * USD,
	"black-forest-labs/flux-schnell-lora":           0.02 * USD,
	"ideogram-ai/ideogram-v2":                       0.08 * USD,
	"ideogram-ai/ideogram-v2-turbo":                 0.05 * USD,
	"recraft-ai/recraft-v3":                         0.04 * USD,
	"recraft-ai/recraft-v3-svg":                     0.08 * USD,
	"stability-ai/stable-diffusion-3":               0.035 * USD,
	"stability-ai/stable-diffusion-3.5-large":       0.065 * USD,
	"stability-ai/stable-diffusion-3.5-large-turbo": 0.04 * USD,
	"stability-ai/stable-diffusion-3.5-medium":      0.035 * USD,
	// replicate chat models
	"ibm-granite/granite-20b-code-instruct-8k":  0.100 * USD,
	"ibm-granite/granite-3.0-2b-instruct":       0.030 * USD,
	"ibm-granite/granite-3.0-8b-instruct":       0.050 * USD,
	"ibm-granite/granite-8b-code-instruct-128k": 0.050 * USD,
	"meta/llama-2-13b":                          0.100 * USD,
	"meta/llama-2-13b-chat":                     0.100 * USD,
	"meta/llama-2-70b":                          0.650 * USD,
	"meta/llama-2-70b-chat":                     0.650 * USD,
	"meta/llama-2-7b":                           0.050 * USD,
	"meta/llama-2-7b-chat":                      0.050 * USD,
	"meta/meta-llama-3.1-405b-instruct":         9.500 * USD,
	"meta/meta-llama-3-70b":                     0.650 * USD,
	"meta/meta-llama-3-70b-instruct":            0.650 * USD,
	"meta/meta-llama-3-8b":                      0.050 * USD,
	"meta/meta-llama-3-8b-instruct":             0.050 * USD,
	"mistralai/mistral-7b-instruct-v0.2":        0.050 * USD,
	"mistralai/mistral-7b-v0.1":                 0.050 * USD,
	"mistralai/mixtral-8x7b-instruct-v0.1":      0.300 * USD,
}

// AudioRatio represents the price ratio between audio tokens and text tokens
var AudioRatio = map[string]float64{
	"gpt-4o-audio-preview":                 16,
	"gpt-4o-audio-preview-2024-12-17":      16,
	"gpt-4o-audio-preview-2024-10-01":      40,
	"gpt-4o-mini-audio-preview":            10 / 0.15,
	"gpt-4o-mini-audio-preview-2024-12-17": 10 / 0.15,
}

// GetAudioPromptRatio returns the audio prompt ratio for the given model.
func GetAudioPromptRatio(actualModelName string) float64 {
	var v float64
	if ratio, ok := AudioRatio[actualModelName]; ok {
		v = ratio
	} else {
		v = 16
	}

	return v
}

// AudioCompletionRatio is the completion ratio for audio models.
var AudioCompletionRatio = map[string]float64{
	"whisper-1":                            0,
	"gpt-4o-audio-preview":                 2,
	"gpt-4o-audio-preview-2024-12-17":      2,
	"gpt-4o-audio-preview-2024-10-01":      2,
	"gpt-4o-mini-audio-preview":            2,
	"gpt-4o-mini-audio-preview-2024-12-17": 2,
}

// GetAudioCompletionRatio returns the completion ratio for audio models.
func GetAudioCompletionRatio(actualModelName string) float64 {
	var v float64
	if ratio, ok := AudioCompletionRatio[actualModelName]; ok {
		v = ratio
	} else {
		v = 2
	}

	return v
}

// AudioTokensPerSecond is the number of audio tokens per second for each model.
var AudioPromptTokensPerSecond = map[string]float64{
	// Whisper API price is $0.0001/sec. One-api's historical ratio is 15,
	// corresponding to $0.03/kilo_tokens.
	// After conversion, tokens per second should be 0.0001/0.03*1000 = 3.3333.
	"whisper-1": 0.0001 / 0.03 * 1000,
	// gpt-4o-audio series processes 10 tokens per second
	"gpt-4o-audio-preview":                 10,
	"gpt-4o-audio-preview-2024-12-17":      10,
	"gpt-4o-audio-preview-2024-10-01":      10,
	"gpt-4o-mini-audio-preview":            10,
	"gpt-4o-mini-audio-preview-2024-12-17": 10,
}

// GetAudioPromptTokensPerSecond returns the number of audio tokens per second
// for the given model.
func GetAudioPromptTokensPerSecond(actualModelName string) float64 {
	var v float64
	if tokensPerSecond, ok := AudioPromptTokensPerSecond[actualModelName]; ok {
		v = tokensPerSecond
	} else {
		v = 10
	}

	return v
}

var CompletionRatio = map[string]float64{
	// aws llama3
	"llama3-8b-8192(33)":  0.0006 / 0.0003,
	"llama3-70b-8192(33)": 0.0035 / 0.00265,
	// whisper
	"whisper-1": 0, // only count input tokens
	// deepseek
	"deepseek-chat":     0.28 / 0.14,
	"deepseek-reasoner": 2.19 / 0.55,
}

var (
	DefaultModelRatio      map[string]float64
	DefaultCompletionRatio map[string]float64
)

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
		if strings.HasPrefix(name, "gpt-4o") {
			if name == "gpt-4o-2024-05-13" {
				return 3
			}
			return 4
		}
		if strings.HasPrefix(name, "gpt-4-turbo") ||
			strings.HasSuffix(name, "preview") {
			return 3
		}
		return 2
	}
	// including o1/o1-preview/o1-mini
	if strings.HasPrefix(name, "o1") ||
		strings.HasPrefix(name, "o3") {
		return 4
	}
	if name == "chatgpt-4o-latest" {
		return 3
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
	case "grok-beta":
		return 3
	// Replicate Models
	// https://replicate.com/pricing
	case "ibm-granite/granite-20b-code-instruct-8k":
		return 5
	case "ibm-granite/granite-3.0-2b-instruct":
		return 8.333333333333334
	case "ibm-granite/granite-3.0-8b-instruct",
		"ibm-granite/granite-8b-code-instruct-128k":
		return 5
	case "meta/llama-2-13b",
		"meta/llama-2-13b-chat",
		"meta/llama-2-7b",
		"meta/llama-2-7b-chat",
		"meta/meta-llama-3-8b",
		"meta/meta-llama-3-8b-instruct":
		return 5
	case "meta/llama-2-70b",
		"meta/llama-2-70b-chat",
		"meta/meta-llama-3-70b",
		"meta/meta-llama-3-70b-instruct":
		return 2.750 / 0.650 // ≈4.230769
	case "meta/meta-llama-3.1-405b-instruct":
		return 1
	case "mistralai/mistral-7b-instruct-v0.2",
		"mistralai/mistral-7b-v0.1":
		return 5
	case "mistralai/mixtral-8x7b-instruct-v0.1":
		return 1.000 / 0.300 // ≈3.333333
	}

	return 1
}
