package utils

import "strings"

// Contains returns true if the given string is in the given slice.
func Contains(slice []string, entry string) bool {
	for _, a := range slice {
		if a == entry {
			return true
		}
	}
	return false
}

// TrimNewline returns the given string without the newline character.
func TrimNewline(s string) string {
	return strings.TrimSuffix(s, "\n")
}
