package aiproxy

import (
	"errors"
	"fmt"
	"one-api/common"
	"one-api/model"
)

func (p *AIProxyProvider) Balance(channel *model.Channel) (float64, error) {
	fullRequestURL := "https://aiproxy.io/api/report/getUserOverview"
	headers := make(map[string]string)
	headers["Api-Key"] = channel.Key

	client := common.NewClient()
	req, err := client.NewRequest("GET", fullRequestURL, common.WithHeader(headers))
	if err != nil {
		return 0, err
	}

	// 发送请求
	var response AIProxyUserOverviewResponse
	_, errWithCode := common.SendRequest(req, &response, false, p.Channel.Proxy)
	if errWithCode != nil {
		return 0, errors.New(errWithCode.OpenAIError.Message)
	}

	if !response.Success {
		return 0, fmt.Errorf("code: %d, message: %s", response.ErrorCode, response.Message)
	}

	channel.UpdateBalance(response.Data.TotalPoints)

	return response.Data.TotalPoints, nil
}
