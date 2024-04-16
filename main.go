package main

import (
	"embed"
	"fmt"
	"one-api/common"
	"one-api/common/config"
	"one-api/common/notify"
	"one-api/common/requester"
	"one-api/common/telegram"
	"one-api/controller"
	"one-api/cron"
	"one-api/middleware"
	"one-api/model"
	"one-api/relay/util"
	"one-api/router"

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

	initHttpServer()
}

func initMemoryCache() {
	if viper.GetBool("MEMORY_CACHE_ENABLED") {
		common.MemoryCacheEnabled = true
	}

	if !common.MemoryCacheEnabled {
		return
	}

	syncFrequency := viper.GetInt("SYNC_FREQUENCY")
	model.TokenCacheSeconds = syncFrequency

	common.SysLog("memory cache enabled")
	common.SysError(fmt.Sprintf("sync frequency: %d seconds", syncFrequency))
	go model.SyncOptions(syncFrequency)
}

func initSync() {
	go controller.AutomaticallyUpdateChannels(viper.GetInt("CHANNEL_UPDATE_FREQUENCY"))
	go controller.AutomaticallyTestChannels(viper.GetInt("CHANNEL_TEST_FREQUENCY"))
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
	port := viper.GetString("PORT")

	err := server.Run(":" + port)
	if err != nil {
		common.FatalLog("failed to start HTTP server: " + err.Error())
	}
}
