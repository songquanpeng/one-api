package utils

func DeDuplication(slice []string) []string {
	m := make(map[string]bool)
	for _, v := range slice {
		m[v] = true
	}
	result := make([]string, 0, len(m))
	for v := range m {
		result = append(result, v)
	}
	return result
}
