package deepseek

import "github.com/songquanpeng/one-api/relay/billing/ratio"

var RatioMap = map[string]ratio.Ratio{
	"deepseek-chat":  {Input: 1 * ratio.MILLI_RMB, Output: 2 * ratio.MILLI_RMB},
	"deepseek-coder": {Input: 1 * ratio.MILLI_RMB, Output: 2 * ratio.MILLI_RMB},
}
