package router

import (
	"one-api/controller"
	"one-api/middleware"
	"one-api/relay"

	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
)

func SetApiRouter(router *gin.Engine) {
	apiRouter := router.Group("/api")
	apiRouter.Use(gzip.Gzip(gzip.DefaultCompression))
	apiRouter.POST("/telegram/:token", middleware.Telegram(), controller.TelegramBotWebHook)
	apiRouter.Use(middleware.GlobalAPIRateLimit())
	{
		apiRouter.GET("/status", controller.GetStatus)
		apiRouter.GET("/notice", controller.GetNotice)
		apiRouter.GET("/about", controller.GetAbout)
		apiRouter.GET("/prices", middleware.PricesAuth(), middleware.CORS(), controller.GetPricesList)
		apiRouter.GET("/ownedby", relay.GetModelOwnedBy)
		apiRouter.GET("/home_page_content", controller.GetHomePageContent)
		apiRouter.GET("/verification", middleware.CriticalRateLimit(), middleware.TurnstileCheck(), controller.SendEmailVerification)
		apiRouter.GET("/reset_password", middleware.CriticalRateLimit(), middleware.TurnstileCheck(), controller.SendPasswordResetEmail)
		apiRouter.POST("/user/reset", middleware.CriticalRateLimit(), controller.ResetPassword)
		apiRouter.GET("/oauth/github", middleware.CriticalRateLimit(), controller.GitHubOAuth)
		apiRouter.GET("/oauth/state", middleware.CriticalRateLimit(), controller.GenerateOAuthCode)
		apiRouter.GET("/oauth/wechat", middleware.CriticalRateLimit(), controller.WeChatAuth)
		apiRouter.GET("/oauth/wechat/bind", middleware.CriticalRateLimit(), middleware.UserAuth(), controller.WeChatBind)
		apiRouter.GET("/oauth/email/bind", middleware.CriticalRateLimit(), middleware.UserAuth(), controller.EmailBind)

		userRoute := apiRouter.Group("/user")
		{
			userRoute.POST("/register", middleware.CriticalRateLimit(), middleware.TurnstileCheck(), controller.Register)
			userRoute.POST("/login", middleware.CriticalRateLimit(), controller.Login)
			userRoute.GET("/logout", controller.Logout)

			selfRoute := userRoute.Group("/")
			selfRoute.Use(middleware.UserAuth())
			{
				selfRoute.GET("/dashboard", controller.GetUserDashboard)
				selfRoute.GET("/self", controller.GetSelf)
				selfRoute.PUT("/self", controller.UpdateSelf)
				// selfRoute.DELETE("/self", controller.DeleteSelf)
				selfRoute.GET("/token", controller.GenerateAccessToken)
				selfRoute.GET("/aff", controller.GetAffCode)
				selfRoute.POST("/topup", controller.TopUp)
				selfRoute.GET("/models", relay.ListModels)
			}

			adminRoute := userRoute.Group("/")
			adminRoute.Use(middleware.AdminAuth())
			{
				adminRoute.GET("/", controller.GetUsersList)
				adminRoute.GET("/:id", controller.GetUser)
				adminRoute.POST("/", controller.CreateUser)
				adminRoute.POST("/manage", controller.ManageUser)
				adminRoute.PUT("/", controller.UpdateUser)
				adminRoute.DELETE("/:id", controller.DeleteUser)
			}
		}
		optionRoute := apiRouter.Group("/option")
		optionRoute.Use(middleware.RootAuth())
		{
			optionRoute.GET("/", controller.GetOptions)
			optionRoute.PUT("/", controller.UpdateOption)
			optionRoute.GET("/telegram", controller.GetTelegramMenuList)
			optionRoute.POST("/telegram", controller.AddOrUpdateTelegramMenu)
			optionRoute.GET("/telegram/status", controller.GetTelegramBotStatus)
			optionRoute.PUT("/telegram/reload", controller.ReloadTelegramBot)
			optionRoute.GET("/telegram/:id", controller.GetTelegramMenu)
			optionRoute.DELETE("/telegram/:id", controller.DeleteTelegramMenu)
		}
		channelRoute := apiRouter.Group("/channel")
		channelRoute.Use(middleware.AdminAuth())
		{
			channelRoute.GET("/", controller.GetChannelsList)
			channelRoute.GET("/models", relay.ListModelsForAdmin)
			channelRoute.GET("/:id", controller.GetChannel)
			channelRoute.GET("/test", controller.TestAllChannels)
			channelRoute.GET("/test/:id", controller.TestChannel)
			channelRoute.GET("/update_balance", controller.UpdateAllChannelsBalance)
			channelRoute.GET("/update_balance/:id", controller.UpdateChannelBalance)
			channelRoute.POST("/", controller.AddChannel)
			channelRoute.PUT("/", controller.UpdateChannel)
			channelRoute.PUT("/batch/azure_api", controller.BatchUpdateChannelsAzureApi)
			channelRoute.PUT("/batch/del_model", controller.BatchDelModelChannels)
			channelRoute.DELETE("/disabled", controller.DeleteDisabledChannel)
			channelRoute.DELETE("/:id", controller.DeleteChannel)
		}
		tokenRoute := apiRouter.Group("/token")
		tokenRoute.Use(middleware.UserAuth())
		{
			tokenRoute.GET("/", controller.GetUserTokensList)
			tokenRoute.GET("/:id", controller.GetToken)
			tokenRoute.POST("/", controller.AddToken)
			tokenRoute.PUT("/", controller.UpdateToken)
			tokenRoute.DELETE("/:id", controller.DeleteToken)
		}
		redemptionRoute := apiRouter.Group("/redemption")
		redemptionRoute.Use(middleware.AdminAuth())
		{
			redemptionRoute.GET("/", controller.GetRedemptionsList)
			redemptionRoute.GET("/:id", controller.GetRedemption)
			redemptionRoute.POST("/", controller.AddRedemption)
			redemptionRoute.PUT("/", controller.UpdateRedemption)
			redemptionRoute.DELETE("/:id", controller.DeleteRedemption)
		}
		logRoute := apiRouter.Group("/log")
		logRoute.GET("/", middleware.AdminAuth(), controller.GetLogsList)
		logRoute.DELETE("/", middleware.AdminAuth(), controller.DeleteHistoryLogs)
		logRoute.GET("/stat", middleware.AdminAuth(), controller.GetLogsStat)
		logRoute.GET("/self/stat", middleware.UserAuth(), controller.GetLogsSelfStat)
		// logRoute.GET("/search", middleware.AdminAuth(), controller.SearchAllLogs)
		logRoute.GET("/self", middleware.UserAuth(), controller.GetUserLogsList)
		// logRoute.GET("/self/search", middleware.UserAuth(), controller.SearchUserLogs)
		groupRoute := apiRouter.Group("/group")
		groupRoute.Use(middleware.AdminAuth())
		{
			groupRoute.GET("/", controller.GetGroups)
		}

		analyticsRoute := apiRouter.Group("/analytics")
		analyticsRoute.Use(middleware.AdminAuth())
		{
			analyticsRoute.GET("/user_statistics", controller.GetUserStatistics)
			analyticsRoute.GET("/channel_statistics", controller.GetChannelStatistics)
			analyticsRoute.GET("/redemption_statistics", controller.GetRedemptionStatistics)
			analyticsRoute.GET("/users_period", controller.GetUserStatisticsByPeriod)
			analyticsRoute.GET("/channel_period", controller.GetChannelExpensesByPeriod)
			analyticsRoute.GET("/redemption_period", controller.GetRedemptionStatisticsByPeriod)
		}

		pricesRoute := apiRouter.Group("/prices")
		pricesRoute.Use(middleware.AdminAuth())
		{
			pricesRoute.GET("/model_list", controller.GetAllModelList)
			pricesRoute.POST("/single", controller.AddPrice)
			pricesRoute.PUT("/single/:model", controller.UpdatePrice)
			pricesRoute.DELETE("/single/:model", controller.DeletePrice)
			pricesRoute.POST("/multiple", controller.BatchSetPrices)
			pricesRoute.PUT("/multiple/delete", controller.BatchDeletePrices)
			pricesRoute.POST("/sync", controller.SyncPricing)

		}

	}

}
