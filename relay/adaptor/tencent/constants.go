package tencent

import "github.com/songquanpeng/one-api/relay/billing/ratio"

// https://cloud.tencent.com/document/product/1729/97731
var RatioMap = map[string]ratio.Ratio{
	"hunyuan-turbo":             {Input: 0.015 * ratio.RMB, Output: 0.05 * ratio.RMB},
	"hunyuan-large":             {Input: 0.004 * ratio.RMB, Output: 0.012 * ratio.RMB},
	"hunyuan-large-longcontext": {Input: 0.006 * ratio.RMB, Output: 0.018 * ratio.RMB},
	"hunyuan-standard":          {Input: 0.0008 * ratio.RMB, Output: 0.002 * ratio.RMB},
	"hunyuan-standard-256K":     {Input: 0.0005 * ratio.RMB, Output: 0.002 * ratio.RMB},
	"hunyuan-translation-lite":  {Input: 0.005 * ratio.RMB, Output: 0.015 * ratio.RMB},
	"hunyuan-role":              {Input: 0.004 * ratio.RMB, Output: 0.008 * ratio.RMB},
	"hunyuan-functioncall":      {Input: 0.004 * ratio.RMB, Output: 0.008 * ratio.RMB},
	"hunyuan-code":              {Input: 0.004 * ratio.RMB, Output: 0.008 * ratio.RMB},
	"hunyuan-turbo-vision":      {Input: 0.08 * ratio.RMB, Output: 0.08 * ratio.RMB},
	"hunyuan-vision":            {Input: 0.018 * ratio.RMB, Output: 0.018 * ratio.RMB},
	"hunyuan-embedding":         {Input: 0.0007 * ratio.RMB, Output: 0.0007 * ratio.RMB},
}
