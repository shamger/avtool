package utils

import (
	"math/rand"
	"strings"
)

func GenRandomString(length int, validChars string) string {
	strs := make([]string, length)
	chars := strings.Split(validChars, "")
	for i := range strs {
		strs[i] = chars[rand.Intn(len(chars))]
	}
	return strings.Join(strs, "")
}
