package test

import (
	"one-api/types"
	"testing"

	"github.com/stretchr/testify/assert"
)

func CheckChat(t *testing.T, response *types.ChatCompletionResponse, modelName string, usage *types.Usage) {
	assert.NotEmpty(t, response.ID)
	assert.NotEmpty(t, response.Object)
	assert.NotEmpty(t, response.Created)
	assert.Equal(t, response.Model, modelName)
	assert.IsType(t, []types.ChatCompletionChoice{}, response.Choices)
	// check choices 长度大于1
	assert.True(t, len(response.Choices) > 0)
	for _, choice := range response.Choices {
		assert.NotNil(t, choice.Index)
		assert.IsType(t, types.ChatCompletionMessage{}, choice.Message)
		assert.NotEmpty(t, choice.Message.Role)
		assert.NotEmpty(t, choice.FinishReason)

		// check message
		if choice.Message.Content != nil {
			multiContents, ok := choice.Message.Content.([]types.ChatMessagePart)
			if ok {
				for _, content := range multiContents {
					assert.NotEmpty(t, content.Type)
					if content.Type == "text" {
						assert.NotEmpty(t, content.Text)
					} else if content.Type == "image_url" {
						assert.IsType(t, types.ChatMessageImageURL{}, content.ImageURL)
					}
				}
			} else {
				content, ok := choice.Message.Content.(string)
				assert.True(t, ok)
				assert.NotEmpty(t, content)
			}
		} else if choice.Message.FunctionCall != nil {
			assert.NotEmpty(t, choice.Message.FunctionCall.Name)
			assert.Equal(t, choice.FinishReason, types.FinishReasonFunctionCall)
		} else if choice.Message.ToolCalls != nil {
			assert.IsType(t, []types.ChatCompletionToolCalls{}, choice.Message.ToolCalls)
			assert.NotEmpty(t, choice.Message.ToolCalls[0].Id)
			assert.NotEmpty(t, choice.Message.ToolCalls[0].Function)
			assert.Equal(t, choice.Message.ToolCalls[0].Function, "function")

			assert.IsType(t, types.ChatCompletionToolCallsFunction{}, choice.Message.ToolCalls[0].Function)
			assert.NotEmpty(t, choice.Message.ToolCalls[0].Function.Name)

			assert.Equal(t, choice.FinishReason, types.FinishReasonToolCalls)
		} else {
			assert.Fail(t, "message content is nil")
		}
	}

	// check usage
	assert.IsType(t, &types.Usage{}, response.Usage)
	assert.Equal(t, response.Usage.PromptTokens, usage.PromptTokens)
	assert.Equal(t, response.Usage.CompletionTokens, usage.CompletionTokens)
	assert.Equal(t, response.Usage.TotalTokens, usage.TotalTokens)

}
