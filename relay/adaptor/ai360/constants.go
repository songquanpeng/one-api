package ai360

import "github.com/songquanpeng/one-api/relay/billing/ratio"

var RatioMap = map[string]ratio.Ratio{
	"360GPT_S2_V9":              {Input: 0.012 * ratio.RMB, Output: 0.012 * ratio.RMB},
	"embedding-bert-512-v1":     {Input: 0.0001 * ratio.RMB, Output: 0},
	"embedding_s1_v1":           {Input: 0.0001 * ratio.RMB, Output: 0},
	"semantic_similarity_s1_v1": {Input: 0.0001 * ratio.RMB, Output: 0},
}
