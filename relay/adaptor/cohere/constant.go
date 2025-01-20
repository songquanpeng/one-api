package cohere

import "github.com/songquanpeng/one-api/relay/billing/ratio"

// https://cohere.com/pricing
var RatioMap = map[string]ratio.Ratio{
	"command":                        {Input: 1 * ratio.MILLI_USD, Output: 2 * ratio.MILLI_USD},
	"command-internet":               {Input: 1 * ratio.MILLI_USD, Output: 2 * ratio.MILLI_USD},
	"command-nightly":                {Input: 1 * ratio.MILLI_USD, Output: 2 * ratio.MILLI_USD},
	"command-nightly-internet":       {Input: 1 * ratio.MILLI_USD, Output: 2 * ratio.MILLI_USD},
	"command-light":                  {Input: 0.3 * ratio.MILLI_USD, Output: 0.6 * ratio.MILLI_USD},
	"command-light-internet":         {Input: 0.3 * ratio.MILLI_USD, Output: 0.6 * ratio.MILLI_USD},
	"command-light-nightly":          {Input: 0.3 * ratio.MILLI_USD, Output: 0.6 * ratio.MILLI_USD},
	"command-light-nightly-internet": {Input: 0.3 * ratio.MILLI_USD, Output: 0.6 * ratio.MILLI_USD},
	"command-r":                      {Input: 0.15 * ratio.MILLI_USD, Output: 0.6 * ratio.MILLI_USD},
	"command-r-internet":             {Input: 0.15 * ratio.MILLI_USD, Output: 0.6 * ratio.MILLI_USD},
	"command-r-plus":                 {Input: 2.5 * ratio.MILLI_USD, Output: 10 * ratio.MILLI_USD},
	"command-r-plus-internet":        {Input: 2.5 * ratio.MILLI_USD, Output: 10 * ratio.MILLI_USD},
}
