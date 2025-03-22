package openai

import (
	"bufio"
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/songquanpeng/one-api/common/render"

	"github.com/gin-gonic/gin"
	"github.com/songquanpeng/one-api/common"
	"github.com/songquanpeng/one-api/common/conv"
	"github.com/songquanpeng/one-api/common/logger"
	"github.com/songquanpeng/one-api/relay/model"
	"github.com/songquanpeng/one-api/relay/relaymode"
)

const (
	dataPrefix       = "data: "
	done             = "[DONE]"
	dataPrefixLength = len(dataPrefix)
)

func StreamHandler(c *gin.Context, resp *http.Response, relayMode int) (*model.ErrorWithStatusCode, string, *model.Usage) {
	responseText := ""
	scanner := bufio.NewScanner(resp.Body)
	scanner.Split(bufio.ScanLines)
	var usage *model.Usage

	common.SetEventStreamHeaders(c)

	// Variables to track <think> tag state across chunks
	inThinkTag := false
	var reasoningBuilder strings.Builder

	doneRendered := false
	for scanner.Scan() {
		data := scanner.Text()
		if len(data) < dataPrefixLength { // ignore blank line or wrong format
			continue
		}
		if data[:dataPrefixLength] != dataPrefix && data[:dataPrefixLength] != done {
			continue
		}
		if strings.HasPrefix(data[dataPrefixLength:], done) {
			render.StringData(c, data)
			doneRendered = true
			continue
		}
		switch relayMode {
		case relaymode.ChatCompletions:
			var streamResponse ChatCompletionsStreamResponse
			err := json.Unmarshal([]byte(data[dataPrefixLength:]), &streamResponse)
			if err != nil {
				logger.SysError("error unmarshalling stream response: " + err.Error())
				render.StringData(c, data)
				continue
			}
			if len(streamResponse.Choices) == 0 && streamResponse.Usage == nil {
				continue
			}
			
			// Process <think> tags before rendering
			for i := range streamResponse.Choices {
				if streamResponse.Choices[i].Delta.Content != nil {
					content := conv.AsString(streamResponse.Choices[i].Delta.Content)
					logger.Debugf(c.Request.Context(), "Original content: %s", content)
					
					// Process the content for <think> tags
					cleanContent, reasoningContent, newInThinkTag := processStreamThinkTag(content, inThinkTag, &reasoningBuilder)
					inThinkTag = newInThinkTag
					
					// Update content
					streamResponse.Choices[i].Delta.Content = cleanContent
					
					// If there's reasoning content, add it to reasoning_content
					if reasoningContent != "" {
						var reasoningContentAny any = reasoningContent
						streamResponse.Choices[i].Delta.ReasoningContent = reasoningContentAny
						logger.Debugf(c.Request.Context(), "Setting reasoning_content: %s", reasoningContent)
					}
					
					logger.Debugf(c.Request.Context(), "Processed content: clean=%s, reasoning=%s, inThinkTag=%v", 
						cleanContent, reasoningContent, inThinkTag)
				}
			}
			
			// Re-marshal the modified response
			modifiedData, err := json.Marshal(streamResponse)
			if err != nil {
				logger.SysError("error marshalling modified stream response: " + err.Error())
				render.StringData(c, data) // if error happened, pass the original data to client
			} else {
				modifiedDataStr := dataPrefix + string(modifiedData)
				logger.Debugf(c.Request.Context(), "Modified response: %s", modifiedDataStr)
				render.StringData(c, modifiedDataStr)
			}
			
			// Update responseText with cleaned content
			for _, choice := range streamResponse.Choices {
				responseText += conv.AsString(choice.Delta.Content)
			}
			if streamResponse.Usage != nil {
				usage = streamResponse.Usage
			}
		case relaymode.Completions:
			render.StringData(c, data)
			var streamResponse CompletionsStreamResponse
			err := json.Unmarshal([]byte(data[dataPrefixLength:]), &streamResponse)
			if err != nil {
				logger.SysError("error unmarshalling stream response: " + err.Error())
				continue
			}
			for _, choice := range streamResponse.Choices {
				responseText += choice.Text
			}
		}
	}

	if err := scanner.Err(); err != nil {
		logger.SysError("error reading stream: " + err.Error())
	}

	if !doneRendered {
		render.Done(c)
	}

	err := resp.Body.Close()
	if err != nil {
		return ErrorWrapper(err, "close_response_body_failed", http.StatusInternalServerError), "", nil
	}

	return nil, responseText, usage
}

