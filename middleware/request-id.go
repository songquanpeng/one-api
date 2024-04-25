package middleware

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/songquanpeng/one-api/common/ctxkey"
	"github.com/songquanpeng/one-api/common/helper"
)

func RequestId() func(c *gin.Context) {
	return func(c *gin.Context) {
		id := helper.GenRequestID()
		c.Set(ctxkey.RequestId, id)
		ctx := context.WithValue(c.Request.Context(), ctxkey.RequestId, id)
		c.Request = c.Request.WithContext(ctx)
		c.Header(ctxkey.RequestId, id)
		c.Next()
	}
}
