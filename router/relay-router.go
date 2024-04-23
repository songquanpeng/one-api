package router

import (
	"one-api/middleware"
	"one-api/relay"
	"one-api/relay/midjourney"

	"github.com/gin-gonic/gin"
)

func SetRelayRouter(router *gin.Engine) {
	router.Use(middleware.CORS())
	// https://platform.openai.com/docs/api-reference/introduction
	setOpenAIRouter(router)
	setMJRouter(router)
}

func setOpenAIRouter(router *gin.Engine) {
	modelsRouter := router.Group("/v1/models")
	modelsRouter.Use(middleware.OpenaiAuth(), middleware.Distribute())
	{
		modelsRouter.GET("", relay.ListModels)
		modelsRouter.GET("/:model", relay.RetrieveModel)
	}
	relayV1Router := router.Group("/v1")
	relayV1Router.Use(middleware.RelayPanicRecover(), middleware.OpenaiAuth(), middleware.Distribute())
	{
		relayV1Router.POST("/completions", relay.Relay)
		relayV1Router.POST("/chat/completions", relay.Relay)
		// relayV1Router.POST("/edits", controller.Relay)
		relayV1Router.POST("/images/generations", relay.Relay)
		relayV1Router.POST("/images/edits", relay.Relay)
		relayV1Router.POST("/images/variations", relay.Relay)
		relayV1Router.POST("/embeddings", relay.Relay)
		// relayV1Router.POST("/engines/:model/embeddings", controller.RelayEmbeddings)
		relayV1Router.POST("/audio/transcriptions", relay.Relay)
		relayV1Router.POST("/audio/translations", relay.Relay)
		relayV1Router.POST("/audio/speech", relay.Relay)
		relayV1Router.POST("/moderations", relay.Relay)

		relayV1Router.Use(middleware.SpecifiedChannel())
		{
			relayV1Router.Any("/files", relay.RelayOnly)
			relayV1Router.Any("/files/*any", relay.RelayOnly)
			relayV1Router.Any("/fine_tuning/*any", relay.RelayOnly)
			relayV1Router.Any("/assistants", relay.RelayOnly)
			relayV1Router.Any("/assistants/*any", relay.RelayOnly)
			relayV1Router.Any("/threads", relay.RelayOnly)
			relayV1Router.Any("/threads/*any", relay.RelayOnly)
			relayV1Router.Any("/batches/*any", relay.RelayOnly)
			relayV1Router.Any("/vector_stores/*any", relay.RelayOnly)
			relayV1Router.DELETE("/models/:model", relay.RelayOnly)
		}
	}
}

func setMJRouter(router *gin.Engine) {
	relayMjRouter := router.Group("/mj")
	registerMjRouterGroup(relayMjRouter)

	relayMjModeRouter := router.Group("/:mode/mj")
	registerMjRouterGroup(relayMjModeRouter)
}

// Author: Calcium-Ion
// GitHub: https://github.com/Calcium-Ion/new-api
// Path: router/relay-router.go
func registerMjRouterGroup(relayMjRouter *gin.RouterGroup) {
	relayMjRouter.GET("/image/:id", midjourney.RelayMidjourneyImage)
	relayMjRouter.Use(middleware.MjAuth(), middleware.Distribute())
	{
		relayMjRouter.POST("/submit/action", midjourney.RelayMidjourney)
		relayMjRouter.POST("/submit/shorten", midjourney.RelayMidjourney)
		relayMjRouter.POST("/submit/modal", midjourney.RelayMidjourney)
		relayMjRouter.POST("/submit/imagine", midjourney.RelayMidjourney)
		relayMjRouter.POST("/submit/change", midjourney.RelayMidjourney)
		relayMjRouter.POST("/submit/simple-change", midjourney.RelayMidjourney)
		relayMjRouter.POST("/submit/describe", midjourney.RelayMidjourney)
		relayMjRouter.POST("/submit/blend", midjourney.RelayMidjourney)
		relayMjRouter.POST("/notify", midjourney.RelayMidjourney)
		relayMjRouter.GET("/task/:id/fetch", midjourney.RelayMidjourney)
		relayMjRouter.GET("/task/:id/image-seed", midjourney.RelayMidjourney)
		relayMjRouter.POST("/task/list-by-condition", midjourney.RelayMidjourney)
		relayMjRouter.POST("/insight-face/swap", midjourney.RelayMidjourney)
	}
}
