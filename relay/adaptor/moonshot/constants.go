package moonshot

import "github.com/songquanpeng/one-api/relay/billing/ratio"

// https://platform.moonshot.cn/docs/pricing/chat#%E4%BA%A7%E5%93%81%E5%AE%9A%E4%BB%B7
var RatioMap = map[string]ratio.Ratio{
	"moonshot-v1-8k":   {Input: 12 * ratio.MILLI_RMB, Output: 12 * ratio.MILLI_RMB},
	"moonshot-v1-32k":  {Input: 24 * ratio.MILLI_RMB, Output: 24 * ratio.MILLI_RMB},
	"moonshot-v1-128k": {Input: 60 * ratio.MILLI_RMB, Output: 60 * ratio.MILLI_RMB},
}
