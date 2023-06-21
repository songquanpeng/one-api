package controller

import (
	"github.com/gin-gonic/gin"
	"one-api/common"
	"one-api/model"
)

func GetSubscription(c *gin.Context) {
	userId := c.GetInt("id")
	quota, err := model.GetUserQuota(userId)
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
	subscription := OpenAISubscriptionResponse{
		Object:             "billing_subscription",
		HasPaymentMethod:   true,
		SoftLimitUSD:       amount,
		HardLimitUSD:       amount,
		SystemHardLimitUSD: amount,
	}
	c.JSON(200, subscription)
	return
}

func GetUsage(c *gin.Context) {
	userId := c.GetInt("id")
	quota, err := model.GetUserUsedQuota(userId)
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
		TotalUsage: amount,
	}
	c.JSON(200, usage)
	return
}
