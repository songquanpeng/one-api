package common

import "fmt"

func LogQuota(quota int) string {
	if DisplayInCurrencyEnabled {
		return fmt.Sprintf("＄%.6f 额度", float64(quota)/QuotaPerUnit)
	} else {
		return fmt.Sprintf("%d 点额度", quota)
	}
}
