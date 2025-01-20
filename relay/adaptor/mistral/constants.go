package mistral

import "github.com/songquanpeng/one-api/relay/billing/ratio"

// https://mistral.ai/technology/#pricing
var RatioMap = map[string]ratio.Ratio{
	"mistral-large-latest":      {Input: 2 * ratio.MILLI_USD, Output: 6 * ratio.MILLI_USD},
	"pixtral-large-latest":      {Input: 2 * ratio.MILLI_USD, Output: 6 * ratio.MILLI_USD},
	"mistral-small-latest":      {Input: 0.2 * ratio.MILLI_USD, Output: 0.6 * ratio.MILLI_USD},
	"codestral-latest":          {Input: 0.2 * ratio.MILLI_USD, Output: 0.6 * ratio.MILLI_USD},
	"ministral-8b-latest":       {Input: 0.1 * ratio.MILLI_USD, Output: 0.1 * ratio.MILLI_USD},
	"ministral-3b-latest":       {Input: 0.04 * ratio.MILLI_USD, Output: 0.04 * ratio.MILLI_USD},
	"mistral-embed":             {Input: 0.1 * ratio.MILLI_USD, Output: 0},
	"mistral-moderation-latest": {Input: 0.1 * ratio.MILLI_USD, Output: 0},
	"pixtral-12b":               {Input: 0.15 * ratio.MILLI_USD, Output: 0.15 * ratio.MILLI_USD},
	"mistral-nemo":              {Input: 0.15 * ratio.MILLI_USD, Output: 0.15 * ratio.MILLI_USD},
	"open-mistral-7b":           {Input: 0.25 * ratio.MILLI_USD, Output: 0.25 * ratio.MILLI_USD},
	"open-mixtral-8x7b":         {Input: 0.7 * ratio.MILLI_USD, Output: 0.7 * ratio.MILLI_USD},
	"open-mixtral-8x22b":        {Input: 2 * ratio.MILLI_USD, Output: 6 * ratio.MILLI_USD},
}
