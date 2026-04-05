package dateparse

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// ToMillis converts a human-friendly date string to a Unix timestamp in milliseconds.
// Supported formats:
//   - ISO date: "2026-04-01"
//   - ISO datetime: "2026-04-01T14:30:00"
//   - Relative days: "7d" (7 days ago from now)
//   - Named: "today", "yesterday"
//
// Returns the timestamp as a string (matching the API's format).
func ToMillis(s string, now time.Time) (string, error) {
	s = strings.TrimSpace(s)

	switch strings.ToLower(s) {
	case "today":
		start := startOfDay(now)
		return msStr(start), nil
	case "yesterday":
		start := startOfDay(now.AddDate(0, 0, -1))
		return msStr(start), nil
	}

	// Relative days: "7d", "30d"
	if strings.HasSuffix(strings.ToLower(s), "d") {
		numStr := s[:len(s)-1]
		days, err := strconv.Atoi(numStr)
		if err != nil {
			return "", fmt.Errorf("invalid relative date %q: %w", s, err)
		}
		t := startOfDay(now.AddDate(0, 0, -days))
		return msStr(t), nil
	}

	// ISO datetime: "2026-04-01T14:30:00"
	if t, err := time.ParseInLocation("2006-01-02T15:04:05", s, now.Location()); err == nil {
		return msStr(t), nil
	}

	// ISO date: "2026-04-01"
	if t, err := time.ParseInLocation("2006-01-02", s, now.Location()); err == nil {
		return msStr(t), nil
	}

	return "", fmt.Errorf("unrecognized date format %q (use YYYY-MM-DD, Nd, today, or yesterday)", s)
}

// EndOfDayMillis converts a date string to the end of that day in milliseconds.
// For relative formats like "7d", it returns end-of-day for 7 days ago.
// For "today"/"yesterday", it returns end of that day.
func EndOfDayMillis(s string, now time.Time) (string, error) {
	ms, err := ToMillis(s, now)
	if err != nil {
		return "", err
	}

	n, _ := strconv.ParseInt(ms, 10, 64)
	t := time.UnixMilli(n).In(now.Location())
	endOfDay := time.Date(t.Year(), t.Month(), t.Day(), 23, 59, 59, 999000000, now.Location())
	return msStr(endOfDay), nil
}

func startOfDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}

func msStr(t time.Time) string {
	return strconv.FormatInt(t.UnixMilli(), 10)
}
