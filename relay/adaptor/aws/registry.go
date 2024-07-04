package aws

import (
	claude "github.com/songquanpeng/one-api/relay/adaptor/aws/claude"
	llama3 "github.com/songquanpeng/one-api/relay/adaptor/aws/llama3"
	"github.com/songquanpeng/one-api/relay/adaptor/aws/utils"
)

type AwsModelType int

const (
	AwsClaude AwsModelType = iota + 1
	AwsLlama3
)

var (
	adaptors = map[string]AwsModelType{}
)

func init() {
	for model := range claude.AwsModelIDMap {
		adaptors[model] = AwsClaude
	}
	for model := range llama3.AwsModelIDMap {
		adaptors[model] = AwsLlama3
	}
}

func GetAdaptor(model string) utils.AwsAdapter {
	adaptorType := adaptors[model]
	switch adaptorType {
	case AwsClaude:
		return &claude.Adaptor{}
	case AwsLlama3:
		return &llama3.Adaptor{}
	default:
		return nil
	}
}
