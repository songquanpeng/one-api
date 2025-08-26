package model

import (
	"fmt"
	"sync"

	"github.com/songquanpeng/one-api/common/logger"
	"github.com/songquanpeng/one-api/relay/channeltype"
)

// ModelAliasMap maps standard model names to channel-specific names
// Format: standardName -> channelType -> actualName
var ModelAliasMap = map[string]map[int]string{
	// OpenAI GPT models
	"gpt-3.5-turbo": {
		channeltype.OpenRouter: "openai/gpt-3.5-turbo",
		channeltype.OpenAI:     "gpt-3.5-turbo",
	},
	"gpt-3.5-turbo-0125": {
		channeltype.OpenRouter: "openai/gpt-3.5-turbo-0125",
		channeltype.OpenAI:     "gpt-3.5-turbo-0125",
	},
	"gpt-4": {
		channeltype.OpenRouter: "openai/gpt-4",
		channeltype.OpenAI:     "gpt-4",
	},
	"gpt-4-turbo": {
		channeltype.OpenRouter: "openai/gpt-4-turbo",
		channeltype.OpenAI:     "gpt-4-turbo",
	},
	"gpt-4o": {
		channeltype.OpenRouter: "openai/gpt-4o",
		channeltype.OpenAI:     "gpt-4o",
	},
	"gpt-4o-mini": {
		channeltype.OpenRouter: "openai/gpt-4o-mini",
		channeltype.OpenAI:     "gpt-4o-mini",
	},
	"o1": {
		channeltype.OpenRouter: "openai/o1",
		channeltype.OpenAI:     "o1",
	},
	"o1-mini": {
		channeltype.OpenRouter: "openai/o1-mini",
		channeltype.OpenAI:     "o1-mini",
	},
	"o1-preview": {
		channeltype.OpenRouter: "openai/o1-preview",
		channeltype.OpenAI:     "o1-preview",
	},

	// Anthropic Claude models
	"claude-3-haiku": {
		channeltype.OpenRouter: "anthropic/claude-3-haiku",
		channeltype.Anthropic:  "claude-3-haiku-20240307",
	},
	"claude-3-sonnet": {
		channeltype.OpenRouter: "anthropic/claude-3-sonnet",
		channeltype.Anthropic:  "claude-3-sonnet-20240229",
	},
	"claude-3-opus": {
		channeltype.OpenRouter: "anthropic/claude-3-opus",
		channeltype.Anthropic:  "claude-3-opus-20240229",
	},
	"claude-3.5-sonnet": {
		channeltype.OpenRouter: "anthropic/claude-3.5-sonnet",
		channeltype.Anthropic:  "claude-3-5-sonnet-20241022",
	},
	"claude-3.5-haiku": {
		channeltype.OpenRouter: "anthropic/claude-3.5-haiku",
		channeltype.Anthropic:  "claude-3-5-haiku-20241022",
	},

	// Google models
	"gemini-pro": {
		channeltype.OpenRouter: "google/gemini-pro",
		channeltype.Gemini:     "gemini-pro",
	},
	"gemini-pro-1.5": {
		channeltype.OpenRouter: "google/gemini-pro-1.5",
		channeltype.Gemini:     "gemini-1.5-pro-latest",
	},
	"gemini-flash-1.5": {
		channeltype.OpenRouter: "google/gemini-flash-1.5",
		channeltype.Gemini:     "gemini-1.5-flash-latest",
	},

	// Meta Llama models
	"llama-3-8b-instruct": {
		channeltype.OpenRouter: "meta-llama/llama-3-8b-instruct",
		channeltype.Groq:       "llama3-8b-8192",
	},
	"llama-3-70b-instruct": {
		channeltype.OpenRouter: "meta-llama/llama-3-70b-instruct",
		channeltype.Groq:       "llama3-70b-8192",
	},
	"llama-3.1-8b-instruct": {
		channeltype.OpenRouter: "meta-llama/llama-3.1-8b-instruct",
		channeltype.Groq:       "llama-3.1-8b-instant",
	},
	"llama-3.1-70b-instruct": {
		channeltype.OpenRouter: "meta-llama/llama-3.1-70b-instruct",
		channeltype.Groq:       "llama-3.1-70b-versatile",
	},

	// Mistral models
	"mistral-7b-instruct": {
		channeltype.OpenRouter: "mistralai/mistral-7b-instruct",
		channeltype.Mistral:    "mistral-small-latest",
	},
	"mixtral-8x7b-instruct": {
		channeltype.OpenRouter: "mistralai/mixtral-8x7b-instruct",
		channeltype.Mistral:    "mixtral-8x7b-instruct-v0.1",
	},
}

var aliasLock sync.RWMutex

// ResolveModelAlias resolves a standard model name to channel-specific name
func ResolveModelAlias(standardName string, channelType int) string {
	aliasLock.RLock()
	defer aliasLock.RUnlock()

	if channelMap, exists := ModelAliasMap[standardName]; exists {
		if actualName, exists := channelMap[channelType]; exists {
			return actualName
		}
	}

	// If no alias found, return original name
	return standardName
}

// RegisterModelAlias registers a new model alias
func RegisterModelAlias(standardName, actualName string, channelType int) {
	aliasLock.Lock()
	defer aliasLock.Unlock()

	if ModelAliasMap[standardName] == nil {
		ModelAliasMap[standardName] = make(map[int]string)
	}

	ModelAliasMap[standardName][channelType] = actualName
	logger.SysLog(fmt.Sprintf("registered alias: %s -> %s (channel type %d)", standardName, actualName, channelType))
}

// GetAllModelAliases returns all registered aliases
func GetAllModelAliases() map[string]map[int]string {
	aliasLock.RLock()
	defer aliasLock.RUnlock()

	// Return a copy to prevent external modification
	result := make(map[string]map[int]string)
	for standard, channelMap := range ModelAliasMap {
		result[standard] = make(map[int]string)
		for channelType, actual := range channelMap {
			result[standard][channelType] = actual
		}
	}

	return result
}

// GetStandardModelName returns the standard name for a channel-specific model
func GetStandardModelName(actualName string, channelType int) string {
	aliasLock.RLock()
	defer aliasLock.RUnlock()

	for standard, channelMap := range ModelAliasMap {
		if channelMap[channelType] == actualName {
			return standard
		}
	}

	// If no reverse mapping found, return original name
	return actualName
}

// IsAliasSupported checks if a standard model name has aliases for given channel type
func IsAliasSupported(standardName string, channelType int) bool {
	aliasLock.RLock()
	defer aliasLock.RUnlock()

	if channelMap, exists := ModelAliasMap[standardName]; exists {
		_, exists := channelMap[channelType]
		return exists
	}

	return false
}

// GetSupportedChannelTypes returns all channel types that support the given standard model
func GetSupportedChannelTypes(standardName string) []int {
	aliasLock.RLock()
	defer aliasLock.RUnlock()

	var channelTypes []int
	if channelMap, exists := ModelAliasMap[standardName]; exists {
		for channelType := range channelMap {
			channelTypes = append(channelTypes, channelType)
		}
	}

	return channelTypes
}

// LoadModelAliasesFromConfig loads aliases from configuration (placeholder for future implementation)
func LoadModelAliasesFromConfig() error {
	// TODO: Implement loading from database or configuration file
	logger.SysLog("loaded model aliases from static configuration")
	return nil
}
