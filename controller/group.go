package controller

import (
	"github.com/gin-gonic/gin"
	"github.com/songquanpeng/one-api/common"
	"net/http"
)

func GetGroups(c *gin.Context) {
	groupNames := make([]string, 0)
	for groupName := range common.GroupRatio {
		groupNames = append(groupNames, groupName)
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data":    groupNames,
	})
}
