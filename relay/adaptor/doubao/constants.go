package doubao

import "github.com/songquanpeng/one-api/relay/billing/ratio"

// https://www.volcengine.com/product/doubao
var RatioMap = map[string]ratio.Ratio{
	"Doubao-vision-pro-32k":  {Input: 0.0030 * ratio.RMB, Output: 0.0090 * ratio.RMB},
	"Doubao-vision-lite-32k": {Input: 0.0015 * ratio.RMB, Output: 0.0045 * ratio.RMB},
	"Doubao-pro-256k":        {Input: 0.0050 * ratio.RMB, Output: 0.0090 * ratio.RMB},
	"Doubao-pro-128k":        {Input: 0.0050 * ratio.RMB, Output: 0.0090 * ratio.RMB},
	"Doubao-pro-32k":         {Input: 0.0008 * ratio.RMB, Output: 0.0020 * ratio.RMB},
	"Doubao-lite-128k":       {Input: 0.0008 * ratio.RMB, Output: 0.0010 * ratio.RMB},
	"Doubao-lite-32k":        {Input: 0.0003 * ratio.RMB, Output: 0.0006 * ratio.RMB},
	"Doubao-embedding":       {Input: 0.0005 * ratio.RMB, Output: 0},
}
