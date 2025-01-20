package anthropic

import "github.com/songquanpeng/one-api/relay/billing/ratio"

// https://www.anthropic.com/api#pricing
var RatioMap = map[string]ratio.Ratio{
	"claude-instant-1.2":         {Input: 0.8 * ratio.MILLI_USD, Output: 2.4 * ratio.MILLI_USD},
	"claude-2.0":                 {Input: 8.0 * ratio.MILLI_USD, Output: 24 * ratio.MILLI_USD},
	"claude-2.1":                 {Input: 8.0 * ratio.MILLI_USD, Output: 24 * ratio.MILLI_USD},
	"claude-3-haiku-20240307":    {Input: 0.25 * ratio.MILLI_USD, Output: 1.25 * ratio.MILLI_USD},
	"claude-3-5-haiku-20241022":  {Input: 0.8 * ratio.MILLI_USD, Output: 4 * ratio.MILLI_USD},
	"claude-3-sonnet-20240229":   {Input: 3 * ratio.MILLI_USD, Output: 15 * ratio.MILLI_USD},
	"claude-3-opus-20240229":     {Input: 15 * ratio.MILLI_USD, Output: 75 * ratio.MILLI_USD},
	"claude-3-5-sonnet-20240620": {Input: 3 * ratio.MILLI_USD, Output: 15 * ratio.MILLI_USD},
	"claude-3-5-sonnet-20241022": {Input: 3 * ratio.MILLI_USD, Output: 15 * ratio.MILLI_USD},
	"claude-3-5-sonnet-latest":   {Input: 3 * ratio.MILLI_USD, Output: 15 * ratio.MILLI_USD},
}
