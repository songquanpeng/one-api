package router

import (
	"one-api/controller"
	"one-api/middleware"

	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
)

// func SetApiRouter(router *gin.Engine) {
// 	apiRouter := router.Group("/api")
// 	apiRouter.Use(gzip.Gzip(gzip.DefaultCompression))
// 	apiRouter.Use(middleware.GlobalAPIRateLimit())
// 	{
// 		apiRouter.GET("/status", controller.GetStatus)
// 		apiRouter.GET("/notice", controller.GetNotice)
// 		apiRouter.GET("/about", controller.GetAbout)
// 		apiRouter.GET("/home_page_content", controller.GetHomePageContent)
// 		apiRouter.GET("/verification", middleware.CriticalRateLimit(), middleware.TurnstileCheck(), controller.SendEmailVerification)
// 		apiRouter.GET("/reset_password", middleware.CriticalRateLimit(), middleware.TurnstileCheck(), controller.SendPasswordResetEmail)
// 		apiRouter.POST("/user/reset", middleware.CriticalRateLimit(), controller.ResetPassword)
// 		apiRouter.GET("/oauth/github", middleware.CriticalRateLimit(), controller.GitHubOAuth)
// 		apiRouter.GET("/oauth/state", middleware.CriticalRateLimit(), controller.GenerateOAuthCode)
// 		apiRouter.GET("/oauth/wechat", middleware.CriticalRateLimit(), controller.WeChatAuth)
// 		apiRouter.GET("/oauth/wechat/bind", middleware.CriticalRateLimit(), middleware.UserAuth(), controller.WeChatBind)
// 		apiRouter.GET("/oauth/email/bind", middleware.CriticalRateLimit(), middleware.UserAuth(), controller.EmailBind)

// 		userRoute := apiRouter.Group("/user")
// 		{
// 			userRoute.POST("/register", middleware.CriticalRateLimit(), middleware.TurnstileCheck(), controller.Register)
// 			userRoute.POST("/login", middleware.CriticalRateLimit(), controller.Login)
// 			userRoute.GET("/logout", controller.Logout)

// 			selfRoute := userRoute.Group("/")
// 			selfRoute.Use(middleware.UserAuth())
// 			{
// 				selfRoute.GET("/self", controller.GetSelf)
// 				selfRoute.PUT("/self", controller.UpdateSelf)
// 				selfRoute.DELETE("/self", controller.DeleteSelf)
// 				selfRoute.GET("/token", controller.GenerateAccessToken)
// 				selfRoute.GET("/aff", controller.GetAffCode)
// 				selfRoute.POST("/topup", controller.TopUp)
// 			}

// 			adminRoute := userRoute.Group("/")
// 			adminRoute.Use(middleware.AdminAuth())
// 			{
// 				adminRoute.GET("/", controller.GetAllUsers)
// 				adminRoute.GET("/search", controller.SearchUsers)
// 				adminRoute.GET("/:id", controller.GetUser)
// 				adminRoute.POST("/", controller.CreateUser)
// 				adminRoute.POST("/manage", controller.ManageUser)
// 				adminRoute.PUT("/", controller.UpdateUser)
// 				adminRoute.DELETE("/:id", controller.DeleteUser)
// 			}
// 		}
// 		optionRoute := apiRouter.Group("/option")
// 		optionRoute.Use(middleware.RootAuth())
// 		{
// 			optionRoute.GET("/", controller.GetOptions)
// 			optionRoute.PUT("/", controller.UpdateOption)
// 		}
// 		channelRoute := apiRouter.Group("/channel")
// 		channelRoute.Use(middleware.AdminAuth())
// 		{
// 			channelRoute.GET("/", controller.GetAllChannels)
// 			channelRoute.GET("/search", controller.SearchChannels)
// 			channelRoute.GET("/models", controller.ListModels)
// 			channelRoute.GET("/:id", controller.GetChannel)
// 			channelRoute.GET("/test", controller.TestAllChannels)
// 			channelRoute.GET("/test/:id", controller.TestChannel)
// 			channelRoute.GET("/update_balance", controller.UpdateAllChannelsBalance)
// 			channelRoute.GET("/update_balance/:id", controller.UpdateChannelBalance)
// 			channelRoute.POST("/", controller.AddChannel)
// 			channelRoute.PUT("/", controller.UpdateChannel)
// 			channelRoute.DELETE("/disabled", controller.DeleteDisabledChannel)
// 			channelRoute.DELETE("/:id", controller.DeleteChannel)
// 		}
// 		tokenRoute := apiRouter.Group("/token")
// 		tokenRoute.Use(middleware.UserAuth())
// 		{
// 			tokenRoute.GET("/", controller.GetAllTokens)
// 			tokenRoute.GET("/search", controller.SearchTokens)
// 			tokenRoute.GET("/:id", controller.GetToken)
// 			tokenRoute.POST("/", controller.AddToken)
// 			tokenRoute.PUT("/", controller.UpdateToken)
// 			tokenRoute.DELETE("/:id", controller.DeleteToken)
// 		}
// 		redemptionRoute := apiRouter.Group("/redemption")
// 		redemptionRoute.Use(middleware.AdminAuth())
// 		{
// 			redemptionRoute.GET("/", controller.GetAllRedemptions)
// 			redemptionRoute.GET("/search", controller.SearchRedemptions)
// 			redemptionRoute.GET("/:id", controller.GetRedemption)
// 			redemptionRoute.POST("/", controller.AddRedemption)
// 			redemptionRoute.PUT("/", controller.UpdateRedemption)
// 			redemptionRoute.DELETE("/:id", controller.DeleteRedemption)
// 		}
// 		logRoute := apiRouter.Group("/log")
// 		logRoute.GET("/", middleware.AdminAuth(), controller.GetAllLogs)
// 		logRoute.DELETE("/", middleware.AdminAuth(), controller.DeleteHistoryLogs)
// 		logRoute.GET("/stat", middleware.AdminAuth(), controller.GetLogsStat)
// 		logRoute.GET("/self/stat", middleware.UserAuth(), controller.GetLogsSelfStat)
// 		logRoute.GET("/search", middleware.AdminAuth(), controller.SearchAllLogs)
// 		logRoute.GET("/self", middleware.UserAuth(), controller.GetUserLogs)
// 		logRoute.GET("/self/search", middleware.UserAuth(), controller.SearchUserLogs)
// 		groupRoute := apiRouter.Group("/group")
// 		groupRoute.Use(middleware.AdminAuth())
// 		{
// 			groupRoute.GET("/", controller.GetGroups)
// 		}
// 	}
// }

