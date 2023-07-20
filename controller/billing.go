package controller

import (
	"one-api/common"
	"one-api/model"

	"github.com/gin-gonic/gin"
)

func GetSubscription(c *gin.Context) {
	var quota int
	var err error
	var expirationDate int64

	tokenId := c.GetInt("token_id")
	token, err := model.GetTokenById(tokenId)

	expirationDate = token.ExpiredTime

	if common.DisplayTokenStatEnabled {
		quota = token.RemainQuota
	} else {
		userId := c.GetInt("id")
		quota, err = model.GetUserQuota(userId)
	}
	if err != nil {
		openAIError := OpenAIError{
			Message: err.Error(),
			Type:    "one_api_error",
		}
		c.JSON(200, gin.H{
			"error": openAIError,
		})
		return
	}
	amount := float64(quota)
	if common.DisplayInCurrencyEnabled {
		amount /= common.QuotaPerUnit
	}
	if token != nil && token.UnlimitedQuota {
		amount = 100000000
	}
	subscription := OpenAISubscriptionResponse{
		Object:             "billing_subscription",
		HasPaymentMethod:   true,
		SoftLimitUSD:       amount,
		HardLimitUSD:       amount,
		SystemHardLimitUSD: amount,
		AccessUntil:        expirationDate,
	}
	c.JSON(200, subscription)
	return
}

func GetUsage(c *gin.Context) {
	var quota int
	var err error
	var token *model.Token
	if common.DisplayTokenStatEnabled {
		tokenId := c.GetInt("token_id")
		token, err = model.GetTokenById(tokenId)
		quota = token.UsedQuota
	} else {
		userId := c.GetInt("id")
		quota, err = model.GetUserUsedQuota(userId)
	}
	if err != nil {
		openAIError := OpenAIError{
			Message: err.Error(),
			Type:    "one_api_error",
		}
		c.JSON(200, gin.H{
			"error": openAIError,
		})
		return
	}
	amount := float64(quota)
	if common.DisplayInCurrencyEnabled {
		amount /= common.QuotaPerUnit
	}
	usage := OpenAIUsageResponse{
		Object:     "list",
		TotalUsage: amount * 100,
	}
	c.JSON(200, usage)
	return
}
