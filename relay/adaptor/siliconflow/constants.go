package siliconflow

import "github.com/songquanpeng/one-api/relay/billing/ratio"

// https://siliconflow.cn/zh-cn/models
// https://siliconflow.cn/zh-cn/pricing
var RatioMap = map[string]ratio.Ratio{
	"Qwen/Qwen2.5-72B-Instruct":               {Input: 41.3 * ratio.MILLI_RMB, Output: 41.3 * ratio.MILLI_RMB},
	"Qwen/Qwen2.5-7B-Instruct":                {Input: 3.5 * ratio.MILLI_RMB, Output: 3.5 * ratio.MILLI_RMB},
	"deepseek-ai/deepseek-llm-67b-chat":       {Input: 0, Output: 0},
	"Qwen/Qwen1.5-14B-Chat":                   {Input: 0, Output: 0},
	"Qwen/Qwen1.5-7B-Chat":                    {Input: 0, Output: 0},
	"Qwen/Qwen1.5-110B-Chat":                  {Input: 0, Output: 0},
	"Qwen/Qwen1.5-32B-Chat":                   {Input: 0, Output: 0},
	"01-ai/Yi-1.5-6B-Chat":                    {Input: 0, Output: 0},
	"01-ai/Yi-1.5-9B-Chat-16K":                {Input: 0, Output: 0},
	"01-ai/Yi-1.5-34B-Chat-16K":               {Input: 0, Output: 0},
	"THUDM/chatglm3-6b":                       {Input: 0, Output: 0},
	"deepseek-ai/DeepSeek-V2-Chat":            {Input: 0, Output: 0},
	"THUDM/glm-4-9b-chat":                     {Input: 0, Output: 0},
	"Qwen/Qwen2-72B-Instruct":                 {Input: 0, Output: 0},
	"Qwen/Qwen2-7B-Instruct":                  {Input: 0, Output: 0},
	"Qwen/Qwen2-57B-A14B-Instruct":            {Input: 0, Output: 0},
	"deepseek-ai/DeepSeek-Coder-V2-Instruct":  {Input: 0, Output: 0},
	"Qwen/Qwen2-1.5B-Instruct":                {Input: 0, Output: 0},
	"internlm/internlm2_5-7b-chat":            {Input: 0, Output: 0},
	"BAAI/bge-large-en-v1.5":                  {Input: 0, Output: 0},
	"BAAI/bge-large-zh-v1.5":                  {Input: 0, Output: 0},
	"Pro/Qwen/Qwen2-7B-Instruct":              {Input: 0, Output: 0},
	"Pro/Qwen/Qwen2-1.5B-Instruct":            {Input: 0, Output: 0},
	"Pro/Qwen/Qwen1.5-7B-Chat":                {Input: 0, Output: 0},
	"Pro/THUDM/glm-4-9b-chat":                 {Input: 0, Output: 0},
	"Pro/THUDM/chatglm3-6b":                   {Input: 0, Output: 0},
	"Pro/01-ai/Yi-1.5-9B-Chat-16K":            {Input: 0, Output: 0},
	"Pro/01-ai/Yi-1.5-6B-Chat":                {Input: 0, Output: 0},
	"Pro/google/gemma-2-9b-it":                {Input: 0, Output: 0},
	"Pro/internlm/internlm2_5-7b-chat":        {Input: 0, Output: 0},
	"Pro/meta-llama/Meta-Llama-3-8B-Instruct": {Input: 0, Output: 0},
	"Pro/mistralai/Mistral-7B-Instruct-v0.2":  {Input: 0, Output: 0},
}
