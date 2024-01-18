package test

import (
	"encoding/json"
	"one-api/types"
	"strings"
)

func GetChatCompletionRequest(chatType, modelName, stream string) *types.ChatCompletionRequest {
	chatJSON := GetChatRequest(chatType, modelName, stream)
	chatCompletionRequest := &types.ChatCompletionRequest{}
	json.NewDecoder(chatJSON).Decode(chatCompletionRequest)
	return chatCompletionRequest
}

func GetChatRequest(chatType, modelName, stream string) *strings.Reader {
	var chatJSON string
	switch chatType {
	case "image":
		chatJSON = `{
			"model": "` + modelName + `",
			"messages": [
			  {
				"role": "user",
				"content": [
				  {
					"type": "text",
					"text": "What’s in this image?"
				  },
				  {
					"type": "image_url",
					"image_url": {
					  "url": "https://upload.wikimedia.org/wikipedia/commons/thumb/d/dd/Gfp-wisconsin-madison-the-nature-boardwalk.jpg/2560px-Gfp-wisconsin-madison-the-nature-boardwalk.jpg"
					}
				  }
				]
			  }
			],
			"max_tokens": 300,
			"stream": ` + stream + `
			}`
	case "default":
		chatJSON = `{
			"model": "` + modelName + `",
			"messages": [
			  {
				"role": "system",
				"content": "You are a helpful assistant."
			  },
			  {
				"role": "user",
				"content": "Hello!"
			  }
			],
			"stream": ` + stream + `
		  }`
	case "function":
		chatJSON = `{
			"model": "` + modelName + `",
			"stream": ` + stream + `,
			"messages": [
			  {
				"role": "user",
				"content": "What is the weather like in Boston?"
			  }
			],
			"tools": [
			  {
				"type": "function",
				"function": {
				  "name": "get_current_weather",
				  "description": "Get the current weather in a given location",
				  "parameters": {
					"type": "object",
					"properties": {
					  "location": {
						"type": "string",
						"description": "The city and state, e.g. San Francisco, CA"
					  },
					  "unit": {
						"type": "string",
						"enum": ["celsius", "fahrenheit"]
					  }
					},
					"required": ["location"]
				  }
				}
			  }
			],
			"tool_choice": "auto"
		  }`

	case "tools":
		chatJSON = `{
			"model": "` + modelName + `",
			"stream": ` + stream + `,
			"messages": [
			  {
				"role": "user",
				"content": "What is the weather like in Boston?"
			  }
			],
			"functions": [
			  {
				"name": "get_current_weather",
				"description": "Get the current weather in a given location",
				"parameters": {
				  "type": "object",
				  "properties": {
					"location": {
					  "type": "string",
					  "description": "The city and state, e.g. San Francisco, CA"
					},
					"unit": {
					  "type": "string",
					  "enum": [
						"celsius",
						"fahrenheit"
					  ]
					}
				  },
				  "required": [
					"location"
				  ]
				}
			  }
			]
		  }`
	}

	return strings.NewReader(chatJSON)
}
