package novita

import "github.com/songquanpeng/one-api/relay/billing/ratio"

// https://novita.ai/llm-api

var RatioMap = map[string]ratio.Ratio{
	"meta-llama/llama-3.3-70b-instruct":           {Input: 0.39 * ratio.MILLI_USD, Output: 0.39 * ratio.MILLI_USD},
	"meta-llama/llama-3.1-8b-instruct":            {Input: 0.05 * ratio.MILLI_USD, Output: 0.05 * ratio.MILLI_USD},
	"meta-llama/llama-3.1-8b-instruct-max":        {Input: 0.05 * ratio.MILLI_USD, Output: 0.05 * ratio.MILLI_USD},
	"meta-llama/llama-3.1-70b-instruct":           {Input: 0.34 * ratio.MILLI_USD, Output: 0.39 * ratio.MILLI_USD},
	"meta-llama/llama-3-8b-instruct":              {Input: 0.04 * ratio.MILLI_USD, Output: 0.04 * ratio.MILLI_USD},
	"meta-llama/llama-3-70b-instruct":             {Input: 0.51 * ratio.MILLI_USD, Output: 0.74 * ratio.MILLI_USD},
	"gryphe/mythomax-l2-13b":                      {Input: 0.09 * ratio.MILLI_USD, Output: 0.09 * ratio.MILLI_USD},
	"google/gemma-2-9b-it":                        {Input: 0.08 * ratio.MILLI_USD, Output: 0.08 * ratio.MILLI_USD},
	"mistralai/mistral-nemo":                      {Input: 0.17 * ratio.MILLI_USD, Output: 0.17 * ratio.MILLI_USD},
	"microsoft/wizardlm-2-8x22b":                  {Input: 0.62 * ratio.MILLI_USD, Output: 0.62 * ratio.MILLI_USD},
	"mistralai/mistral-7b-instruct":               {Input: 0.059 * ratio.MILLI_USD, Output: 0.059 * ratio.MILLI_USD},
	"openchat/openchat-7b":                        {Input: 0.06 * ratio.MILLI_USD, Output: 0.06 * ratio.MILLI_USD},
	"nousresearch/hermes-2-pro-llama-3-8b":        {Input: 0.14 * ratio.MILLI_USD, Output: 0.14 * ratio.MILLI_USD},
	"sao10k/l3-70b-euryale-v2.1":                  {Input: 1.48 * ratio.MILLI_USD, Output: 1.48 * ratio.MILLI_USD},
	"cognitivecomputations/dolphin-mixtral-8x22b": {Input: 0.9 * ratio.MILLI_USD, Output: 0.9 * ratio.MILLI_USD},
	"jondurbin/airoboros-l2-70b":                  {Input: 0.5 * ratio.MILLI_USD, Output: 0.5 * ratio.MILLI_USD},
	"nousresearch/nous-hermes-llama2-13b":         {Input: 0.17 * ratio.MILLI_USD, Output: 0.17 * ratio.MILLI_USD},
	"teknium/openhermes-2.5-mistral-7b":           {Input: 0.17 * ratio.MILLI_USD, Output: 0.17 * ratio.MILLI_USD},
	"sophosympatheia/midnight-rose-70b":           {Input: 0.8 * ratio.MILLI_USD, Output: 0.8 * ratio.MILLI_USD},
	"Sao10K/L3-8B-Stheno-v3.2":                    {Input: 0.05 * ratio.MILLI_USD, Output: 0.05 * ratio.MILLI_USD},
	"sao10k/l3-8b-lunaris":                        {Input: 0.05 * ratio.MILLI_USD, Output: 0.05 * ratio.MILLI_USD},
	"qwen/qwen-2-vl-72b-instruct":                 {Input: 0.45 * ratio.MILLI_USD, Output: 0.45 * ratio.MILLI_USD},
	"meta-llama/llama-3.2-1b-instruct":            {Input: 0.02 * ratio.MILLI_USD, Output: 0.02 * ratio.MILLI_USD},
	"meta-llama/llama-3.2-11b-vision-instruct":    {Input: 0.06 * ratio.MILLI_USD, Output: 0.06 * ratio.MILLI_USD},
	"meta-llama/llama-3.2-3b-instruct":            {Input: 0.03 * ratio.MILLI_USD, Output: 0.05 * ratio.MILLI_USD},
	"meta-llama/llama-3.1-8b-instruct-bf16":       {Input: 0.06 * ratio.MILLI_USD, Output: 0.06 * ratio.MILLI_USD},
	"qwen/qwen-2.5-72b-instruct":                  {Input: 0.38 * ratio.MILLI_USD, Output: 0.4 * ratio.MILLI_USD},
	"sao10k/l31-70b-euryale-v2.2":                 {Input: 1.48 * ratio.MILLI_USD, Output: 1.48 * ratio.MILLI_USD},
	"qwen/qwen-2-7b-instruct":                     {Input: 0.054 * ratio.MILLI_USD, Output: 0.054 * ratio.MILLI_USD},
	"qwen/qwen-2-72b-instruct":                    {Input: 0.34 * ratio.MILLI_USD, Output: 0.39 * ratio.MILLI_USD},
}
