package main

import (
	"embed"
	"fmt"
	"one-api/common"
	"one-api/common/config"
	"one-api/common/notify"
	"one-api/common/requester"
	"one-api/common/storage"
	"one-api/common/telegram"
	"one-api/controller"
	"one-api/cron"
	"one-api/middleware"
	"one-api/model"
	"one-api/relay/util"
	"one-api/router"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

//go:embed web/build
var buildFS embed.FS

//go:embed web/build/index.html
var indexPage []byte

func main() {
	config.InitConf()
	common.SetupLogger()
	common.SysLog("One API " + common.Version + " started")
	// Initialize SQL Database
	model.SetupDB()
	defer model.CloseDB()
	// Initialize Redis
	common.InitRedisClient()
	// Initialize options
	model.InitOptionMap()
	util.NewPricing()
	initMemoryCache()
	initSync()

	common.InitTokenEncoders()
	requester.InitHttpClient()
	// Initialize Telegram bot
	telegram.InitTelegramBot()

	controller.InitMidjourneyTask()
	notify.InitNotifier()
	cron.InitCron()
	storage.InitStorage()

	initHttpServer()
}

func initMemoryCache() {
	if viper.GetBool("memory_cache_enabled") {
		common.MemoryCacheEnabled = true
	}

	if !common.MemoryCacheEnabled {
		return
	}

	syncFrequency := viper.GetInt("sync_frequency")
	model.TokenCacheSeconds = syncFrequency

	common.SysLog("memory cache enabled")
	common.SysError(fmt.Sprintf("sync frequency: %d seconds", syncFrequency))
	go model.SyncOptions(syncFrequency)
	go SyncChannelCache(syncFrequency)
}

func initSync() {
	// go controller.AutomaticallyUpdateChannels(viper.GetInt("channel.update_frequency"))
	go controller.AutomaticallyTestChannels(viper.GetInt("channel.test_frequency"))
}

func initHttpServer() {
	if viper.GetString("gin_mode") != "debug" {
		gin.SetMode(gin.ReleaseMode)
	}

	server := gin.New()
	server.Use(gin.Recovery())
	server.Use(middleware.RequestId())
	middleware.SetUpLogger(server)

	store := cookie.NewStore([]byte(common.SessionSecret))
	server.Use(sessions.Sessions("session", store))

	router.SetRouter(server, buildFS, indexPage)
	port := viper.GetString("port")

	err := server.Run(":" + port)
	if err != nil {
		common.FatalLog("failed to start HTTP server: " + err.Error())
	}
}

func SyncChannelCache(frequency int) {
	// 只有 从 服务器端获取数据的时候才会用到
	if common.IsMasterNode {
		common.SysLog("master node does't synchronize the channel")
		return
	}
	for {
		time.Sleep(time.Duration(frequency) * time.Second)
		common.SysLog("syncing channels from database")
		model.ChannelGroup.Load()
		util.PricingInstance.Init()
	}
}
