package openai

import (
	"errors"
	"fmt"
	"time"
)

func (p *OpenAIProvider) Balance() (float64, error) {
	if !p.BalanceAction {
		return 0, errors.New("不支持余额查询")
	}

	fullRequestURL := p.GetFullRequestURL("/v1/dashboard/billing/subscription", "")
	headers := p.GetRequestHeaders()

	req, err := p.Requester.NewRequest("GET", fullRequestURL, p.Requester.WithHeader(headers))
	if err != nil {
		return 0, err
	}

	// 发送请求
	var subscription OpenAISubscriptionResponse
	_, errWithCode := p.Requester.SendRequest(req, &subscription, false)
	if errWithCode != nil {
		return 0, errors.New(errWithCode.OpenAIError.Message)
	}

	now := time.Now()
	startDate := fmt.Sprintf("%s-01", now.Format("2006-01"))
	endDate := now.Format("2006-01-02")
	if !subscription.HasPaymentMethod {
		startDate = now.AddDate(0, 0, -100).Format("2006-01-02")
	}

	fullRequestURL = p.GetFullRequestURL(fmt.Sprintf("/v1/dashboard/billing/usage?start_date=%s&end_date=%s", startDate, endDate), "")
	req, err = p.Requester.NewRequest("GET", fullRequestURL, p.Requester.WithHeader(headers))
	if err != nil {
		return 0, err
	}
	usage := OpenAIUsageResponse{}
	_, errWithCode = p.Requester.SendRequest(req, &usage, false)
	if errWithCode != nil {
		return 0, errWithCode
	}

	balance := subscription.HardLimitUSD - usage.TotalUsage/100
	p.Channel.UpdateBalance(balance)
	return balance, nil
}
