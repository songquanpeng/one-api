package controller

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"one-api/common"
	"one-api/model"
	"strconv"
	"time"
)

func UpdateMidjourneyTask() {
	//revocer
	imageModel := "midjourney"
	for {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("UpdateMidjourneyTask panic: %v", err)
			}
		}()
		time.Sleep(time.Duration(15) * time.Second)
		tasks := model.GetAllUnFinishTasks()
		if len(tasks) != 0 {
			for _, task := range tasks {
				midjourneyChannel, err := model.GetChannelById(task.ChannelId, true)
				if err != nil {
					log.Printf("UpdateMidjourneyTask: %v", err)
					task.FailReason = fmt.Sprintf("获取渠道信息失败，请联系管理员，渠道ID：%d", task.ChannelId)
					task.Status = "FAILURE"
					task.Progress = "100%"
					err := task.Update()
					if err != nil {
						log.Printf("UpdateMidjourneyTask error: %v", err)
					}
					continue
				}
				requestUrl := fmt.Sprintf("%s/mj/task/%s/fetch", midjourneyChannel.BaseURL, task.MjId)

				req, err := http.NewRequest("GET", requestUrl, bytes.NewBuffer([]byte("")))
				if err != nil {
					log.Printf("UpdateMidjourneyTask error: %v", err)
					continue
				}

				req.Header.Set("Content-Type", "application/json")
				req.Header.Set("mj-api-secret", midjourneyChannel.Key)
				resp, err := httpClient.Do(req)
				if err != nil {
					log.Printf("UpdateMidjourneyTask error: %v", err)
					continue
				}
				defer resp.Body.Close()
				var responseItem Midjourney
				err = json.NewDecoder(resp.Body).Decode(&responseItem)
				if err != nil {
					log.Printf("UpdateMidjourneyTask error: %v", err)
					continue
				}
				task.Code = 1
				task.Progress = responseItem.Progress
				task.PromptEn = responseItem.PromptEn
				task.State = responseItem.State
				task.SubmitTime = responseItem.SubmitTime
				task.StartTime = responseItem.StartTime
				task.FinishTime = responseItem.FinishTime
				task.ImageUrl = responseItem.ImageUrl
				task.Status = responseItem.Status
				task.FailReason = responseItem.FailReason
				if task.Progress != "100%" && responseItem.FailReason != "" {
					log.Println(task.MjId + " 构建失败，" + task.FailReason)
					task.Progress = "100%"
					err = model.CacheUpdateUserQuota(task.UserId)
					if err != nil {
						log.Println("error update user quota cache: " + err.Error())
					} else {
						modelRatio := common.GetModelRatio(imageModel)
						groupRatio := common.GetGroupRatio("default")
						ratio := modelRatio * groupRatio
						quota := int(ratio * 1 * 1000)
						if quota != 0 {
							err := model.IncreaseUserQuota(task.UserId, quota)
							if err != nil {
								log.Println("fail to increase user quota")
							}
							logContent := fmt.Sprintf("%s 构图失败，补偿 %s", task.MjId, common.LogQuota(quota))
							model.RecordLog(task.UserId, 1, logContent)
						}
					}
				}

				err = task.Update()
				if err != nil {
					log.Printf("UpdateMidjourneyTask error: %v", err)
				}
				log.Printf("UpdateMidjourneyTask success: %v", task)
			}
		}
	}
}

func GetAllMidjourney(c *gin.Context) {
	p, _ := strconv.Atoi(c.Query("p"))
	if p < 0 {
		p = 0
	}
	logs := model.GetAllTasks(p*common.ItemsPerPage, common.ItemsPerPage)
	if logs == nil {
		logs = make([]*model.Midjourney, 0)
	}
	c.JSON(200, gin.H{
		"success": true,
		"message": "",
		"data":    logs,
	})
}

func GetUserMidjourney(c *gin.Context) {
	p, _ := strconv.Atoi(c.Query("p"))
	if p < 0 {
		p = 0
	}
	userId := c.GetInt("id")
	log.Printf("userId = %d \n", userId)
	logs := model.GetAllUserTask(userId, p*common.ItemsPerPage, common.ItemsPerPage)
	if logs == nil {
		logs = make([]*model.Midjourney, 0)
	}
	c.JSON(200, gin.H{
		"success": true,
		"message": "",
		"data":    logs,
	})
}
