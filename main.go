package main

import (
	"embed"
	"one-api/common"
	"one-api/controller"
	"one-api/middleware"
	"one-api/model"
	"one-api/router"
	"os"
	"strconv"

	"github.com/getsentry/sentry-go"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

//go:embed web/build
var buildFS embed.FS

//go:embed web/build/index.html
var indexPage []byte

func main() {
	// This will load .env file if it exists, to set environment variables instead of exporting them one by one
	envErr := godotenv.Load()
	if envErr != nil {
		common.SysLog("Cannot load .env file, using environment variables, this is not an error, just a reminder")
	}

	// Sentry is a cross-platform crash reporting and aggregation platform.
	// It provides the ability to capture, index and store exceptions generated
	// This will only activate when SENTRY_DSN is set, if you worry about privacy, you can set it to an empty string
	sentrDSN := os.Getenv("SENTRY_DSN")
	if sentrDSN != "" {
		sentry.Init(sentry.ClientOptions{
			Dsn:              sentrDSN,
			TracesSampleRate: 0.1,
		})
		common.SysLog("Sentry initialized")
	}

	common.SetupLogger()
	common.SysLog("One API " + common.Version + " started")
	if os.Getenv("GIN_MODE") != "debug" {
		gin.SetMode(gin.ReleaseMode)
	}
	if common.DebugEnabled {
		common.SysLog("running in debug mode")
	}
	// Initialize SQL Database
	err := model.InitDB()
	if err != nil {
		common.FatalLog("failed to initialize database: " + err.Error())
	}
	defer func() {
		err := model.CloseDB()
		if err != nil {
			common.FatalLog("failed to close database: " + err.Error())
		}
	}()

	// Initialize Redis
	err = common.InitRedisClient()
	if err != nil {
		common.FatalLog("failed to initialize Redis: " + err.Error())
	}

	// Initialize options
	model.InitOptionMap()
	if common.RedisEnabled {
		model.InitChannelCache()
	}
	if os.Getenv("SYNC_FREQUENCY") != "" {
		frequency, err := strconv.Atoi(os.Getenv("SYNC_FREQUENCY"))
		if err != nil {
			common.FatalLog("failed to parse SYNC_FREQUENCY: " + err.Error())
		}
		common.SyncFrequency = frequency
		go model.SyncOptions(frequency)
		if common.RedisEnabled {
			go model.SyncChannelCache(frequency)
		}
	}
	if os.Getenv("CHANNEL_UPDATE_FREQUENCY") != "" {
		frequency, err := strconv.Atoi(os.Getenv("CHANNEL_UPDATE_FREQUENCY"))
		if err != nil {
			common.FatalLog("failed to parse CHANNEL_UPDATE_FREQUENCY: " + err.Error())
		}
		go controller.AutomaticallyUpdateChannels(frequency)
	}
	if os.Getenv("CHANNEL_TEST_FREQUENCY") != "" {
		frequency, err := strconv.Atoi(os.Getenv("CHANNEL_TEST_FREQUENCY"))
		if err != nil {
			common.FatalLog("failed to parse CHANNEL_TEST_FREQUENCY: " + err.Error())
		}
		go controller.AutomaticallyTestChannels(frequency)
	}
	if os.Getenv("BATCH_UPDATE_ENABLED") == "true" {
		common.BatchUpdateEnabled = true
		common.SysLog("batch update enabled with interval " + strconv.Itoa(common.BatchUpdateInterval) + "s")
		model.InitBatchUpdater()
	}
	if os.Getenv("TOKEN_ENCODER_STARTUP_INIT_DISABLED") != "true" {
		controller.InitTokenEncoders()
	}

	// Initialize HTTP server
	server := gin.New()
	server.Use(gin.Recovery())
	// This will cause SSE not to work!!!
	//server.Use(gzip.Gzip(gzip.DefaultCompression))
	server.Use(middleware.RequestId())
	middleware.SetUpLogger(server)
	// Initialize session store
	store := cookie.NewStore([]byte(common.SessionSecret))
	server.Use(sessions.Sessions("session", store))

	router.SetRouter(server, buildFS, indexPage)
	var port = os.Getenv("PORT")
	if port == "" {
		port = strconv.Itoa(*common.Port)
	}
	err = server.Run(":" + port)
	if err != nil {
		common.FatalLog("failed to start HTTP server: " + err.Error())
	}
}
