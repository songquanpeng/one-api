package conv

func AsString(v any) string {
	if str, ok := v.(string); ok {
		return str
	}

	return ""
}
