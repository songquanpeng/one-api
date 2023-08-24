package common

import (
	"encoding/json"
	"strings"
)

// ModelRatio
// https://platform.openai.com/docs/models/model-endpoint-compatibility
// https://cloud.baidu.com/doc/WENXINWORKSHOP/s/Blfmc9dlf
// https://openai.com/pricing
// TODO: when a new api is enabled, check the pricing here
// 1 === $0.002 / 1K tokens
// 1 === ￥0.014 / 1k tokens
var ModelRatio = map[string]float64{
	"gpt-4":                   15,
	"gpt-4-0314":              15,
	"gpt-4-0613":              15,
	"gpt-4-32k":               30,
	"gpt-4-32k-0314":          30,
	"gpt-4-32k-0613":          30,
	"gpt-3.5-turbo":           0.75, // $0.0015 / 1K tokens
	"gpt-3.5-turbo-0301":      0.75,
	"gpt-3.5-turbo-0613":      0.75,
	"gpt-3.5-turbo-16k":       1.5, // $0.003 / 1K tokens
	"gpt-3.5-turbo-16k-0613":  1.5,
	"text-ada-001":            0.2,
	"text-babbage-001":        0.25,
	"text-curie-001":          1,
	"text-davinci-002":        10,
	"text-davinci-003":        10,
	"text-davinci-edit-001":   10,
	"code-davinci-edit-001":   10,
	"whisper-1":               10,
	"davinci":                 10,
	"curie":                   10,
	"babbage":                 10,
	"ada":                     10,
	"text-embedding-ada-002":  0.05,
	"text-search-ada-doc-001": 10,
	"text-moderation-stable":  0.1,
	"text-moderation-latest":  0.1,
	"dall-e":                  8,
	"claude-instant-1":        0.815,  // $1.63 / 1M tokens
	"claude-2":                5.51,   // $11.02 / 1M tokens
	"ERNIE-Bot":               0.8572, // ￥0.012 / 1k tokens
	"ERNIE-Bot-turbo":         0.5715, // ￥0.008 / 1k tokens
	"Embedding-V1":            0.1429, // ￥0.002 / 1k tokens
	"PaLM-2":                  1,
	"chatglm_pro":             0.7143, // ￥0.01 / 1k tokens
	"chatglm_std":             0.3572, // ￥0.005 / 1k tokens
	"chatglm_lite":            0.1429, // ￥0.002 / 1k tokens
	"qwen-v1":                 0.8572, // TBD: https://help.aliyun.com/document_detail/2399482.html?spm=a2c4g.2399482.0.0.1ad347feilAgag
	"qwen-plus-v1":            0.5715, // Same as above
	"SparkDesk":               0.8572, // TBD
}

func ModelRatio2JSONString() string {
	jsonBytes, err := json.Marshal(ModelRatio)
	if err != nil {
		SysError("error marshalling model ratio: " + err.Error())
	}
	return string(jsonBytes)
}

func UpdateModelRatioByJSONString(jsonStr string) error {
	ModelRatio = make(map[string]float64)
	return json.Unmarshal([]byte(jsonStr), &ModelRatio)
}

func GetModelRatio(name string) float64 {
	ratio, ok := ModelRatio[name]
	if !ok {
		SysError("model ratio not found: " + name)
		return 30
	}
	return ratio
}

func GetCompletionRatio(name string) float64 {
	if strings.HasPrefix(name, "gpt-3.5") {
		return 1.333333
	}
	if strings.HasPrefix(name, "gpt-4") {
		return 2
	}
	if strings.HasPrefix(name, "claude-instant-1") {
		return 3.38
	}
	if strings.HasPrefix(name, "claude-2") {
		return 2.965517
	}
	return 1
}
