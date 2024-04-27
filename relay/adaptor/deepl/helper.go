package deepl

import "strings"

func parseLangFromModelName(modelName string) string {
	parts := strings.Split(modelName, "-")
	if len(parts) == 1 {
		return "ZH"
	}
	return parts[1]
}
