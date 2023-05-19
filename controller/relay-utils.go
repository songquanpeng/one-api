package controller

import (
	"fmt"
	"github.com/pkoukk/tiktoken-go"
	"one-api/common"
	"strings"
)

var tokenEncoderMap = map[string]*tiktoken.Tiktoken{}

func getTokenEncoder(model string) *tiktoken.Tiktoken {
	if tokenEncoder, ok := tokenEncoderMap[model]; ok {
		return tokenEncoder
	}
	tokenEncoder, err := tiktoken.EncodingForModel(model)
	if err != nil {
		common.FatalLog(fmt.Sprintf("failed to get token encoder for model %s: %s", model, err.Error()))
	}
	tokenEncoderMap[model] = tokenEncoder
	return tokenEncoder
}

func countTokenMessages(messages []Message, model string) int {
	tokenEncoder := getTokenEncoder(model)
	// Reference:
	// https://github.com/openai/openai-cookbook/blob/main/examples/How_to_count_tokens_with_tiktoken.ipynb
	// https://github.com/pkoukk/tiktoken-go/issues/6
	//
	// Every message follows <|start|>{role/name}\n{content}<|end|>\n
	var tokensPerMessage int
	var tokensPerName int
	if strings.HasPrefix(model, "gpt-3.5") {
		tokensPerMessage = 4
		tokensPerName = -1 // If there's a name, the role is omitted
	} else if strings.HasPrefix(model, "gpt-4") {
		tokensPerMessage = 3
		tokensPerName = 1
	} else {
		tokensPerMessage = 3
		tokensPerName = 1
	}
	tokenNum := 0
	for _, message := range messages {
		tokenNum += tokensPerMessage
		tokenNum += len(tokenEncoder.Encode(message.Content, nil, nil))
		tokenNum += len(tokenEncoder.Encode(message.Role, nil, nil))
		if message.Name != nil {
			tokenNum += tokensPerName
			tokenNum += len(tokenEncoder.Encode(*message.Name, nil, nil))
		}
	}
	tokenNum += 3 // Every reply is primed with <|start|>assistant<|message|>
	return tokenNum
}

func countTokenText(text string, model string) int {
	tokenEncoder := getTokenEncoder(model)
	token := tokenEncoder.Encode(text, nil, nil)
	return len(token)
}
