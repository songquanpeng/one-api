package controller

import (
	"net/http"
	"one-api/model"
	"strconv"

	"github.com/gin-gonic/gin"
)

func GetUserStatistics(c *gin.Context) {
	userStatistics, err := model.GetStatisticsUser()
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "无法获取用户统计信息.",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data":    userStatistics,
	})
}

func GetChannelStatistics(c *gin.Context) {
	channelStatistics, err := model.GetStatisticsChannel()
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "无法获取渠道统计信息.",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data":    channelStatistics,
	})
}

func GetRedemptionStatistics(c *gin.Context) {
	redemptionStatistics, err := model.GetStatisticsRedemption()
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "无法获取充值卡统计信息.",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data":    redemptionStatistics,
	})
}

func GetUserStatisticsByPeriod(c *gin.Context) {
	startTimestamp, _ := strconv.ParseInt(c.Query("start_timestamp"), 10, 64)
	endTimestamp, _ := strconv.ParseInt(c.Query("end_timestamp"), 10, 64)
	logStatistics, err := model.GetUserStatisticsByPeriod(startTimestamp, endTimestamp)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "无法获取用户区间统计信息.",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data":    logStatistics,
	})
}

func GetChannelExpensesByPeriod(c *gin.Context) {
	startTimestamp, _ := strconv.ParseInt(c.Query("start_timestamp"), 10, 64)
	endTimestamp, _ := strconv.ParseInt(c.Query("end_timestamp"), 10, 64)
	logStatistics, err := model.GetChannelExpensesByPeriod(startTimestamp, endTimestamp)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "无法获取渠道区间统计信息.",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data":    logStatistics,
	})
}

func GetRedemptionStatisticsByPeriod(c *gin.Context) {
	startTimestamp, _ := strconv.ParseInt(c.Query("start_timestamp"), 10, 64)
	endTimestamp, _ := strconv.ParseInt(c.Query("end_timestamp"), 10, 64)
	logStatistics, err := model.GetStatisticsRedemptionByPeriod(startTimestamp, endTimestamp)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "无法获取充值区间统计信息.",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data":    logStatistics,
	})
}
