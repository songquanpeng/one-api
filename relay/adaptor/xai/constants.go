package xai

import "github.com/songquanpeng/one-api/relay/billing/ratio"

var RatioMap = map[string]ratio.Ratio{
	"grok-beta": {Input: 5.0 * ratio.MILLI_USD, Output: 15.0 * ratio.MILLI_USD},
}
