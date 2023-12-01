package router

import (
	"one-api/controller"
	"one-api/middleware"

	"github.com/gin-gonic/gin"
)

func SetRelayRouter(router *gin.Engine) {
	router.Use(middleware.CORS())
	// https://platform.openai.com/docs/api-reference/introduction
	modelsRouter := router.Group("/v1/models")
	modelsRouter.Use(middleware.TokenAuth())
	{
		modelsRouter.GET("", controller.ListModels)
		modelsRouter.GET("/:model", controller.RetrieveModel)
	}
	relayV1Router := router.Group("/v1")
	relayV1Router.Use(middleware.TokenAuth(), middleware.Distribute())
	{
		relayV1Router.POST("/completions", controller.RelayCompletions)
		relayV1Router.POST("/chat/completions", controller.RelayChat)
		// relayV1Router.POST("/edits", controller.Relay)
		relayV1Router.POST("/images/generations", controller.RelayImageGenerations)
		relayV1Router.POST("/images/edits", controller.RelayImageEdits)
		relayV1Router.POST("/images/variations", controller.RelayImageVariations)
		relayV1Router.POST("/embeddings", controller.RelayEmbeddings)
		// relayV1Router.POST("/engines/:model/embeddings", controller.RelayEmbeddings)
		relayV1Router.POST("/audio/transcriptions", controller.RelayTranscriptions)
		relayV1Router.POST("/audio/translations", controller.RelayTranslations)
		relayV1Router.POST("/audio/speech", controller.RelaySpeech)
		relayV1Router.POST("/moderations", controller.RelayModerations)
		relayV1Router.GET("/files", controller.RelayNotImplemented)
		relayV1Router.POST("/files", controller.RelayNotImplemented)
		relayV1Router.DELETE("/files/:id", controller.RelayNotImplemented)
		relayV1Router.GET("/files/:id", controller.RelayNotImplemented)
		relayV1Router.GET("/files/:id/content", controller.RelayNotImplemented)
		relayV1Router.POST("/fine-tunes", controller.RelayNotImplemented)
		relayV1Router.GET("/fine-tunes", controller.RelayNotImplemented)
		relayV1Router.GET("/fine-tunes/:id", controller.RelayNotImplemented)
		relayV1Router.POST("/fine-tunes/:id/cancel", controller.RelayNotImplemented)
		relayV1Router.GET("/fine-tunes/:id/events", controller.RelayNotImplemented)
		relayV1Router.DELETE("/models/:model", controller.RelayNotImplemented)
	}
}
