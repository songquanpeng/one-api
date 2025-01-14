package lingyiwanwu

import "github.com/songquanpeng/one-api/relay/billing/ratio"

// https://platform.lingyiwanwu.com/docs#%E6%A8%A1%E5%9E%8B%E4%B8%8E%E8%AE%A1%E8%B4%B9
var RatioMap = map[string]ratio.Ratio{
	"yi-lightning": {Input: 0.99 * ratio.MILLI_RMB, Output: 0.99 * ratio.MILLI_RMB},
	"yi-vision-v2": {Input: 6 * ratio.MILLI_RMB, Output: 6 * ratio.MILLI_RMB},
}
