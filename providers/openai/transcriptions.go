package openai

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"net/http"
	"one-api/common"
	"one-api/common/config"
	"one-api/common/requester"
	"one-api/types"
	"regexp"
	"strconv"
	"strings"
)

func (p *OpenAIProvider) CreateTranscriptions(request *types.AudioRequest) (*types.AudioResponseWrapper, *types.OpenAIErrorWithStatusCode) {
	req, errWithCode := p.getRequestAudioBody(config.RelayModeAudioTranscription, request.Model, request)
	if errWithCode != nil {
		return nil, errWithCode
	}
	defer req.Body.Close()

	var textResponse string
	var resp *http.Response
	var err error
	audioResponseWrapper := &types.AudioResponseWrapper{}
	if hasJSONResponse(request) {
		openAIProviderTranscriptionsResponse := &OpenAIProviderTranscriptionsResponse{}
		resp, errWithCode = p.Requester.SendRequest(req, openAIProviderTranscriptionsResponse, true)
		if errWithCode != nil {
			return nil, errWithCode
		}
		textResponse = openAIProviderTranscriptionsResponse.Text
	} else {
		openAIProviderTranscriptionsTextResponse := new(OpenAIProviderTranscriptionsTextResponse)
		resp, errWithCode = p.Requester.SendRequest(req, openAIProviderTranscriptionsTextResponse, true)
		if errWithCode != nil {
			return nil, errWithCode
		}
		textResponse = getTextContent(*openAIProviderTranscriptionsTextResponse.GetString(), request.ResponseFormat)
	}

	defer resp.Body.Close()

	audioResponseWrapper.Headers = map[string]string{
		"Content-Type": resp.Header.Get("Content-Type"),
	}

	audioResponseWrapper.Body, err = io.ReadAll(resp.Body)
	if err != nil {
		return nil, common.ErrorWrapper(err, "read_response_body_failed", http.StatusInternalServerError)
	}

	completionTokens := common.CountTokenText(textResponse, request.Model)

	p.Usage.CompletionTokens = completionTokens
	p.Usage.TotalTokens = p.Usage.PromptTokens + p.Usage.CompletionTokens

	return audioResponseWrapper, nil
}

func hasJSONResponse(request *types.AudioRequest) bool {
	return request.ResponseFormat == "" || request.ResponseFormat == "json" || request.ResponseFormat == "verbose_json"
}

func (p *OpenAIProvider) getRequestAudioBody(relayMode int, ModelName string, request *types.AudioRequest) (*http.Request, *types.OpenAIErrorWithStatusCode) {
	url, errWithCode := p.GetSupportedAPIUri(relayMode)
	if errWithCode != nil {
		return nil, errWithCode
	}
	// 获取请求地址
	fullRequestURL := p.GetFullRequestURL(url, ModelName)

	// 获取请求头
	headers := p.GetRequestHeaders()
	// 创建请求
	var req *http.Request
	var err error
	if p.OriginalModel != request.Model {
		var formBody bytes.Buffer
		builder := p.Requester.CreateFormBuilder(&formBody)
		if err := audioMultipartForm(request, builder); err != nil {
			return nil, common.ErrorWrapper(err, "create_form_builder_failed", http.StatusInternalServerError)
		}
		req, err = p.Requester.NewRequest(
			http.MethodPost,
			fullRequestURL,
			p.Requester.WithBody(&formBody),
			p.Requester.WithHeader(headers),
			p.Requester.WithContentType(builder.FormDataContentType()))
		req.ContentLength = int64(formBody.Len())
	} else {
		req, err = p.Requester.NewRequest(
			http.MethodPost,
			fullRequestURL,
			p.Requester.WithBody(p.Context.Request.Body),
			p.Requester.WithHeader(headers),
			p.Requester.WithContentType(p.Context.Request.Header.Get("Content-Type")))
		req.ContentLength = p.Context.Request.ContentLength
	}

	if err != nil {
		return nil, common.ErrorWrapper(err, "new_request_failed", http.StatusInternalServerError)
	}

	return req, nil
}

func audioMultipartForm(request *types.AudioRequest, b requester.FormBuilder) error {

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
	var text []string
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
	var text []string
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
