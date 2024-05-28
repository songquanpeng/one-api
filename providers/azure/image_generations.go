package azure

import (
	"errors"
	"net/http"
	"one-api/common"
	"one-api/common/config"
	"one-api/providers/openai"
	"one-api/types"
	"time"
)

func (p *AzureProvider) CreateImageGenerations(request *types.ImageRequest) (*types.ImageResponse, *types.OpenAIErrorWithStatusCode) {
	if !openai.IsWithinRange(request.Model, request.N) {
		return nil, common.StringErrorWrapper("n_not_within_range", "n_not_within_range", http.StatusBadRequest)
	}

	req, errWithCode := p.GetRequestTextBody(config.RelayModeImagesGenerations, request.Model, request)
	if errWithCode != nil {
		return nil, errWithCode
	}
	defer req.Body.Close()

	var response *types.ImageResponse
	var resp *http.Response
	if request.Model == "dall-e-2" {
		imageAzureResponse := &ImageAzureResponse{}
		resp, errWithCode = p.Requester.SendRequest(req, imageAzureResponse, false)
		if errWithCode != nil {
			return nil, errWithCode
		}
		response, errWithCode = p.ResponseAzureImageHandler(resp, imageAzureResponse)
		if errWithCode != nil {
			return nil, errWithCode
		}
	} else {
		var openaiResponse openai.OpenAIProviderImageResponse
		_, errWithCode = p.Requester.SendRequest(req, &openaiResponse, false)
		if errWithCode != nil {
			return nil, errWithCode
		}
		// 检测是否错误
		openaiErr := openai.ErrorHandle(&openaiResponse.OpenAIErrorResponse)
		if openaiErr != nil {
			errWithCode = &types.OpenAIErrorWithStatusCode{
				OpenAIError: *openaiErr,
				StatusCode:  http.StatusBadRequest,
			}
			return nil, errWithCode
		}
		response = &openaiResponse.ImageResponse
	}

	p.Usage.TotalTokens = p.Usage.PromptTokens

	return response, nil

}

func (p *AzureProvider) ResponseAzureImageHandler(resp *http.Response, azure *ImageAzureResponse) (OpenAIResponse *types.ImageResponse, errWithCode *types.OpenAIErrorWithStatusCode) {
	if azure.Status == "canceled" || azure.Status == "failed" {
		errWithCode = &types.OpenAIErrorWithStatusCode{
			OpenAIError: types.OpenAIError{
				Message: azure.Error.Message,
				Type:    "one_api_error",
				Code:    azure.Error.Code,
			},
			StatusCode: resp.StatusCode,
		}
		return
	}

	operationLocation := resp.Header.Get("operation-location")
	if operationLocation == "" {
		return nil, common.ErrorWrapper(errors.New("image url is empty"), "get_images_url_failed", http.StatusInternalServerError)
	}

	req, err := p.Requester.NewRequest("GET", operationLocation, p.Requester.WithHeader(p.GetRequestHeaders()))
	if err != nil {
		return nil, common.ErrorWrapper(err, "get_images_request_failed", http.StatusInternalServerError)
	}

	getImageAzureResponse := ImageAzureResponse{}
	for i := 0; i < 3; i++ {
		// 休眠 2 秒
		time.Sleep(2 * time.Second)
		_, errWithCode = p.Requester.SendRequest(req, &getImageAzureResponse, false)
		if errWithCode != nil {
			return
		}

		if getImageAzureResponse.Status == "canceled" || getImageAzureResponse.Status == "failed" {
			return nil, &types.OpenAIErrorWithStatusCode{
				OpenAIError: types.OpenAIError{
					Message: getImageAzureResponse.Error.Message,
					Type:    "get_images_request_failed",
					Code:    getImageAzureResponse.Error.Code,
				},
				StatusCode: resp.StatusCode,
			}
		}
		if getImageAzureResponse.Status == "succeeded" {
			return &getImageAzureResponse.Result, nil
		}
	}

	return nil, common.ErrorWrapper(errors.New("get image Timeout"), "get_images_url_failed", http.StatusInternalServerError)
}
