package gemini

import "github.com/songquanpeng/one-api/relay/billing/ratio"

// https://ai.google.dev/models/gemini

var gemini15FlashRatio = ratio.Ratio{
	Input:         0.075 * ratio.MILLI_USD,
	Output:        0.30 * ratio.MILLI_USD,
	LongThreshold: 128000,
	LongInput:     0.15 * ratio.MILLI_USD,
	LongOutput:    0.60 * ratio.MILLI_USD,
}

var gemini15ProRatio = ratio.Ratio{
	Input:         1.25 * ratio.MILLI_USD,
	Output:        5.00 * ratio.MILLI_USD,
	LongThreshold: 128000,
	LongInput:     2.50 * ratio.MILLI_USD,
	LongOutput:    10.00 * ratio.MILLI_USD,
}

var gemini10ProRatio = ratio.Ratio{
	Input:  0.50 * ratio.MILLI_USD,
	Output: 1.50 * ratio.MILLI_USD,
}

var gemini15Flash8bRatio = ratio.Ratio{
	Input:         0.0375 * ratio.MILLI_USD,
	Output:        0.15 * ratio.MILLI_USD,
	LongThreshold: 128000,
	LongInput:     0.075 * ratio.MILLI_USD,
	LongOutput:    0.30 * ratio.MILLI_USD,
}

// https://ai.google.dev/pricing
// https://ai.google.dev/gemini-api/docs/models/gemini
// https://cloud.google.com/vertex-ai/generative-ai/pricing?hl=zh-cn#google_models
var RatioMap = map[string]ratio.Ratio{
	"gemini-2.0-flash-exp":          {Input: 0.1, Output: 0.1}, // currently free of charge
	"gemini-2.0-flash-thinking-exp": {Input: 0.1, Output: 0.1}, // currently free of charge
	"gemini-1.5-flash":              gemini15FlashRatio,
	"gemini-1.5-flash-001":          gemini15FlashRatio,
	"gemini-1.5-flash-002":          gemini15FlashRatio,
	"gemini-1.5-pro":                gemini15ProRatio,
	"gemini-1.5-pro-001":            gemini15ProRatio,
	"gemini-1.5-pro-002":            gemini15ProRatio,
	"gemini-1.0-pro":                gemini10ProRatio,
	"gemini-1.0-pro-001":            gemini10ProRatio,
	"gemini-1.5-flash-8b":           gemini15Flash8bRatio,
	"gemini-1.5-flash-8b-001":       gemini15Flash8bRatio,
	"text-embedding-004":            {Input: 0.1, Output: 0.1}, // free of charge
}
