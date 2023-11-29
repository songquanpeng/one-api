package openai

import (
	"net/http"
	"one-api/common"
	"one-api/types"
)

func (p *OpenAIProvider) SpeechAction(request *types.SpeechAudioRequest, isModelMapped bool, promptTokens int) (usage *types.Usage, errWithCode *types.OpenAIErrorWithStatusCode) {

	requestBody, err := p.getRequestBody(&request, isModelMapped)
	if err != nil {
		return nil, types.ErrorWrapper(err, "json_marshal_failed", http.StatusInternalServerError)
	}

	fullRequestURL := p.GetFullRequestURL(p.AudioSpeech, request.Model)
	headers := p.GetRequestHeaders()

	client := common.NewClient()
	req, err := client.NewRequest(p.Context.Request.Method, fullRequestURL, common.WithBody(requestBody), common.WithHeader(headers))
	if err != nil {
		return nil, types.ErrorWrapper(err, "new_request_failed", http.StatusInternalServerError)
	}

	errWithCode = p.SendRequestRaw(req)
	if errWithCode != nil {
		return
	}

	usage = &types.Usage{
		PromptTokens:     promptTokens,
		CompletionTokens: 0,
		TotalTokens:      promptTokens,
	}

	return
}
