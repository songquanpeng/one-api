package controller

import (
	"errors"
	"net/http"
	"one-api/common"
	"one-api/model"
	"one-api/relay/util"

	"github.com/gin-gonic/gin"
)

func GetPricesList(c *gin.Context) {
	pricesType := c.DefaultQuery("type", "db")

	prices := util.GetPricesList(pricesType)

	if len(prices) == 0 {
		common.APIRespondWithError(c, http.StatusOK, errors.New("pricing data not found"))
		return
	}

	if pricesType == "old" {
		c.JSON(http.StatusOK, prices)
	} else {
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": "",
			"data":    prices,
		})
	}
}

func GetAllModelList(c *gin.Context) {
	prices := util.PricingInstance.GetAllPrices()
	channelModel := model.ChannelGroup.Rule

	modelsMap := make(map[string]bool)
	for modelName := range prices {
		modelsMap[modelName] = true
	}

	for _, modelMap := range channelModel {
		for modelName := range modelMap {
			if _, ok := prices[modelName]; !ok {
				modelsMap[modelName] = true
			}
		}
	}

	var models []string
	for modelName := range modelsMap {
		models = append(models, modelName)
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data":    models,
	})
}

func AddPrice(c *gin.Context) {
	var price model.Price
	if err := c.ShouldBindJSON(&price); err != nil {
		common.APIRespondWithError(c, http.StatusOK, err)
		return
	}

	if err := util.PricingInstance.AddPrice(&price); err != nil {
		common.APIRespondWithError(c, http.StatusOK, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
	})
}

func UpdatePrice(c *gin.Context) {
	modelName := c.Param("model")
	if modelName == "" {
		common.APIRespondWithError(c, http.StatusOK, errors.New("model name is required"))
		return
	}

	var price model.Price
	if err := c.ShouldBindJSON(&price); err != nil {
		common.APIRespondWithError(c, http.StatusOK, err)
		return
	}

	if err := util.PricingInstance.UpdatePrice(modelName, &price); err != nil {
		common.APIRespondWithError(c, http.StatusOK, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
	})
}

func DeletePrice(c *gin.Context) {
	modelName := c.Param("model")
	if modelName == "" {
		common.APIRespondWithError(c, http.StatusOK, errors.New("model name is required"))
		return
	}

	if err := util.PricingInstance.DeletePrice(modelName); err != nil {
		common.APIRespondWithError(c, http.StatusOK, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
	})
}

type PriceBatchRequest struct {
	OriginalModels []string `json:"original_models"`
	util.BatchPrices
}

func BatchSetPrices(c *gin.Context) {
	pricesBatch := &PriceBatchRequest{}
	if err := c.ShouldBindJSON(pricesBatch); err != nil {
		common.APIRespondWithError(c, http.StatusOK, err)
		return
	}

	if err := util.PricingInstance.BatchSetPrices(&pricesBatch.BatchPrices, pricesBatch.OriginalModels); err != nil {
		common.APIRespondWithError(c, http.StatusOK, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
	})
}

type PriceBatchDeleteRequest struct {
	Models []string `json:"models" binding:"required"`
}

func BatchDeletePrices(c *gin.Context) {
	pricesBatch := &PriceBatchDeleteRequest{}
	if err := c.ShouldBindJSON(pricesBatch); err != nil {
		common.APIRespondWithError(c, http.StatusOK, err)
		return
	}

	if err := util.PricingInstance.BatchDeletePrices(pricesBatch.Models); err != nil {
		common.APIRespondWithError(c, http.StatusOK, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
	})
}

func SyncPricing(c *gin.Context) {
	overwrite := c.DefaultQuery("overwrite", "false")

	prices := make([]*model.Price, 0)
	if err := c.ShouldBindJSON(&prices); err != nil {
		common.APIRespondWithError(c, http.StatusOK, err)
		return
	}

	if len(prices) == 0 {
		common.APIRespondWithError(c, http.StatusOK, errors.New("prices is required"))
		return
	}

	err := util.PricingInstance.SyncPricing(prices, overwrite == "true")
	if err != nil {
		common.APIRespondWithError(c, http.StatusOK, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
	})
}
