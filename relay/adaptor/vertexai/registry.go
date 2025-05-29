package vertexai

import (
	claude "github.com/songquanpeng/one-api/relay/adaptor/vertexai/claude"
	embedding "github.com/songquanpeng/one-api/relay/adaptor/vertexai/embedding"
	gemini "github.com/songquanpeng/one-api/relay/adaptor/vertexai/gemini"
	"github.com/songquanpeng/one-api/relay/adaptor/vertexai/model"
)

type VertexAIModelType int

const (
	VertexAIClaude VertexAIModelType = iota + 1
	VertexAIGemini
	VertexAIEmbedding
)

var modelMapping = map[string]VertexAIModelType{}
var modelList = []string{}

func init() {
	modelList = append(modelList, claude.ModelList...)
	for _, model := range claude.ModelList {
		modelMapping[model] = VertexAIClaude
	}

	modelList = append(modelList, gemini.ModelList...)
	for _, model := range gemini.ModelList {
		modelMapping[model] = VertexAIGemini
	}

	modelList = append(modelList, embedding.ModelList...)
	for _, model := range embedding.ModelList {
		modelMapping[model] = VertexAIEmbedding
	}
}

func GetAdaptor(model string) model.InnerAIAdapter {
	adaptorType := modelMapping[model]
	switch adaptorType {
	case VertexAIClaude:
		return &claude.Adaptor{}
	case VertexAIGemini:
		return &gemini.Adaptor{}
	case VertexAIEmbedding:
		return &embedding.Adaptor{}
	default:
		adaptorType = PredictModelType(model)
		switch adaptorType {
		case VertexAIGemini:
			return &gemini.Adaptor{}
		case VertexAIEmbedding:
			return &embedding.Adaptor{}
		}
		return nil
	}
}
