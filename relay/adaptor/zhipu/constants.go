package zhipu

import "github.com/songquanpeng/one-api/relay/billing/ratio"

var RatioMap = map[string]ratio.Ratio{
	"glm-zero-preview": {Input: 0.01 * ratio.RMB, Output: 0.01 * ratio.RMB},
	"glm-4-plus":       {Input: 0.05 * ratio.RMB, Output: 0.05 * ratio.RMB},
	"glm-4-0520":       {Input: 0.1 * ratio.RMB, Output: 0.1 * ratio.RMB},
	"glm-4-airx":       {Input: 0.01 * ratio.RMB, Output: 0.01 * ratio.RMB},
	"glm-4-air":        {Input: 0.0005 * ratio.RMB, Output: 0.0005 * ratio.RMB},
	"glm-4-long":       {Input: 0.001 * ratio.RMB, Output: 0.001 * ratio.RMB},
	"glm-4-flashx":     {Input: 0.0001 * ratio.RMB, Output: 0.0001 * ratio.RMB},
	"glm-4v-plus":      {Input: 0.004 * ratio.RMB, Output: 0.004 * ratio.RMB},
	"glm-4v":           {Input: 0.05 * ratio.RMB, Output: 0},
	"cogview-3-plus":   {Input: 0.06 * ratio.RMB, Output: 0},
	"cogview-3":        {Input: 0.1 * ratio.RMB, Output: 0},
	"cogvideox":        {Input: 0.5 * ratio.RMB, Output: 0},
	"embedding-3":      {Input: 0.0005 * ratio.RMB, Output: 0},
	"embedding-2":      {Input: 0.0005 * ratio.RMB, Output: 0},
	"glm-4-flash":      {Input: 0, Output: 0}, // 免费
	"glm-4v-flash":     {Input: 0, Output: 0}, // 免费
	"cogview-3-flash":  {Input: 0, Output: 0}, // 免费
	"cogvideox-flash":  {Input: 0, Output: 0}, // 免费
}
