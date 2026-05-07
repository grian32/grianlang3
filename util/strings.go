package util

import "strings"

func ContainsOnly(s, chars string) bool {
	return strings.IndexFunc(s, func(r rune) bool {
		return !strings.ContainsRune(chars, r)
	}) == -1
}
