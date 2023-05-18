package router

import (
	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
	"one-api/controller"
	"one-api/middleware"
)

func SetDashboardRouter(router *gin.Engine) {
	apiRouter := router.Group("/dashboard")
	apiRouter.Use(gzip.Gzip(gzip.DefaultCompression))
	apiRouter.Use(middleware.GlobalAPIRateLimit())
	apiRouter.Use(middleware.TokenAuth())
	{
		apiRouter.GET("/billing/credit_grants", controller.GetTokenStatus)
		apiRouter.GET("/billing/subscription", controller.GetTokenStatus)
		apiRouter.GET("/billing/usage", controller.GetTokenStatus)
	}
}
