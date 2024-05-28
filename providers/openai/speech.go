package openai

import (
	"net/http"
	"one-api/common/config"
	"one-api/common/requester"
	"one-api/types"
)

func (p *OpenAIProvider) CreateSpeech(request *types.SpeechAudioRequest) (*http.Response, *types.OpenAIErrorWithStatusCode) {
	req, errWithCode := p.GetRequestTextBody(config.RelayModeAudioSpeech, request.Model, request)
	if errWithCode != nil {
		return nil, errWithCode
	}
	defer req.Body.Close()

	// 发送请求
	var resp *http.Response
	resp, errWithCode = p.Requester.SendRequestRaw(req)
	if errWithCode != nil {
		return nil, errWithCode
	}

	if resp.Header.Get("Content-Type") == "application/json" {
		return nil, requester.HandleErrorResp(resp, p.Requester.ErrorHandler, p.Requester.IsOpenAI)
	}

	p.Usage.TotalTokens = p.Usage.PromptTokens

	return resp, nil
}
