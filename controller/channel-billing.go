package controller

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"one-api/common/config"
	"one-api/model"
	"one-api/providers"
	providersBase "one-api/providers/base"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// https://github.com/MartialBE/one-api/issues/79

type OpenAISubscriptionResponse struct {
	Object             string  `json:"object"`
	HasPaymentMethod   bool    `json:"has_payment_method"`
	SoftLimitUSD       float64 `json:"soft_limit_usd"`
	HardLimitUSD       float64 `json:"hard_limit_usd"`
	SystemHardLimitUSD float64 `json:"system_hard_limit_usd"`
	AccessUntil        int64   `json:"access_until"`
}

type OpenAIUsageDailyCost struct {
	Timestamp float64 `json:"timestamp"`
	LineItems []struct {
		Name string  `json:"name"`
		Cost float64 `json:"cost"`
	}
}

type OpenAICreditGrants struct {
	Object         string  `json:"object"`
	TotalGranted   float64 `json:"total_granted"`
	TotalUsed      float64 `json:"total_used"`
	TotalAvailable float64 `json:"total_available"`
}

type OpenAIUsageResponse struct {
	Object string `json:"object"`
	//DailyCosts []OpenAIUsageDailyCost `json:"daily_costs"`
	TotalUsage float64 `json:"total_usage"` // unit: 0.01 dollar
}

func updateChannelBalance(channel *model.Channel) (float64, error) {
	req, err := http.NewRequest("POST", "/balance", nil)
	if err != nil {
		return 0, err
	}
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	req.Header.Set("Content-Type", "application/json")

	provider := providers.GetProvider(channel, c)
	if provider == nil {
		return 0, errors.New("provider not found")
	}

	balanceProvider, ok := provider.(providersBase.BalanceInterface)
	if !ok {
		return 0, errors.New("provider not implemented")
	}

	return balanceProvider.Balance()

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
	channel, err := model.GetChannelById(id)
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
}

func updateAllChannelsBalance() error {
	channels, err := model.GetAllChannels()
	if err != nil {
		return err
	}
	for _, channel := range channels {
		if channel.Status != config.ChannelStatusEnabled {
			continue
		}
		// TODO: support Azure
		if channel.Type != config.ChannelTypeOpenAI && channel.Type != config.ChannelTypeCustom {
			continue
		}
		balance, err := updateChannelBalance(channel)
		if err != nil {
			continue
		} else {
			// err is nil & balance <= 0 means quota is used up
			if balance <= 0 {
				DisableChannel(channel.Id, channel.Name, "余额不足", true)
			}
		}
		time.Sleep(config.RequestInterval)
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
}

// func AutomaticallyUpdateChannels(frequency int) {
// 	if frequency <= 0 {
// 		return
// 	}

// 	for {
// 		time.Sleep(time.Duration(frequency) * time.Minute)
// 		common.SysLog("updating all channels")
// 		_ = updateAllChannelsBalance()
// 		common.SysLog("channels update done")
// 	}
// }