// 设置API路由
func SetApiRouter(router *gin.Engine) {
    // 创建API路由分组
    apiRouter := router.Group("/api")
    // 使用gzip压缩中间件
    apiRouter.Use(gzip.Gzip(gzip.DefaultCompression))
    // 使用全局API速率限制中间件
    apiRouter.Use(middleware.GlobalAPIRateLimit())
    {
   	 // 设置GET请求的路由处理函数
   	 apiRouter.GET("/status", controller.GetStatus)
   	 apiRouter.GET("/notice", controller.GetNotice)
   	 apiRouter.GET("/about", controller.GetAbout)
   	 apiRouter.GET("/home_page_content", controller.GetHomePageContent)
   	 apiRouter.GET("/verification", middleware.CriticalRateLimit(), middleware.TurnstileCheck(), controller.SendEmailVerification)
   	 apiRouter.GET("/reset_password", middleware.CriticalRateLimit(), middleware.TurnstileCheck(), controller.SendPasswordResetEmail)
   	 apiRouter.POST("/user/reset", middleware.CriticalRateLimit(), controller.ResetPassword)
   	 apiRouter.GET("/oauth/github", middleware.CriticalRateLimit(), controller.GitHubOAuth)
   	 apiRouter.GET("/oauth/state", middleware.CriticalRateLimit(), controller.GenerateOAuthCode)
   	 apiRouter.GET("/oauth/wechat", middleware.CriticalRateLimit(), controller.WeChatAuth)
   	 apiRouter.GET("/oauth/wechat/bind", middleware.CriticalRateLimit(), middleware.UserAuth(), controller.WeChatBind)
   	 apiRouter.GET("/oauth/email/bind", middleware.CriticalRateLimit(), middleware.UserAuth(), controller.EmailBind)
	 // 添加新的路由
	 

   	 // 创建user路由分组
   	 userRoute := apiRouter.Group("/user")
   	 {
   		 // 设置POST请求的路由处理函数
   		 userRoute.POST("/register", middleware.CriticalRateLimit(), middleware.TurnstileCheck(), controller.Register)
   		 userRoute.POST("/login", middleware.CriticalRateLimit(), controller.Login)
   		 userRoute.GET("/logout", controller.Logout)

   		 // 创建self路由分组，并使用用户认证中间件
   		 selfRoute := userRoute.Group("/")
   		 selfRoute.Use(middleware.UserAuth())
   		 {
   			 // 设置GET请求的路由处理函数
   			 selfRoute.GET("/self", controller.GetSelf)
   			 selfRoute.PUT("/self", controller.UpdateSelf)
   			 selfRoute.DELETE("/self", controller.DeleteSelf)
   			 selfRoute.GET("/token", controller.GenerateAccessToken)
   			 selfRoute.GET("/aff", controller.GetAffCode)
   			 selfRoute.POST("/topup", controller.TopUp)
   		 }

   		 // 创建admin路由分组，并使用管理员认证中间件
   		 adminRoute := userRoute.Group("/")
   		 adminRoute.Use(middleware.AdminAuth())
   		 {
   			 // 设置GET请求的路由处理函数
   			 adminRoute.GET("/", controller.GetAllUsers)
   			 adminRoute.GET("/search", controller.SearchUsers)
   			 adminRoute.GET("/:id", controller.GetUser)
   			 adminRoute.POST("/", controller.CreateUser)
   			 adminRoute.POST("/manage", controller.ManageUser)
   			 adminRoute.PUT("/", controller.UpdateUser)
   			 adminRoute.DELETE("/:id", controller.DeleteUser)
   		 }
   	 }

   	 // 创建option路由分组，并使用根认证中间件
   	 optionRoute := apiRouter.Group("/option")
   	 optionRoute.Use(middleware.RootAuth())
   	 {
   		 // 设置GET请求的路由处理函数
   		 optionRoute.GET("/", controller.GetOptions)
   		 optionRoute.PUT("/", controller.UpdateOption)
   	 }

   	 // 创建channel路由分组，并使用管理员认证中间件
   	 channelRoute := apiRouter.Group("/channel")
   	 channelRoute.Use(middleware.AdminAuth())
   	 {
   		 // 设置GET请求的路由处理函数
   		 channelRoute.GET("/", controller.GetAllChannels)
   		 channelRoute.GET("/search", controller.SearchChannels)
   		 channelRoute.GET("/:id", controller.GetChannel)
   		 channelRoute.GET("/test", controller.TestAllChannels)
   		 channelRoute.GET("/test/:id", controller.TestChannel)
   		 channelRoute.GET("/update_balance", controller.UpdateAllChannelsBalance)
   		 channelRoute.GET("/update_balance/:id", controller.UpdateChannelBalance)
   		 channelRoute.POST("/", controller.AddChannel)
   		 channelRoute.PUT("/", controller.UpdateChannel)
   		 channelRoute.DELETE("/disabled", controller.DeleteDisabledChannel)
   		 channelRoute.DELETE("/:id", controller.DeleteChannel)
   	 }

   	 // 创建token路由分组，并使用用户认证中间件
   	 tokenRoute := apiRouter.Group("/token")
   	 tokenRoute.Use(middleware.UserAuth())
   	 {
   		 // 设置GET请求的路由处理函数
   		 tokenRoute.GET("/", controller.GetAllTokens)
   		 tokenRoute.GET("/search", controller.SearchTokens)
   		 tokenRoute.GET("/:id", controller.GetToken)
   		 tokenRoute.POST("/", controller.AddToken)
   		 tokenRoute.PUT("/", controller.UpdateToken)
   		 tokenRoute.DELETE("/:id", controller.DeleteToken)
   	 }

   	 // 创建redemption路由分组，并使用管理员认证中间件
   	 redemptionRoute := apiRouter.Group("/redemption")
   	 redemptionRoute.Use(middleware.AdminAuth())
   	 {
   		 // 设置GET请求的路由处理函数
   		 redemptionRoute.GET("/", controller.GetAllRedemptions)
   		 redemptionRoute.GET("/search", controller.SearchRedemptions)
   		 redemptionRoute.GET("/:id", controller.GetRedemption)
   		 redemptionRoute.POST("/", controller.AddRedemption)
   		 redemptionRoute.PUT("/", controller.UpdateRedemption)
   		 redemptionRoute.DELETE("/:id", controller.DeleteRedemption)
   	 }

   	 // 创建log路由分组
   	 logRoute := apiRouter.Group("/")
   	 logRoute.GET("/", middleware.AdminAuth(), controller.GetAllLogs)
   	 logRoute.DELETE("/", middleware.AdminAuth(), controller.DeleteHistoryLogs)
   	 logRoute.GET("/stat", middleware.AdminAuth(), controller.GetLogsStat)
   	 logRoute.GET("/self/stat", middleware.UserAuth(), controller.GetLogsSelfStat)
   	 logRoute.GET("/search", middleware.AdminAuth(), controller.SearchAllLogs)
   	 logRoute.GET("/self", middleware.UserAuth(), controller.GetUserLogs)
   	 logRoute.GET("/self/search", middleware.UserAuth(), controller.SearchUserLogs)

   	 // 创建group路由分组，并使用管理员认证中间件
   	 groupRoute := apiRouter.Group("/group")
   	 groupRoute.Use(middleware.AdminAuth())
   	 {
   		 // 设置GET请求的路由处理函数
   		 groupRoute.GET("/", controller.GetGroups)
   	 }
    }
}