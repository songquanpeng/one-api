package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/songquanpeng/one-api/common/i18n"
)

func Language() gin.HandlerFunc {
	return func(c *gin.Context) {
		lang := c.GetHeader("Accept-Language")
		if lang == "" {
			lang = "en"
		}
		if strings.HasPrefix(strings.ToLower(lang), "zh") {
			lang = "zh-CN"
		} else {
			lang = "en"
		}
		c.Set(i18n.ContextKey, lang)
		c.Next()
	}
}
