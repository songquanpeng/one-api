package common

import (
	"fmt"
	"one-api/common/config"
)

func LogQuota(quota int) string {
	if config.DisplayInCurrencyEnabled {
		return fmt.Sprintf("＄%.6f 额度", float64(quota)/config.QuotaPerUnit)
	} else {
		return fmt.Sprintf("%d 点额度", quota)
	}
}
