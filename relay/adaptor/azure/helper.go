package azure

import (
	"github.com/gin-gonic/gin"
	"github.com/songquanpeng/one-api/common/ctxkey"
)

func GetAPIVersion(c *gin.Context) string {
	query := c.Request.URL.Query()
	apiVersion := query.Get("api-version")
	if apiVersion == "" {
		apiVersion = c.GetString(ctxkey.ConfigAPIVersion)
	}
	return apiVersion
}
