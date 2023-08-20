package controller

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/pkoukk/tiktoken-go"
	"gorm.io/gorm/utils"
	"one-api/common"
	"strings"
)

var stopFinishReason = "stop"

var tokenEncoderMap = map[string]*tiktoken.Tiktoken{}

func getTokenEncoder(model string) *tiktoken.Tiktoken {
	if tokenEncoder, ok := tokenEncoderMap[model]; ok {
		return tokenEncoder
	}
	tokenEncoder, err := tiktoken.EncodingForModel(model)
	if err != nil {
		common.SysError(fmt.Sprintf("failed to get token encoder for model %s: %s, using encoder for gpt-3.5-turbo", model, err.Error()))
		tokenEncoder, err = tiktoken.EncodingForModel("gpt-3.5-turbo")
		if err != nil {
			common.FatalLog(fmt.Sprintf("failed to get token encoder for model gpt-3.5-turbo: %s", err.Error()))
		}
	}
	tokenEncoderMap[model] = tokenEncoder
	return tokenEncoder
}

func getTokenNum(tokenEncoder *tiktoken.Tiktoken, text string) int {
	if common.ApproximateTokenEnabled {
		return int(float64(len(text)) * 0.38)
	}
	return len(tokenEncoder.Encode(text, nil, nil))
}

func countTokenFunctionCall(functionCall any, model string) int {
	tokenEncoder := getTokenEncoder(model)
	jsonBytes, err := json.Marshal(functionCall)
	if err != nil {
		return 0
	}
	return getTokenNum(tokenEncoder, string(jsonBytes))
}

func countTokenFunctions(functions []Function, model string) int {
	// https://community.openai.com/t/how-to-know-of-tokens-beforehand-when-i-make-function-calling-chat-history-request-witn-nodejs/289060/6
	if len(functions) == 0 {
		return 0
	}
	tokenEncoder := getTokenEncoder(model)

	paramSignature := func(name string, pSpec Property, pRequired []string) string {
		var requiredString string
		if utils.Contains(pRequired, name) == false {
			requiredString = "?"
		}
		var enumString string
		if len(pSpec.Enum) > 0 {
			enumValues := make([]string, len(pSpec.Enum))
			for i, v := range pSpec.Enum {
				enumValues[i] = fmt.Sprintf("\"%s\"", v)
			}
			enumString = strings.Join(enumValues, " | ")
		} else {
			enumString = pSpec.Type
		}
		signature := fmt.Sprintf("%s%s: %s, ", name, requiredString, enumString)
		if pSpec.Description != "" {
			signature = fmt.Sprintf("// %s\n%s", pSpec.Description, signature)
		}
		return signature
	}

	functionSignature := func(fSpec Function) string {
		var params []string
		for name, p := range fSpec.Parameters.Properties {
			params = append(params, paramSignature(name, p, fSpec.Parameters.Required))
		}
		var descriptionString string
		if fSpec.Description != "" {
			descriptionString = fmt.Sprintf("// %s\n", fSpec.Description)
		}

		var paramString string
		if len(params) > 0 {
			paramString = fmt.Sprintf("_: {\n%s\n}", strings.Join(params, "\n"))
		}

		return fmt.Sprintf("%stype %s = (%s) => any;", descriptionString, fSpec.Name, paramString)
	}

	var functionSignatures []string
	for _, f := range functions {
		functionSignatures = append(functionSignatures, functionSignature(f))
	}
	functionString := fmt.Sprintf("# Tools\n\n## functions\n\nnamespace functions {\n\n%s\n\n} // namespace functions", strings.Join(functionSignatures, "\n\n"))

	return getTokenNum(tokenEncoder, functionString)
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
	if model == "gpt-3.5-turbo-0301" {
		tokensPerMessage = 4
		tokensPerName = -1 // If there's a name, the role is omitted
	} else {
		tokensPerMessage = 3
		tokensPerName = 1
	}
	tokenNum := 0
	for _, message := range messages {
		tokenNum += tokensPerMessage
		tokenNum += getTokenNum(tokenEncoder, message.Content)
		tokenNum += getTokenNum(tokenEncoder, message.Role)
		if message.Name != nil {
			tokenNum += tokensPerName
			tokenNum += getTokenNum(tokenEncoder, *message.Name)
		}
	}
	tokenNum += 3 // Every reply is primed with <|start|>assistant<|message|>
	return tokenNum
}

func countTokenInput(input any, model string) int {
	switch input.(type) {
	case string:
		return countTokenText(input.(string), model)
	case []string:
		text := ""
		for _, s := range input.([]string) {
			text += s
		}
		return countTokenText(text, model)
	}
	return 0
}

func countTokenText(text string, model string) int {
	tokenEncoder := getTokenEncoder(model)
	return getTokenNum(tokenEncoder, text)
}

func errorWrapper(err error, code string, statusCode int) *OpenAIErrorWithStatusCode {
	openAIError := OpenAIError{
		Message: err.Error(),
		Type:    "one_api_error",
		Code:    code,
	}
	return &OpenAIErrorWithStatusCode{
		OpenAIError: openAIError,
		StatusCode:  statusCode,
	}
}

func shouldDisableChannel(err *OpenAIError) bool {
	if !common.AutomaticDisableChannelEnabled {
		return false
	}
	if err == nil {
		return false
	}
	if err.Type == "insufficient_quota" || err.Code == "invalid_api_key" || err.Code == "account_deactivated" {
		return true
	}
	return false
}

func setEventStreamHeaders(c *gin.Context) {
	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Header().Set("Transfer-Encoding", "chunked")
	c.Writer.Header().Set("X-Accel-Buffering", "no")
}
