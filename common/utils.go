package common

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/songquanpeng/one-api/common/config"
)

func LogQuota(quota int64) string {
	if config.DisplayInCurrencyEnabled {
		return fmt.Sprintf("＄%.6f 额度", float64(quota)/config.QuotaPerUnit)
	} else {
		return fmt.Sprintf("%d 点额度", quota)
	}
}

func RenderStringData(c *gin.Context, data string) {
	data = strings.TrimPrefix(data, "data: ")
	c.Render(-1, CustomEvent{Data: "data: " + strings.TrimSuffix(data, "\r")})
	c.Writer.Flush()
}

func RenderData(c *gin.Context, response interface{}) error {
	jsonResponse, err := json.Marshal(response)
	if err != nil {
		return fmt.Errorf("error marshalling stream response: %w", err)
	}
	RenderStringData(c, string(jsonResponse))
	return nil
}
