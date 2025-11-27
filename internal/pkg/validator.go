package pkg

import (
	"regexp"
	"strings"
)

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

func IsValidEmail(email string) bool {
	return emailRegex.MatchString(email)
}

func IsValidPassword(password string) bool {
	// 최소 6자 이상
	return len(password) >= 6
}

func SanitizeString(s string) string {
	return strings.TrimSpace(s)
}

func IsValidLevel(level int) bool {
	return level >= 0 && level <= 5
}
