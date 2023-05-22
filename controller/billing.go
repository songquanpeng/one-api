package controller

import (
	"github.com/gin-gonic/gin"
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
	subscription := OpenAISubscriptionResponse{
		Object:             "billing_subscription",
		HasPaymentMethod:   true,
		SoftLimitUSD:       float64(quota),
		HardLimitUSD:       float64(quota),
		SystemHardLimitUSD: float64(quota),
	}
	c.JSON(200, subscription)
	return
}

func GetUsage(c *gin.Context) {
	//userId := c.GetInt("id")
	// TODO: get usage from database
	usage := OpenAIUsageResponse{
		Object:     "list",
		TotalUsage: 0,
	}
	c.JSON(200, usage)
	return
}