func Handler(c *gin.Context, resp *http.Response, promptTokens int, modelName string) (*model.ErrorWithStatusCode, *model.Usage) {
	var textResponse SlimTextResponse
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return ErrorWrapper(err, "read_response_body_failed", http.StatusInternalServerError), nil
	}
	err = resp.Body.Close()
	if err != nil {
		return ErrorWrapper(err, "close_response_body_failed", http.StatusInternalServerError), nil
	}
	err = json.Unmarshal(responseBody, &textResponse)
	if err != nil {
		return ErrorWrapper(err, "unmarshal_response_body_failed", http.StatusInternalServerError), nil
	}
	if textResponse.Error.Type != "" {
		return &model.ErrorWithStatusCode{
			Error:      textResponse.Error,
			StatusCode: resp.StatusCode,
		}, nil
	}

	// Process <think> tags in the response
	modified := false
	for i := range textResponse.Choices {
		if textResponse.Choices[i].Message.Content != nil {
			content := textResponse.Choices[i].Message.StringContent()
			cleanContent, reasoningContent := extractThinkContent(content)
			
			// If content was modified, update it
			if content != cleanContent || reasoningContent != "" {
				textResponse.Choices[i].Message.Content = cleanContent
				
				// If there's reasoning content, add it to reasoning_content
				if reasoningContent != "" {
					// Make sure ReasoningContent is set as a string, not any other type
					var reasoningContentAny any = reasoningContent
					textResponse.Choices[i].Message.ReasoningContent = reasoningContentAny
				}
				
				modified = true
			}
		}
	}

	// If the response was modified, re-marshal it
	if modified {
		modifiedResponseBody, err := json.Marshal(textResponse)
		if err != nil {
			logger.SysError("error marshalling modified response: " + err.Error())
			// If there's an error, use the original response body
			resp.Body = io.NopCloser(bytes.NewBuffer(responseBody))
		} else {
			// Use the modified response body
			responseBody = modifiedResponseBody
			resp.Body = io.NopCloser(bytes.NewBuffer(responseBody))
		}
	} else {
		// Reset response body with original content
		resp.Body = io.NopCloser(bytes.NewBuffer(responseBody))
	}

	// We shouldn't set the header before we parse the response body, because the parse part may fail.
	// And then we will have to send an error response, but in this case, the header has already been set.
	// So the HTTPClient will be confused by the response.
	// For example, Postman will report error, and we cannot check the response at all.
	for k, v := range resp.Header {
		c.Writer.Header().Set(k, v[0])
	}
	c.Writer.WriteHeader(resp.StatusCode)
	_, err = io.Copy(c.Writer, resp.Body)
	if err != nil {
		return ErrorWrapper(err, "copy_response_body_failed", http.StatusInternalServerError), nil
	}
	err = resp.Body.Close()
	if err != nil {
		return ErrorWrapper(err, "close_response_body_failed", http.StatusInternalServerError), nil
	}

	if textResponse.Usage.TotalTokens == 0 || (textResponse.Usage.PromptTokens == 0 && textResponse.Usage.CompletionTokens == 0) {
		completionTokens := 0
		for _, choice := range textResponse.Choices {
			completionTokens += CountTokenText(choice.Message.StringContent(), modelName)
		}
		textResponse.Usage = model.Usage{
			PromptTokens:     promptTokens,
			CompletionTokens: completionTokens,
			TotalTokens:      promptTokens + completionTokens,
		}
	}
	return nil, &textResponse.Usage
}
