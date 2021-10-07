package indexer

func stringSliceContains(slice []string, s string) bool {
	for _, item := range slice {
		if item == s {
			return true
		}
	}
	return false
}

func boolToInt64(v bool) int64 {
	if v {
		return 1
	}
	return 0
}
