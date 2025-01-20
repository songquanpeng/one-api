package xunfei

import "github.com/songquanpeng/one-api/relay/billing/ratio"

var RatioMap = map[string]ratio.Ratio{
	"SparkDesk":           {Input: 1.2858, Output: 1.2858}, // ￥0.018 / 1k tokens
	"SparkDesk-v1.1":      {Input: 1.2858, Output: 1.2858}, // ￥0.018 / 1k tokens
	"SparkDesk-v2.1":      {Input: 1.2858, Output: 1.2858}, // ￥0.018 / 1k tokens
	"SparkDesk-v3.1":      {Input: 1.2858, Output: 1.2858}, // ￥0.018 / 1k tokens
	"SparkDesk-v3.1-128K": {Input: 1.2858, Output: 1.2858}, // ￥0.018 / 1k tokens
	"SparkDesk-v3.5":      {Input: 1.2858, Output: 1.2858}, // ￥0.018 / 1k tokens
	"SparkDesk-v3.5-32K":  {Input: 1.2858, Output: 1.2858}, // ￥0.018 / 1k tokens
	"SparkDesk-v4.0":      {Input: 1.2858, Output: 1.2858}, // ￥0.018 / 1k tokens
}
