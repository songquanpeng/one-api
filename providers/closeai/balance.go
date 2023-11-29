package closeai

import (
	"fmt"
	"one-api/common"
	"one-api/model"
)

func (p *CloseaiProxyProvider) Balance(channel *model.Channel) (float64, error) {
	fullRequestURL := p.GetFullRequestURL("/sb-api/user/status", "")
	fullRequestURL = fmt.Sprintf("%s?api_key=%s", fullRequestURL, channel.Key)
	headers := p.GetRequestHeaders()

	client := common.NewClient()
	req, err := client.NewRequest("GET", fullRequestURL, common.WithBody(nil), common.WithHeader(headers))
	if err != nil {
		return 0, err
	}

	// 发送请求
	var response OpenAICreditGrants
	err = client.SendRequest(req, &response)
	if err != nil {
		return 0, err
	}

	channel.UpdateBalance(response.TotalAvailable)

	return response.TotalAvailable, nil
}
