package cli

import (
	"bufio"
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var conferenceSlugPattern = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)

func promptField(reader *bufio.Reader, out io.Writer, label string, defaultValue string) (string, error) {
	if defaultValue == "" {
		_, _ = fmt.Fprintf(out, "%s: ", label)
	} else {
		_, _ = fmt.Fprintf(out, "%s [%s]: ", label, defaultValue)
	}

	line, err := reader.ReadString('\n')
	if err != nil && line == "" {
		return "", err
	}

	value := strings.TrimSpace(line)
	if value == "" {
		return defaultValue, nil
	}

	return value, nil
}

func normalizeConferenceSlug(raw string) string {
	normalized := strings.ToLower(strings.TrimSpace(raw))
	normalized = strings.ReplaceAll(normalized, "_", "-")
	normalized = strings.ReplaceAll(normalized, " ", "-")
	normalized = regexp.MustCompile(`-+`).ReplaceAllString(normalized, "-")
	normalized = strings.Trim(normalized, "-")
	return normalized
}

func parseConferenceDate(value string) (time.Time, error) {
	return time.Parse("2006-01-02", strings.TrimSpace(value))
}

func parseConferenceCapacity(value string) (int, error) {
	capacity, err := strconv.Atoi(strings.TrimSpace(value))
	if err != nil || capacity <= 0 {
		return 0, fmt.Errorf("capacity must be a positive integer")
	}
	return capacity, nil
}
