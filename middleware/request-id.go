package middleware

import (
	"context"
	"one-api/common"

	"github.com/gin-gonic/gin"
)

func RequestId() func(c *gin.Context) {
	return func(c *gin.Context) {
		type RequestIdKeyType string

		const RequestIdKey RequestIdKeyType = common.RequestIdKey

		id := common.GetTimeString() + common.GetRandomString(8)
		c.Set(common.RequestIdKey, id)
		ctx := context.WithValue(c.Request.Context(), RequestIdKey, id)
		c.Request = c.Request.WithContext(ctx)
		c.Header(common.RequestIdKey, id)
		c.Next()
	}
}
