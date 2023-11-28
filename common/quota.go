package common

// type Quota struct {
// 	ModelName  string
// 	ModelRatio float64
// 	GroupRatio float64
// 	Ratio      float64
// 	UserQuota  int
// }

// func CreateQuota(modelName string, userQuota int, group string) *Quota {
// 	modelRatio := GetModelRatio(modelName)
// 	groupRatio := GetGroupRatio(group)

// 	return &Quota{
// 		ModelName:  modelName,
// 		ModelRatio: modelRatio,
// 		GroupRatio: groupRatio,
// 		Ratio:      modelRatio * groupRatio,
// 		UserQuota:  userQuota,
// 	}
// }

// func (q *Quota) getTokenNum(tokenEncoder *tiktoken.Tiktoken, text string) int {
// 	if ApproximateTokenEnabled {
// 		return int(float64(len(text)) * 0.38)
// 	}
// 	return len(tokenEncoder.Encode(text, nil, nil))
// }

// func (q *Quota) CountTokenMessages(messages []Message, model string) int {
// 	tokenEncoder := q.getTokenEncoder(model)
// 	// Reference:
// 	// https://github.com/openai/openai-cookbook/blob/main/examples/How_to_count_tokens_with_tiktoken.ipynb
// 	// https://github.com/pkoukk/tiktoken-go/issues/6
// 	//
// 	// Every message follows <|start|>{role/name}\n{content}<|end|>\n
// 	var tokensPerMessage int
// 	var tokensPerName int
// 	if model == "gpt-3.5-turbo-0301" {
// 		tokensPerMessage = 4
// 		tokensPerName = -1 // If there's a name, the role is omitted
// 	} else {
// 		tokensPerMessage = 3
// 		tokensPerName = 1
// 	}
// 	tokenNum := 0
// 	for _, message := range messages {
// 		tokenNum += tokensPerMessage
// 		tokenNum += q.getTokenNum(tokenEncoder, message.StringContent())
// 		tokenNum += q.getTokenNum(tokenEncoder, message.Role)
// 		if message.Name != nil {
// 			tokenNum += tokensPerName
// 			tokenNum += q.getTokenNum(tokenEncoder, *message.Name)
// 		}
// 	}
// 	tokenNum += 3 // Every reply is primed with <|start|>assistant<|message|>
// 	return tokenNum
// }
