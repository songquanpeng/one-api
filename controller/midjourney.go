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
				log.Printf("UpdateMidjourneyTask: %v", err)
			}
		}()
		time.Sleep(time.Duration(15) * time.Second)
		tasks := model.GetAllUnFinishTasks()
		if len(tasks) != 0 {
			//log.Printf("UpdateMidjourneyTask: %v", time.Now())
			ids := make([]string, 0)
			for _, task := range tasks {
				ids = append(ids, task.MjId)
			}
			requestUrl := "http://107.173.171.147:8080/mj/task/list-by-condition"
			requestBody := map[string]interface{}{
				"ids": ids,
			}
			jsonStr, err := json.Marshal(requestBody)
			if err != nil {
				log.Printf("UpdateMidjourneyTask: %v", err)
				continue
			}
			req, err := http.NewRequest("POST", requestUrl, bytes.NewBuffer(jsonStr))
			if err != nil {
				log.Printf("UpdateMidjourneyTask: %v", err)
				continue
			}
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("mj-api-secret", "uhiftyuwadbkjshbiklahcuitguasguzhxliawodawdu")
			resp, err := httpClient.Do(req)
			if err != nil {
				log.Printf("UpdateMidjourneyTask: %v", err)
				continue
			}
			defer resp.Body.Close()
			var response []Midjourney
			err = json.NewDecoder(resp.Body).Decode(&response)
			if err != nil {
				log.Printf("UpdateMidjourneyTask: %v", err)
				continue
			}
			for _, responseItem := range response {
				var midjourneyTask *model.Midjourney
				for _, mj := range tasks {
					mj.MjId = responseItem.MjId
					midjourneyTask = model.GetMjByuId(mj.Id)
				}
				if midjourneyTask != nil {
					midjourneyTask.Code = 1
					midjourneyTask.Progress = responseItem.Progress
					midjourneyTask.PromptEn = responseItem.PromptEn
					midjourneyTask.State = responseItem.State
					midjourneyTask.SubmitTime = responseItem.SubmitTime
					midjourneyTask.StartTime = responseItem.StartTime
					midjourneyTask.FinishTime = responseItem.FinishTime
					midjourneyTask.ImageUrl = responseItem.ImageUrl
					midjourneyTask.Status = responseItem.Status
					midjourneyTask.FailReason = responseItem.FailReason
					if midjourneyTask.Progress != "100%" && responseItem.FailReason != "" {
						log.Println(midjourneyTask.MjId + " 构建失败，" + midjourneyTask.FailReason)
						midjourneyTask.Progress = "100%"
						err = model.CacheUpdateUserQuota(midjourneyTask.UserId)
						if err != nil {
							log.Println("error update user quota cache: " + err.Error())
						} else {
							modelRatio := common.GetModelRatio(imageModel)
							groupRatio := common.GetGroupRatio("default")
							ratio := modelRatio * groupRatio
							quota := int(ratio * 1 * 1000)
							if quota != 0 {
								err := model.IncreaseUserQuota(midjourneyTask.UserId, quota)
								if err != nil {
									log.Println("fail to increase user quota")
								}
								logContent := fmt.Sprintf("%s 构图失败，补偿 %s", midjourneyTask.MjId, common.LogQuota(quota))
								model.RecordLog(midjourneyTask.UserId, 1, logContent)
							}
						}
					}

					err = midjourneyTask.Update()
					if err != nil {
						log.Printf("UpdateMidjourneyTaskFail: %v", err)
					}
					log.Printf("UpdateMidjourneyTask: %v", midjourneyTask)
				}
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
