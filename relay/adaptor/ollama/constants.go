package ollama

import "github.com/songquanpeng/one-api/relay/billing/ratio"

var RatioMap = map[string]ratio.Ratio{
	"codellama:7b-instruct": {Input: 0, Output: 0},
	"llama2:7b":             {Input: 0, Output: 0},
	"llama2:latest":         {Input: 0, Output: 0},
	"llama3:latest":         {Input: 0, Output: 0},
	"phi3:latest":           {Input: 0, Output: 0},
	"qwen:0.5b-chat":        {Input: 0, Output: 0},
	"qwen:7b":               {Input: 0, Output: 0},
}
