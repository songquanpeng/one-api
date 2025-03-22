package openai

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/songquanpeng/one-api/common/logger"
	"github.com/songquanpeng/one-api/relay/model"
)

func ErrorWrapper(err error, code string, statusCode int) *model.ErrorWithStatusCode {
	logger.Error(context.TODO(), fmt.Sprintf("[%s]%+v", code, err))

	Error := model.Error{
		Message: err.Error(),
		Type:    "one_api_error",
		Code:    code,
	}
	return &model.ErrorWithStatusCode{
		Error:      Error,
		StatusCode: statusCode,
	}
}

// extractThinkContent extracts the content inside <think> tags and returns the cleaned content and reasoning content.
// The cleaned content is the original content with all <think> tags and their content removed.
// The reasoning content is the concatenation of all content inside <think> tags.
func extractThinkContent(content string) (cleanContent string, reasoningContent string) {
	// If content is nil or empty, return as is
	if content == "" {
		return content, ""
	}

	// Use regular expression to match <think>...</think> tags
	re := regexp.MustCompile(`<think>([\s\S]*?)</think>`)
	matches := re.FindAllStringSubmatch(content, -1)
	
	if len(matches) == 0 {
		// No <think> tags found, return original content
		return content, ""
	}
	
	// Extract all content inside <think> tags
	var reasoningBuilder strings.Builder
	for _, match := range matches {
		if len(match) > 1 {
			reasoningBuilder.WriteString(match[1])
			reasoningBuilder.WriteString("\n")
		}
	}
	
	// Remove all <think> tags and their content from the original content
	cleanContent = re.ReplaceAllString(content, "")
	
	// Fix multiple spaces that might have been created
	spaceRe := regexp.MustCompile(`\s+`)
	cleanContent = spaceRe.ReplaceAllString(cleanContent, " ")
	
	// Remove any extra whitespace that might have been created
	cleanContent = strings.TrimSpace(cleanContent)
	
	return cleanContent, reasoningBuilder.String()
}

// processStreamThinkTag processes a chunk of content for <think> tags in streaming mode.
// It handles partial tags that may be split across chunks.
// Returns:
//   - cleanContent: the content with <think> tags removed
//   - reasoningContent: the content inside <think> tags
//   - inThinkTag: whether we're currently inside a <think> tag
func processStreamThinkTag(content string, inThinkTag bool, reasoningBuilder *strings.Builder) (cleanContent string, reasoningContent string, stillInThinkTag bool) {
	if content == "" {
		return content, "", inThinkTag
	}
	
	// Initialize reasoningContent as empty
	reasoningContent = ""
	
	// Handle case where content contains both <think> and </think> tags
	if strings.Contains(content, "<think>") && strings.Contains(content, "</think>") {
		// Extract content before <think> tag
		beforeThink := strings.Split(content, "<think>")[0]
		
		// Extract content between <think> and </think> tags
		betweenTags := strings.Split(strings.Split(content, "<think>")[1], "</think>")[0]
		reasoningBuilder.WriteString(betweenTags)
		reasoningContent = betweenTags
		
		// Extract content after </think> tag
		afterThink := strings.Split(content, "</think>")[1]
		
		// Combine content before and after tags
		cleanContent = beforeThink + afterThink
		
		// Fix multiple spaces that might have been created
		spaceRe := regexp.MustCompile(`\s+`)
		cleanContent = spaceRe.ReplaceAllString(cleanContent, " ")
		
		// Remove any extra whitespace that might have been created
		cleanContent = strings.TrimSpace(cleanContent)
		
		stillInThinkTag = false
		
		return cleanContent, reasoningContent, stillInThinkTag
	}
	
	// Handle other cases
	switch {
	case strings.Contains(content, "<think>"):
		stillInThinkTag = true
		parts := strings.Split(content, "<think>")
		if len(parts) > 0 && parts[0] != "" {
			cleanContent = parts[0]
		} else {
			cleanContent = ""
		}
		
		if len(parts) > 1 {
			reasoningBuilder.WriteString(parts[1])
			reasoningContent = parts[1]
		}
		
	case strings.Contains(content, "</think>"):
		stillInThinkTag = false
		parts := strings.Split(content, "</think>")
		if len(parts) > 1 && parts[1] != "" {
			cleanContent = parts[1]
		} else {
			cleanContent = ""
		}
		
		if len(parts) > 0 {
			reasoningBuilder.WriteString(parts[0])
			reasoningContent = parts[0]
		}
		
	case inThinkTag:
		reasoningBuilder.WriteString(content)
		reasoningContent = content
		cleanContent = ""
		stillInThinkTag = true
		
	default:
		cleanContent = content
		stillInThinkTag = false
	}
	
	return cleanContent, reasoningContent, stillInThinkTag
}
