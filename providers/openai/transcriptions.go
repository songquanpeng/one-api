package openai

import (
	"bufio"
	"bytes"
	"fmt"
	"net/http"
	"one-api/common"
	"one-api/types"
	"regexp"
	"strconv"
	"strings"
)

func (c *OpenAIProviderTranscriptionsResponse) ResponseHandler(resp *http.Response) (OpenAIResponse any, errWithCode *types.OpenAIErrorWithStatusCode) {
	if c.Error.Type != "" {
		errWithCode = &types.OpenAIErrorWithStatusCode{
			OpenAIError: c.Error,
			StatusCode:  resp.StatusCode,
		}
		return
	}
	return nil, nil
}

func (c *OpenAIProviderTranscriptionsTextResponse) ResponseHandler(resp *http.Response) (OpenAIResponse any, errWithCode *types.OpenAIErrorWithStatusCode) {
	return nil, nil
}

func (p *OpenAIProvider) TranscriptionsAction(request *types.AudioRequest, isModelMapped bool, promptTokens int) (usage *types.Usage, errWithCode *types.OpenAIErrorWithStatusCode) {
	fullRequestURL := p.GetFullRequestURL(p.AudioTranscriptions, request.Model)
	headers := p.GetRequestHeaders()

	client := common.NewClient()

	var formBody bytes.Buffer
	var req *http.Request
	var err error
	if isModelMapped {
		builder := client.CreateFormBuilder(&formBody)
		if err := audioMultipartForm(request, builder); err != nil {
			return nil, types.ErrorWrapper(err, "create_form_builder_failed", http.StatusInternalServerError)
		}
		req, err = client.NewRequest(p.Context.Request.Method, fullRequestURL, common.WithBody(&formBody), common.WithHeader(headers), common.WithContentType(builder.FormDataContentType()))
		req.ContentLength = int64(formBody.Len())

	} else {
		req, err = client.NewRequest(p.Context.Request.Method, fullRequestURL, common.WithBody(p.Context.Request.Body), common.WithHeader(headers), common.WithContentType(p.Context.Request.Header.Get("Content-Type")))
		req.ContentLength = p.Context.Request.ContentLength
	}

	if err != nil {
		return nil, types.ErrorWrapper(err, "new_request_failed", http.StatusInternalServerError)
	}

	var textResponse string
	if hasJSONResponse(request) {
		openAIProviderTranscriptionsResponse := &OpenAIProviderTranscriptionsResponse{}
		errWithCode = p.SendRequest(req, openAIProviderTranscriptionsResponse, true)
		if errWithCode != nil {
			return
		}
		textResponse = openAIProviderTranscriptionsResponse.Text
	} else {
		openAIProviderTranscriptionsTextResponse := new(OpenAIProviderTranscriptionsTextResponse)
		errWithCode = p.SendRequest(req, openAIProviderTranscriptionsTextResponse, true)
		if errWithCode != nil {
			return
		}
		textResponse = getTextContent(*openAIProviderTranscriptionsTextResponse.GetString(), request.ResponseFormat)
	}

	completionTokens := common.CountTokenText(textResponse, request.Model)
	usage = &types.Usage{
		PromptTokens:     promptTokens,
		CompletionTokens: completionTokens,
		TotalTokens:      promptTokens + completionTokens,
	}
	return
}

func hasJSONResponse(request *types.AudioRequest) bool {
	return request.ResponseFormat == "" || request.ResponseFormat == "json" || request.ResponseFormat == "verbose_json"
}

func audioMultipartForm(request *types.AudioRequest, b common.FormBuilder) error {

	err := b.CreateFormFile("file", request.File)
	if err != nil {
		return fmt.Errorf("creating form file: %w", err)
	}

	err = b.WriteField("model", request.Model)
	if err != nil {
		return fmt.Errorf("writing model name: %w", err)
	}

	if request.Prompt != "" {
		err = b.WriteField("prompt", request.Prompt)
		if err != nil {
			return fmt.Errorf("writing prompt: %w", err)
		}
	}

	if request.ResponseFormat != "" {
		err = b.WriteField("response_format", request.ResponseFormat)
		if err != nil {
			return fmt.Errorf("writing format: %w", err)
		}
	}

	if request.Temperature != 0 {
		err = b.WriteField("temperature", fmt.Sprintf("%.2f", request.Temperature))
		if err != nil {
			return fmt.Errorf("writing temperature: %w", err)
		}
	}

	if request.Language != "" {
		err = b.WriteField("language", request.Language)
		if err != nil {
			return fmt.Errorf("writing language: %w", err)
		}
	}

	return b.Close()
}

func getTextContent(text, format string) string {
	switch format {
	case "srt":
		return extractTextFromSRT(text)
	case "vtt":
		return extractTextFromVTT(text)
	default:
		return text
	}
}

func extractTextFromVTT(vttContent string) string {
	scanner := bufio.NewScanner(strings.NewReader(vttContent))
	re := regexp.MustCompile(`\d{2}:\d{2}:\d{2}\.\d{3} --> \d{2}:\d{2}:\d{2}\.\d{3}`)
	text := []string{}
	isStart := true

	for scanner.Scan() {
		line := scanner.Text()
		if isStart && strings.HasPrefix(line, "WEBVTT") {
			isStart = false
			continue
		}
		if !re.MatchString(line) && !isNumber(line) && line != "" {
			text = append(text, line)
		}
	}

	return strings.Join(text, " ")
}

func extractTextFromSRT(srtContent string) string {
	scanner := bufio.NewScanner(strings.NewReader(srtContent))
	re := regexp.MustCompile(`\d{2}:\d{2}:\d{2},\d{3} --> \d{2}:\d{2}:\d{2},\d{3}`)
	text := []string{}
	isContent := false

	for scanner.Scan() {
		line := scanner.Text()
		if re.MatchString(line) {
			isContent = true
		} else if line == "" {
			isContent = false
		} else if isContent {
			text = append(text, line)
		}
	}

	return strings.Join(text, " ")
}

func isNumber(s string) bool {
	_, err := strconv.Atoi(s)
	return err == nil
}
