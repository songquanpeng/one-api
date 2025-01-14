package groq

import "github.com/songquanpeng/one-api/relay/billing/ratio"

// https://groq.com/pricing/
// https://console.groq.com/docs/models
var RatioMap = map[string]ratio.Ratio{
	"distil-whisper-large-v3-en": {Input: 0.02 / 3600 * 20 * ratio.USD, Output: 0},
	"gemma2-9b-it":               {Input: 0.20 * ratio.MILLI_USD, Output: 0.20 * ratio.MILLI_USD},
	"llama-3.3-70b-versatile":    {Input: 0.59 * ratio.MILLI_USD, Output: 0.79 * ratio.MILLI_USD},
	"llama-3.1-8b-instant":       {Input: 0.05 * ratio.MILLI_USD, Output: 0.08 * ratio.MILLI_USD},
	"llama-guard-3-8b":           {Input: 0.20 * ratio.MILLI_USD, Output: 0.20 * ratio.MILLI_USD},
	"llama3-70b-8192":            {Input: 0.59 * ratio.MILLI_USD, Output: 0.79 * ratio.MILLI_USD},
	"llama3-8b-8192":             {Input: 0.05 * ratio.MILLI_USD, Output: 0.08 * ratio.MILLI_USD},
	"mixtral-8x7b-32768":         {Input: 0.24 * ratio.MILLI_USD, Output: 0.24 * ratio.MILLI_USD},
	"whisper-large-v3":           {Input: 0.111 / 3600 * 20 * ratio.USD, Output: 0},
	"whisper-large-v3-turbo":     {Input: 0.04 / 3600 * 20 * ratio.USD, Output: 0},
}
