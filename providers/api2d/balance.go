package api2d

import (
	"errors"
	"one-api/model"
	"one-api/providers/base"
)

func (p *Api2dProvider) Balance(channel *model.Channel) (float64, error) {
	fullRequestURL := p.GetFullRequestURL("/dashboard/billing/credit_grants", "")
	headers := p.GetRequestHeaders()

	req, err := p.Requester.NewRequest("GET", fullRequestURL, p.Requester.WithHeader(headers))
	if err != nil {
		return 0, err
	}

	// 发送请求
	var response base.BalanceResponse
	_, errWithCode := p.Requester.SendRequest(req, &response, false)
	if errWithCode != nil {
		return 0, errors.New(errWithCode.OpenAIError.Message)
	}

	channel.UpdateBalance(response.TotalAvailable)

	return response.TotalAvailable, nil
}
