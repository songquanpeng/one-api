package controller

import (
	"github.com/gin-gonic/gin"
	"one-api/common"
	"one-api/model"
	"strconv"
	"strings"
)

func GetAllChannels(c *gin.Context) {
	p, _ := strconv.Atoi(c.Query("p"))
	if p < 0 {
		p = 0
	}
	channels, err := model.GetAllChannels(p*common.ItemsPerPage, common.ItemsPerPage, false)
	if err != nil {
		Err(c, err)
		return
	}
	Success(c, channels)
}

func SearchChannels(c *gin.Context) {
	keyword := c.Query("keyword")
	channels, err := model.SearchChannels(keyword)
	if err != nil {
		Err(c, err)
		return
	}
	Success(c, channels)
}

func GetChannel(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		Err(c, err)
		return
	}
	channel, err := model.GetChannelById(id, false)
	if err != nil {
		Err(c, err)
		return
	}
	Success(c, channel)
}

func AddChannel(c *gin.Context) {
	channel := model.Channel{}
	err := c.ShouldBindJSON(&channel)
	if err != nil {
		Err(c, err)
		return
	}
	channel.CreatedTime = common.GetTimestamp()
	keys := strings.Split(channel.Key, "\n")
	channels := make([]model.Channel, 0, len(keys))
	for _, key := range keys {
		if key == "" {
			continue
		}
		localChannel := channel
		localChannel.Key = key
		channels = append(channels, localChannel)
	}
	err = model.BatchInsertChannels(channels)
	if err != nil {
		Err(c, err)
		return
	}
	Success(c, channel)
}

func DeleteChannel(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	channel := model.Channel{Id: id}
	err := channel.Delete()
	if err != nil {
		Err(c, err)
		return
	}
	Success(c, channel)
}

func DeleteDisabledChannel(c *gin.Context) {
	rows, err := model.DeleteDisabledChannel()
	if err != nil {
		Err(c, err)
		return
	}
	Success(c, rows)
}

func UpdateChannel(c *gin.Context) {
	channel := model.Channel{}
	err := c.ShouldBindJSON(&channel)
	if err != nil {
		Err(c, err)
		return
	}
	err = channel.Update()
	if err != nil {
		Err(c, err)
		return
	}
	Success(c, channel)
}
