package utils

import "strings"

// INFO: Return last n strings joined by newline
func Tail(strs []string, n int) string {
	if len(strs) > n {
		strs = strs[len(strs)-n:]
	}

	return strings.Join(strs, "\n")
}
