package util

import (
	"errors"
	"github.com/songquanpeng/one-api/relay/constant"
	"github.com/songquanpeng/one-api/relay/model"
	"math"
)

func ValidateTextRequest(textRequest *model.GeneralOpenAIRequest, relayMode int) error {
	if textRequest.MaxTokens < 0 || textRequest.MaxTokens > math.MaxInt32/2 {
		return errors.New("max_tokens is invalid")
	}
	if textRequest.Model == "" {
		return errors.New("model is required")
	}
	switch relayMode {
	case constant.RelayModeCompletions:
		if textRequest.Prompt == "" {
			return errors.New("field prompt is required")
		}
	case constant.RelayModeChatCompletions:
		if textRequest.Messages == nil || len(textRequest.Messages) == 0 {
			return errors.New("field messages is required")
		}
	case constant.RelayModeEmbeddings:
	case constant.RelayModeModerations:
		if textRequest.Input == "" {
			return errors.New("field input is required")
		}
	case constant.RelayModeEdits:
		if textRequest.Instruction == "" {
			return errors.New("field instruction is required")
		}
	}
	return nil
}
