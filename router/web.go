package router

import (
	"embed"
	"fmt"
	"github.com/gin-contrib/gzip"
	"github.com/gin-contrib/static"
	"github.com/gin-gonic/gin"
	"github.com/songquanpeng/one-api/common"
	"github.com/songquanpeng/one-api/common/config"
	"github.com/songquanpeng/one-api/controller"
	"github.com/songquanpeng/one-api/middleware"
	"net/http"
	"strings"
)

func SetWebRouter(engine *gin.Engine, baseUrl string, buildFS embed.FS) {
	basePath := fmt.Sprintf("web/build/%s", config.Theme)
	indexPageData, _ := buildFS.ReadFile(fmt.Sprintf("%s/index.html", basePath))
	engine.Use(gzip.Gzip(gzip.DefaultCompression))
	engine.Use(middleware.GlobalWebRateLimit())
	engine.Use(middleware.Cache())
	engine.Use(static.Serve(baseUrl, common.EmbedFolder(buildFS, basePath)))

	engine.NoRoute(func(c *gin.Context) {
		if strings.HasPrefix(c.Request.RequestURI, "/v1") || strings.HasPrefix(c.Request.RequestURI, "/api") {
			controller.RelayNotFound(c)
			return
		}
		c.Header("Cache-Control", "no-cache")
		c.Data(http.StatusOK, "text/html; charset=utf-8", indexPageData)
	})
}
