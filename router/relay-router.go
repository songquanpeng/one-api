package router

import (
	"one-api/controller"
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
		relayV1Router.GET("/files", controller.RelayNotImplemented)
		relayV1Router.POST("/files", controller.RelayNotImplemented)
		relayV1Router.DELETE("/files/:id", controller.RelayNotImplemented)
		relayV1Router.GET("/files/:id", controller.RelayNotImplemented)
		relayV1Router.GET("/files/:id/content", controller.RelayNotImplemented)
		relayV1Router.POST("/fine_tuning/jobs", controller.RelayNotImplemented)
		relayV1Router.GET("/fine_tuning/jobs", controller.RelayNotImplemented)
		relayV1Router.GET("/fine_tuning/jobs/:id", controller.RelayNotImplemented)
		relayV1Router.POST("/fine_tuning/jobs/:id/cancel", controller.RelayNotImplemented)
		relayV1Router.GET("/fine_tuning/jobs/:id/events", controller.RelayNotImplemented)
		relayV1Router.DELETE("/models/:model", controller.RelayNotImplemented)
		relayV1Router.POST("/assistants", controller.RelayNotImplemented)
		relayV1Router.GET("/assistants/:id", controller.RelayNotImplemented)
		relayV1Router.POST("/assistants/:id", controller.RelayNotImplemented)
		relayV1Router.DELETE("/assistants/:id", controller.RelayNotImplemented)
		relayV1Router.GET("/assistants", controller.RelayNotImplemented)
		relayV1Router.POST("/assistants/:id/files", controller.RelayNotImplemented)
		relayV1Router.GET("/assistants/:id/files/:fileId", controller.RelayNotImplemented)
		relayV1Router.DELETE("/assistants/:id/files/:fileId", controller.RelayNotImplemented)
		relayV1Router.GET("/assistants/:id/files", controller.RelayNotImplemented)
		relayV1Router.POST("/threads", controller.RelayNotImplemented)
		relayV1Router.GET("/threads/:id", controller.RelayNotImplemented)
		relayV1Router.POST("/threads/:id", controller.RelayNotImplemented)
		relayV1Router.DELETE("/threads/:id", controller.RelayNotImplemented)
		relayV1Router.POST("/threads/:id/messages", controller.RelayNotImplemented)
		relayV1Router.GET("/threads/:id/messages/:messageId", controller.RelayNotImplemented)
		relayV1Router.POST("/threads/:id/messages/:messageId", controller.RelayNotImplemented)
		relayV1Router.GET("/threads/:id/messages/:messageId/files/:filesId", controller.RelayNotImplemented)
		relayV1Router.GET("/threads/:id/messages/:messageId/files", controller.RelayNotImplemented)
		relayV1Router.POST("/threads/:id/runs", controller.RelayNotImplemented)
		relayV1Router.GET("/threads/:id/runs/:runsId", controller.RelayNotImplemented)
		relayV1Router.POST("/threads/:id/runs/:runsId", controller.RelayNotImplemented)
		relayV1Router.GET("/threads/:id/runs", controller.RelayNotImplemented)
		relayV1Router.POST("/threads/:id/runs/:runsId/submit_tool_outputs", controller.RelayNotImplemented)
		relayV1Router.POST("/threads/:id/runs/:runsId/cancel", controller.RelayNotImplemented)
		relayV1Router.GET("/threads/:id/runs/:runsId/steps/:stepId", controller.RelayNotImplemented)
		relayV1Router.GET("/threads/:id/runs/:runsId/steps", controller.RelayNotImplemented)
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
