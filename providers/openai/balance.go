package openai

import (
	"errors"
	"fmt"
	"one-api/common"
	"one-api/model"
	"time"
)

func (p *OpenAIProvider) Balance(channel *model.Channel) (float64, error) {
	fullRequestURL := p.GetFullRequestURL("/v1/dashboard/billing/subscription", "")
	headers := p.GetRequestHeaders()

	client := common.NewClient()
	req, err := client.NewRequest("GET", fullRequestURL, common.WithHeader(headers))
	if err != nil {
		return 0, err
	}

	// 发送请求
	var subscription OpenAISubscriptionResponse
	_, errWithCode := common.SendRequest(req, &subscription, false)
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
	req, err = client.NewRequest("GET", fullRequestURL, common.WithHeader(headers))
	if err != nil {
		return 0, err
	}
	usage := OpenAIUsageResponse{}
	_, errWithCode = common.SendRequest(req, &usage, false)

	balance := subscription.HardLimitUSD - usage.TotalUsage/100
	channel.UpdateBalance(balance)
	return balance, nil
}
