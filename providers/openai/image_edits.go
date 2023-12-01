package openai

import (
	"bytes"
	"fmt"
	"net/http"
	"one-api/common"
	"one-api/types"
)

func (p *OpenAIProvider) ImageEditsAction(request *types.ImageEditRequest, isModelMapped bool, promptTokens int) (usage *types.Usage, errWithCode *types.OpenAIErrorWithStatusCode) {
	fullRequestURL := p.GetFullRequestURL(p.ImagesEdit, request.Model)
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

func imagesEditsMultipartForm(request *types.ImageEditRequest, b common.FormBuilder) error {
	err := b.CreateFormFile("image", request.Image)
	if err != nil {
		return fmt.Errorf("creating form file: %w", err)
	}

	err = b.WriteField("prompt", request.Prompt)
	if err != nil {
		return fmt.Errorf("writing prompt: %w", err)
	}

	err = b.WriteField("model", request.Model)
	if err != nil {
		return fmt.Errorf("writing model name: %w", err)
	}

	if request.Mask != nil {
		err = b.CreateFormFile("mask", request.Mask)
		if err != nil {
			return fmt.Errorf("writing format: %w", err)
		}
	}

	if request.ResponseFormat != "" {
		err = b.WriteField("response_format", request.ResponseFormat)
		if err != nil {
			return fmt.Errorf("writing format: %w", err)
		}
	}

	if request.N != 0 {
		err = b.WriteField("n", fmt.Sprintf("%.2f", request.N))
		if err != nil {
			return fmt.Errorf("writing temperature: %w", err)
		}
	}

	if request.Size != "" {
		err = b.WriteField("size", request.Size)
		if err != nil {
			return fmt.Errorf("writing language: %w", err)
		}
	}

	if request.User != "" {
		err = b.WriteField("user", request.User)
		if err != nil {
			return fmt.Errorf("writing language: %w", err)
		}
	}

	return b.Close()
}
