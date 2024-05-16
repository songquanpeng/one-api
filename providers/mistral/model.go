package mistral

import (
	"errors"
	"net/http"
)

func (p *MistralProvider) GetModelList() ([]string, error) {
	fullRequestURL := p.GetFullRequestURL(p.Config.ModelList, "")
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
	for _, model := range response.Data {
		modelList = append(modelList, model.Id)
	}

	return modelList, nil
}
