package cloudflareAI

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"one-api/common"
	"one-api/common/requester"
	"one-api/types"
)

func (p *CloudflareAIProvider) CreateTranscriptions(request *types.AudioRequest) (*types.AudioResponseWrapper, *types.OpenAIErrorWithStatusCode) {
	req, errWithCode := p.getRequestAudioBody(request.Model, request)
	if errWithCode != nil {
		return nil, errWithCode
	}
	defer req.Body.Close()

	var resp *http.Response
	var err error

	audioResponse := &AudioResponse{}
	resp, errWithCode = p.Requester.SendRequest(req, audioResponse, false)
	if errWithCode != nil {
		return nil, errWithCode
	}

	errWithOP := errorHandle(&audioResponse.CloudflareAIError)
	if errWithOP != nil {
		errWithCode = &types.OpenAIErrorWithStatusCode{
			OpenAIError: *errWithOP,
			StatusCode:  http.StatusBadRequest,
		}
		return nil, errWithCode
	}

	chatResult := audioResponse.Result

	audioResponseWrapper := &types.AudioResponseWrapper{}
	audioResponseWrapper.Headers = map[string]string{
		"Content-Type": resp.Header.Get("Content-Type"),
	}

	audioResponseWrapper.Body, err = json.Marshal(&chatResult)
	if err != nil {
		return nil, common.ErrorWrapper(err, "read_response_body_failed", http.StatusInternalServerError)
	}

	completionTokens := common.CountTokenText(chatResult.Text, request.Model)

	p.Usage.CompletionTokens = completionTokens
	p.Usage.TotalTokens = p.Usage.PromptTokens + p.Usage.CompletionTokens

	return audioResponseWrapper, nil
}

func (p *CloudflareAIProvider) getRequestAudioBody(ModelName string, request *types.AudioRequest) (*http.Request, *types.OpenAIErrorWithStatusCode) {
	// 获取请求地址
	fullRequestURL := p.GetFullRequestURL(ModelName)

	// 获取请求头
	headers := p.GetRequestHeaders()
	// 创建请求
	var req *http.Request
	var err error

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
	return b.Close()
}
