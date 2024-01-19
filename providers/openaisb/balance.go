package openaisb

import (
	"errors"
	"fmt"
	"strconv"
)

func (p *OpenaiSBProvider) Balance() (float64, error) {
	fullRequestURL := p.GetFullRequestURL("/sb-api/user/status", "")
	fullRequestURL = fmt.Sprintf("%s?api_key=%s", fullRequestURL, p.Channel.Key)
	headers := p.GetRequestHeaders()

	req, err := p.Requester.NewRequest("GET", fullRequestURL, p.Requester.WithHeader(headers))
	if err != nil {
		return 0, err
	}

	// 发送请求
	var response OpenAISBUsageResponse
	_, errWithCode := p.Requester.SendRequest(req, &response, false)
	if err != nil {
		return 0, errors.New(errWithCode.OpenAIError.Message)
	}

	if response.Data == nil {
		return 0, errors.New(response.Msg)
	}
	balance, err := strconv.ParseFloat(response.Data.Credit, 64)
	if err != nil {
		return 0, err
	}
	p.Channel.UpdateBalance(balance)
	return balance, nil
}
