package aiproxy

import "github.com/songquanpeng/one-api/relay/channel/openai"

var ModelList = []string{""}

func init() {
	ModelList = openai.ModelList
}
