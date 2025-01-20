package stepfun

import "github.com/songquanpeng/one-api/relay/billing/ratio"

// https://platform.stepfun.com/docs/pricing/details
var RatioMap = map[string]ratio.Ratio{
	"step-1-8k":      {Input: 5 * ratio.MILLI_RMB, Output: 20 * ratio.MILLI_RMB},
	"step-1-32k":     {Input: 15 * ratio.MILLI_RMB, Output: 70 * ratio.MILLI_RMB},
	"step-1-128k":    {Input: 40 * ratio.MILLI_RMB, Output: 200 * ratio.MILLI_RMB},
	"step-1-256k":    {Input: 95 * ratio.MILLI_RMB, Output: 300 * ratio.MILLI_RMB},
	"step-1-flash":   {Input: 1 * ratio.MILLI_RMB, Output: 4 * ratio.MILLI_RMB},
	"step-2-16k":     {Input: 38 * ratio.MILLI_RMB, Output: 120 * ratio.MILLI_RMB},
	"step-1v-8k":     {Input: 5 * ratio.MILLI_RMB, Output: 20 * ratio.MILLI_RMB},
	"step-1v-32k":    {Input: 15 * ratio.MILLI_RMB, Output: 70 * ratio.MILLI_RMB},
	"step-1.5v-mini": {Input: 8 * ratio.MILLI_RMB, Output: 35 * ratio.MILLI_RMB},
	"step-1x-medium": {Input: 0.1 * ratio.RMB, Output: 0},
}
