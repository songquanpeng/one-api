package router

import (
	"github.com/gin-gonic/gin"
	"one-api/controller"
	"one-api/middleware"
)

func SetRelayRouter(router *gin.Engine) {
	relayV1Router := router.Group("/v1")
	relayV1Router.Use(middleware.TokenAuth(), middleware.Distribute())
	{
		relayV1Router.Any("/*path", controller.Relay)
	}
	relayDashboardRouter := router.Group("/dashboard")
	relayDashboardRouter.Use(middleware.TokenAuth(), middleware.Distribute())
	{
		relayDashboardRouter.Any("/*path", controller.Relay)
	}
}
