package api2gpt

import (
	"errors"
	"one-api/common"
	"one-api/model"
	"one-api/providers/base"
)

func (p *Api2gptProvider) Balance(channel *model.Channel) (float64, error) {
	fullRequestURL := p.GetFullRequestURL("/dashboard/billing/credit_grants", "")
	headers := p.GetRequestHeaders()

	client := common.NewClient()
	req, err := client.NewRequest("GET", fullRequestURL, common.WithHeader(headers))
	if err != nil {
		return 0, err
	}

	// 发送请求
	var response base.BalanceResponse
	_, errWithCode := common.SendRequest(req, &response, false)
	if errWithCode != nil {
		return 0, errors.New(errWithCode.OpenAIError.Message)
	}

	channel.UpdateBalance(response.TotalAvailable)

	return response.TotalRemaining, nil
}
