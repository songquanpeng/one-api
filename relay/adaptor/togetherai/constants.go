package togetherai

import "github.com/songquanpeng/one-api/relay/billing/ratio"

// https://www.together.ai/pricing
// https://docs.together.ai/docs/inference-models
var RatioMap = map[string]ratio.Ratio{
	"meta-llama/Meta-Llama-3.1-70B-Instruct-Turbo": {Input: 0.88 * ratio.MILLI_USD, Output: 0.88 * ratio.MILLI_USD},
	"deepseek-ai/deepseek-coder-33b-instruct":      {Input: 1.25 * ratio.MILLI_USD, Output: 1.25 * ratio.MILLI_USD},
	"mistralai/Mixtral-8x22B-Instruct-v0.1":        {Input: 1.20 * ratio.MILLI_USD, Output: 1.20 * ratio.MILLI_USD},
	"Qwen/Qwen2-72B-Instruct":                      {Input: 0.90 * ratio.MILLI_USD, Output: 0.90 * ratio.MILLI_USD},
}
