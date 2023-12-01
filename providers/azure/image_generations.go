package azure

import (
	"errors"
	"fmt"
	"net/http"
	"one-api/common"
	"one-api/providers/openai"
	"one-api/types"
	"time"
)

func (c *ImageAzureResponse) ResponseHandler(resp *http.Response) (OpenAIResponse any, errWithCode *types.OpenAIErrorWithStatusCode) {
	if c.Status == "canceled" || c.Status == "failed" {
		errWithCode = &types.OpenAIErrorWithStatusCode{
			OpenAIError: types.OpenAIError{
				Message: c.Error.Message,
				Type:    "one_api_error",
				Code:    c.Error.Code,
			},
			StatusCode: resp.StatusCode,
		}
		return
	}

	operation_location := resp.Header.Get("operation-location")
	if operation_location == "" {
		return nil, common.ErrorWrapper(errors.New("image url is empty"), "get_images_url_failed", http.StatusInternalServerError)
	}

	client := common.NewClient()
	req, err := client.NewRequest("GET", operation_location, common.WithHeader(c.Header))
	if err != nil {
		return nil, common.ErrorWrapper(err, "get_images_request_failed", http.StatusInternalServerError)
	}

	getImageAzureResponse := ImageAzureResponse{}
	for i := 0; i < 3; i++ {
		// 休眠 2 秒
		time.Sleep(2 * time.Second)
		_, errWithCode = common.SendRequest(req, &getImageAzureResponse, false)
		fmt.Println("getImageAzureResponse", getImageAzureResponse)
		if errWithCode != nil {
			return
		}

		if getImageAzureResponse.Status == "canceled" || getImageAzureResponse.Status == "failed" {
			return nil, &types.OpenAIErrorWithStatusCode{
				OpenAIError: types.OpenAIError{
					Message: c.Error.Message,
					Type:    "get_images_request_failed",
					Code:    c.Error.Code,
				},
				StatusCode: resp.StatusCode,
			}
		}
		if getImageAzureResponse.Status == "succeeded" {
			return getImageAzureResponse.Result, nil
		}
	}

	return nil, common.ErrorWrapper(errors.New("get image Timeout"), "get_images_url_failed", http.StatusInternalServerError)
}

func (p *AzureProvider) ImageGenerationsAction(request *types.ImageRequest, isModelMapped bool, promptTokens int) (usage *types.Usage, errWithCode *types.OpenAIErrorWithStatusCode) {

	requestBody, err := p.GetRequestBody(&request, isModelMapped)
	if err != nil {
		return nil, common.ErrorWrapper(err, "json_marshal_failed", http.StatusInternalServerError)
	}

	fullRequestURL := p.GetFullRequestURL(p.ImagesGenerations, request.Model)
	headers := p.GetRequestHeaders()

	client := common.NewClient()
	req, err := client.NewRequest(p.Context.Request.Method, fullRequestURL, common.WithBody(requestBody), common.WithHeader(headers))
	if err != nil {
		return nil, common.ErrorWrapper(err, "new_request_failed", http.StatusInternalServerError)
	}

	if request.Model == "dall-e-2" {
		imageAzureResponse := &ImageAzureResponse{
			Header: headers,
		}
		errWithCode = p.SendRequest(req, imageAzureResponse, false)
	} else {
		openAIProviderImageResponseResponse := &openai.OpenAIProviderImageResponseResponse{}
		errWithCode = p.SendRequest(req, openAIProviderImageResponseResponse, true)
	}

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
