package middleware

import (
	"context"
	"one-api/common"
	"one-api/common/utils"
	"time"

	"github.com/gin-gonic/gin"
)

func RequestId() func(c *gin.Context) {
	return func(c *gin.Context) {
		id := utils.GetTimeString() + utils.GetRandomString(8)
		c.Set(common.RequestIdKey, id)
		ctx := context.WithValue(c.Request.Context(), common.RequestIdKey, id)
		ctx = context.WithValue(ctx, "requestStartTime", time.Now())
		c.Request = c.Request.WithContext(ctx)
		c.Header(common.RequestIdKey, id)
		c.Next()
	}
}
