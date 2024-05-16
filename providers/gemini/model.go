package gemini

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

func (p *GeminiProvider) GetModelList() ([]string, error) {
	params := url.Values{}
	params.Add("page_size", "1000")
	queryString := params.Encode()

	baseURL := strings.TrimSuffix(p.GetBaseURL(), "/")
	version := "v1beta"
	fullRequestURL := fmt.Sprintf("%s/%s%s?%s", baseURL, version, p.Config.ModelList, queryString)

	headers := p.GetRequestHeaders()

	req, err := p.Requester.NewRequest(http.MethodGet, fullRequestURL, p.Requester.WithHeader(headers))
	if err != nil {
		return nil, errors.New("new_request_failed")
	}

	response := &ModelListResponse{}
	_, errWithCode := p.Requester.SendRequest(req, response, false)
	if errWithCode != nil {
		return nil, errors.New(errWithCode.Message)
	}

	var modelList []string
	for _, model := range response.Models {
		for _, modelType := range model.SupportedGenerationMethods {
			if modelType == "generateContent" {
				modelName := strings.TrimPrefix(model.Name, "models/")
				modelList = append(modelList, modelName)
				break
			}
		}
	}

	return modelList, nil
}
