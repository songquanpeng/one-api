package openai

import (
	"bytes"
	"net/http"
	"one-api/common"
	"one-api/types"
)

func (p *OpenAIProvider) ImageVariationsAction(request *types.ImageEditRequest, isModelMapped bool, promptTokens int) (usage *types.Usage, errWithCode *types.OpenAIErrorWithStatusCode) {
	fullRequestURL := p.GetFullRequestURL(p.ImagesVariations, request.Model)
	headers := p.GetRequestHeaders()

	client := common.NewClient()

	var formBody bytes.Buffer
	var req *http.Request
	var err error
	if isModelMapped {
		builder := client.CreateFormBuilder(&formBody)
		if err := imagesEditsMultipartForm(request, builder); err != nil {
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

	openAIProviderImageResponseResponse := &OpenAIProviderImageResponseResponse{}
	errWithCode = p.SendRequest(req, openAIProviderImageResponseResponse, true)
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
