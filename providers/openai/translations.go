package openai

import (
	"io"
	"net/http"
	"one-api/common"
	"one-api/common/config"
	"one-api/types"
)

func (p *OpenAIProvider) CreateTranslation(request *types.AudioRequest) (*types.AudioResponseWrapper, *types.OpenAIErrorWithStatusCode) {
	req, errWithCode := p.getRequestAudioBody(config.RelayModeAudioTranslation, request.Model, request)
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
