package deepl

import "github.com/songquanpeng/one-api/relay/billing/ratio"

// https://developers.deepl.com/docs/api-reference/glossaries
var RatioMap = map[string]ratio.Ratio{
	"deepl-zh": {Input: 25.0 * ratio.MILLI_USD, Output: 0},
	"deepl-en": {Input: 25.0 * ratio.MILLI_USD, Output: 0},
	"deepl-ja": {Input: 25.0 * ratio.MILLI_USD, Output: 0},
}
