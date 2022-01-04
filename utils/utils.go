package utils

// Contains returns true if the given string is in the given slice.
func Contains(slice []string, entry string) bool {
	for _, a := range slice {
		if a == entry {
			return true
		}
	}
	return false
}
