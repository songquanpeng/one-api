package minimax

import "github.com/songquanpeng/one-api/relay/billing/ratio"

// https://platform.minimaxi.com/document/Price
// https://platform.minimaxi.com/document/ChatCompletion%20v2

var RatioMap = map[string]ratio.Ratio{
	"abab7-chat-preview": {Input: 0.01 * ratio.RMB, Output: 0.01 * ratio.RMB},
	"abab6.5s-chat":      {Input: 0.001 * ratio.RMB, Output: 0.001 * ratio.RMB},
	"abab6.5g-chat":      {Input: 0.005 * ratio.RMB, Output: 0.005 * ratio.RMB},
	"abab6.5t-chat":      {Input: 0.005 * ratio.RMB, Output: 0.005 * ratio.RMB},
	"abab5.5s-chat":      {Input: 0.005 * ratio.RMB, Output: 0.005 * ratio.RMB},
	"abab5.5-chat":       {Input: 0.015 * ratio.RMB, Output: 0.015 * ratio.RMB},
}
