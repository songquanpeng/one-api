package baichuan

import "github.com/songquanpeng/one-api/relay/billing/ratio"

// https://platform.baichuan-ai.com/price
var RatioMap = map[string]ratio.Ratio{
	"Baichuan2-Turbo":         {Input: 0.008 * ratio.RMB, Output: 0.008 * ratio.RMB},
	"Baichuan2-Turbo-192k":    {Input: 0.016 * ratio.RMB, Output: 0.016 * ratio.RMB},
	"Baichuan-Text-Embedding": {Input: 0.001 * ratio.RMB, Output: 0.001 * ratio.RMB},
}
