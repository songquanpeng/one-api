package router

import (
	"github.com/gin-gonic/gin"
	"one-api/controller"
	"one-api/middleware"
)

func SetRelayRouter(router *gin.Engine) {
	// https://platform.openai.com/docs/api-reference/introduction
	relayV1Router := router.Group("/v1")
	relayV1Router.Use(middleware.TokenAuth(), middleware.Distribute())
	{
		relayV1Router.GET("/models", controller.ListModels)
		relayV1Router.GET("/models/:model", controller.RetrieveModel)
		relayV1Router.POST("/completions", controller.RelayNotImplemented)
		relayV1Router.POST("/chat/completions", controller.Relay)
		relayV1Router.POST("/edits", controller.RelayNotImplemented)
		relayV1Router.POST("/images/generations", controller.RelayNotImplemented)
		relayV1Router.POST("/images/edits", controller.RelayNotImplemented)
		relayV1Router.POST("/images/variations", controller.RelayNotImplemented)
		relayV1Router.POST("/embeddings", controller.Relay)
		relayV1Router.POST("/audio/transcriptions", controller.RelayNotImplemented)
		relayV1Router.POST("/audio/translations", controller.RelayNotImplemented)
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
		relayV1Router.POST("/moderations", controller.RelayNotImplemented)
	}
}
