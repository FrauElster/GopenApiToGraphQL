package util

func IsInSlice[T comparable](obj T, slice []T) bool {
	for _, it := range slice {
		if it == obj {
			return true
		}
	}
	return false
}

func FilterSlice[T any](slice []T, decider func(T) bool) []T {
	result := make([]T, 0)
	for _, it := range slice {
		if decider(it) {
			result = append(result, it)
		}
	}
	return result
}
