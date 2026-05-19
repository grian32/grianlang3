package util

import (
	"strings"
)

func ContainsOnly(s, chars string) bool {
	return strings.IndexFunc(s, func(r rune) bool {
		return !strings.ContainsRune(chars, r)
	}) == -1
}

// GetFileNamePath gets the filename out of a full path, f.e a/b/c/example.gl3 returns example.gl3 and a standalone exmaple.gl3 will return example.gl3
func GetFileNamePath(s string) string {
	if !strings.Contains(s, "/") {
		return s
	}
	_, fileName, _ := strings.Cut(s, "/")
	return fileName
}
