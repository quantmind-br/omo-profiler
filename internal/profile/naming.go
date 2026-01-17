package profile

import (
	"errors"
	"regexp"
	"strings"
)

var validNameRegex = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)

var ErrInvalidName = errors.New("profile name must contain only letters, numbers, underscores, and hyphens")

var ErrEmptyName = errors.New("profile name cannot be empty")

func ValidateName(name string) error {
	if name == "" {
		return ErrEmptyName
	}
	if !validNameRegex.MatchString(name) {
		return ErrInvalidName
	}
	return nil
}

func SanitizeName(name string) string {
	re := regexp.MustCompile(`[^a-zA-Z0-9_-]`)
	sanitized := re.ReplaceAllString(name, "")
	sanitized = strings.Trim(sanitized, "-_")
	return sanitized
}
