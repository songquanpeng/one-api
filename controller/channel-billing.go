package controller

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"io"
	"net/http"
	"one-api/common"
	"one-api/model"
	"strconv"
	"time"
)

// https://github.com/songquanpeng/one-api/issues/79

type OpenAISubscriptionResponse struct {
	Object             string  `json:"object"`
	HasPaymentMethod   bool    `json:"has_payment_method"`
	SoftLimitUSD       float64 `json:"soft_limit_usd"`
	HardLimitUSD       float64 `json:"hard_limit_usd"`
	SystemHardLimitUSD float64 `json:"system_hard_limit_usd"`
}

type OpenAIUsageDailyCost struct {
	Timestamp float64 `json:"timestamp"`
	LineItems []struct {
		Name string  `json:"name"`
		Cost float64 `json:"cost"`
	}
}

type OpenAIUsageResponse struct {
	Object string `json:"object"`
	//DailyCosts []OpenAIUsageDailyCost `json:"daily_costs"`
	TotalUsage float64 `json:"total_usage"` // unit: 0.01 dollar
}

func updateChannelBalance(channel *model.Channel) (float64, error) {
	baseURL := common.ChannelBaseURLs[channel.Type]
	switch channel.Type {
	case common.ChannelTypeOpenAI:
		if channel.BaseURL != "" {
			baseURL = channel.BaseURL
		}
	case common.ChannelTypeAzure:
		return 0, errors.New("尚未实现")
	case common.ChannelTypeCustom:
		baseURL = channel.BaseURL
	default:
		return 0, errors.New("尚未实现")
	}
	url := fmt.Sprintf("%s/v1/dashboard/billing/subscription", baseURL)

	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return 0, err
	}
	auth := fmt.Sprintf("Bearer %s", channel.Key)
	req.Header.Add("Authorization", auth)
	res, err := client.Do(req)
	if err != nil {
		return 0, err
	}
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return 0, err
	}
	err = res.Body.Close()
	if err != nil {
		return 0, err
	}
	subscription := OpenAISubscriptionResponse{}
	err = json.Unmarshal(body, &subscription)
	if err != nil {
		return 0, err
	}
	now := time.Now()
	startDate := fmt.Sprintf("%s-01", now.Format("2006-01"))
	endDate := now.Format("2006-01-02")
	if !subscription.HasPaymentMethod {
		startDate = now.AddDate(0, 0, -100).Format("2006-01-02")
	}
	url = fmt.Sprintf("%s/v1/dashboard/billing/usage?start_date=%s&end_date=%s", baseURL, startDate, endDate)
	req, err = http.NewRequest("GET", url, nil)
	if err != nil {
		return 0, err
	}
	req.Header.Add("Authorization", auth)
	res, err = client.Do(req)
	if err != nil {
		return 0, err
	}
	body, err = io.ReadAll(res.Body)
	if err != nil {
		return 0, err
	}
	err = res.Body.Close()
	if err != nil {
		return 0, err
	}
	usage := OpenAIUsageResponse{}
	err = json.Unmarshal(body, &usage)
	if err != nil {
		return 0, err
	}
	balance := subscription.HardLimitUSD - usage.TotalUsage/100
	channel.UpdateBalance(balance)
	return balance, nil
}

func UpdateChannelBalance(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	channel, err := model.GetChannelById(id, true)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	balance, err := updateChannelBalance(channel)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"balance": balance,
	})
	return
}

func updateAllChannelsBalance() error {
	channels, err := model.GetAllChannels(0, 0, true)
	if err != nil {
		return err
	}
	for _, channel := range channels {
		if channel.Status != common.ChannelStatusEnabled {
			continue
		}
		// TODO: support Azure
		if channel.Type != common.ChannelTypeOpenAI && channel.Type != common.ChannelTypeCustom {
			continue
		}
		balance, err := updateChannelBalance(channel)
		if err != nil {
			continue
		} else {
			// err is nil & balance <= 0 means quota is used up
			if balance <= 0 {
				disableChannel(channel.Id, channel.Name, "余额不足")
			}
		}
	}
	return nil
}

func UpdateAllChannelsBalance(c *gin.Context) {
	// TODO: make it async
	err := updateAllChannelsBalance()
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
	})
	return
}
