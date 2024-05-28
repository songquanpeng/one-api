package cron

import (
	"one-api/common/logger"
	"one-api/model"

	"github.com/go-co-op/gocron/v2"
)

func InitCron() {
	scheduler, err := gocron.NewScheduler()
	if err != nil {
		logger.SysLog("Cron scheduler error: " + err.Error())
		return
	}

	// 添加删除cache的任务
	_, err = scheduler.NewJob(
		gocron.DailyJob(
			1,
			gocron.NewAtTimes(
				gocron.NewAtTime(0, 5, 0),
			)),
		gocron.NewTask(func() {
			model.RemoveChatCache()
			logger.SysLog("删除过期缓存数据")
		}),
	)

	if err != nil {
		logger.SysLog("Cron job error: " + err.Error())
		return
	}

	scheduler.Start()
}
