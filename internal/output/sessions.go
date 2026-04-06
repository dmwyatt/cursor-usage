package output

import (
	"sort"
	"strconv"
	"time"
)

// ActiveHours calculates total active hours from a set of event timestamps.
// Events are grouped into sessions: consecutive events less than gap apart
// belong to the same session. The duration of each session is the time between
// its first and last event. Returns total hours across all sessions.
func ActiveHours(timestamps []string, gap time.Duration) float64 {
	times := parseMsTimestamps(timestamps)
	if len(times) < 2 {
		return 0
	}

	sort.Slice(times, func(i, j int) bool { return times[i].Before(times[j]) })

	var totalDuration time.Duration
	sessionStart := times[0]
	prev := times[0]

	for _, t := range times[1:] {
		if t.Sub(prev) > gap {
			totalDuration += prev.Sub(sessionStart)
			sessionStart = t
		}
		prev = t
	}
	totalDuration += prev.Sub(sessionStart)

	return totalDuration.Hours()
}

func parseMsTimestamps(timestamps []string) []time.Time {
	times := make([]time.Time, 0, len(timestamps))
	for _, ts := range timestamps {
		n, err := strconv.ParseInt(ts, 10, 64)
		if err != nil {
			continue
		}
		times = append(times, time.UnixMilli(n))
	}
	return times
}
