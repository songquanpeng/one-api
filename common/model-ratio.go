package common

import "encoding/json"

// https://platform.openai.com/docs/models/model-endpoint-compatibility
// https://openai.com/pricing
// TODO: when a new api is enabled, check the pricing here
var ModelRatio = map[string]float64{
	"gpt-4":                   15,
	"gpt-4-0314":              15,
	"gpt-4-32k":               30,
	"gpt-4-32k-0314":          30,
	"gpt-3.5-turbo":           1, // $0.002 / 1K tokens
	"gpt-3.5-turbo-0301":      1,
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
	"text-embedding-ada-002":  0.2,
	"text-search-ada-doc-001": 10,
	"text-moderation-stable":  0.1,
	"text-moderation-latest":  0.1,
}

func ModelRatio2JSONString() string {
	jsonBytes, err := json.Marshal(ModelRatio)
	if err != nil {
		SysError("Error marshalling model ratio: " + err.Error())
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
		SysError("Model ratio not found: " + name)
		return 1
	}
	return ratio
}
