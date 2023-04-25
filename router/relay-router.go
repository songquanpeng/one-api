package router

import (
	"github.com/gin-gonic/gin"
	"one-api/controller"
	"one-api/middleware"
)

func SetRelayRouter(router *gin.Engine) {
	relayRouter := router.Group("/v1")
	relayRouter.Use(middleware.GlobalAPIRateLimit(), middleware.TokenAuth(), middleware.Distribute())
	{
		relayRouter.Any("/*path", controller.Relay)
	}
}
