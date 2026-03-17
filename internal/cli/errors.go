package cli

import (
	"strings"
)

const (
	Success      = 0
	GeneralError = 1
	UsageError   = 2
)

func ExitCodeForError(err error) int {
	if err == nil {
		return Success
	}

	if isUsageError(err) {
		return UsageError
	}

	return GeneralError
}

func isUsageError(err error) bool {
	if err == nil {
		return false
	}

	errText := strings.ToLower(err.Error())
	usageErrorFragments := []string{
		"unknown flag",
		"invalid argument",
		"requires at least",
		"accepts",
	}

	for _, fragment := range usageErrorFragments {
		if strings.Contains(errText, fragment) {
			return true
		}
	}

	return false
}
