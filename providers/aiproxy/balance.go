package aiproxy

import (
	"errors"
	"fmt"
)

func (p *AIProxyProvider) Balance() (float64, error) {
	fullRequestURL := "https://aiproxy.io/api/report/getUserOverview"
	headers := make(map[string]string)
	headers["Api-Key"] = p.Channel.Key

	req, err := p.Requester.NewRequest("GET", fullRequestURL, p.Requester.WithHeader(headers))
	if err != nil {
		return 0, err
	}

	// 发送请求
	var response AIProxyUserOverviewResponse
	_, errWithCode := p.Requester.SendRequest(req, &response, false)
	if errWithCode != nil {
		return 0, errors.New(errWithCode.OpenAIError.Message)
	}

	if !response.Success {
		return 0, fmt.Errorf("code: %d, message: %s", response.ErrorCode, response.Message)
	}

	p.Channel.UpdateBalance(response.Data.TotalPoints)

	return response.Data.TotalPoints, nil
}
