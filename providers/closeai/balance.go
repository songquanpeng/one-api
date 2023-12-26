package closeai

import (
	"errors"
	"one-api/common"
	"one-api/model"
)

func (p *CloseaiProxyProvider) Balance(channel *model.Channel) (float64, error) {
	fullRequestURL := p.GetFullRequestURL("/dashboard/billing/credit_grants", "")
	headers := p.GetRequestHeaders()

	client := common.NewClient()
	req, err := client.NewRequest("GET", fullRequestURL, common.WithHeader(headers))
	if err != nil {
		return 0, err
	}

	// 发送请求
	var response OpenAICreditGrants
	_, errWithCode := common.SendRequest(req, &response, false, p.Channel.Proxy)
	if errWithCode != nil {
		return 0, errors.New(errWithCode.OpenAIError.Message)
	}

	channel.UpdateBalance(response.TotalAvailable)

	return response.TotalAvailable, nil
}
