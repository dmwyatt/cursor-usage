package output

import (
	"fmt"
	"testing"
	"time"
)

func TestActiveHoursBasicSessions(t *testing.T) {
	// Two sessions:
	// Session 1: 2:00, 2:03, 2:15, 2:18 (18 min)
	// Session 2: 4:30, 4:35 (5 min)
	// Total: 23 min
	timestamps := []string{
		msFromTime(t, "2026-04-05T14:00:00Z"),
		msFromTime(t, "2026-04-05T14:03:00Z"),
		msFromTime(t, "2026-04-05T14:15:00Z"),
		msFromTime(t, "2026-04-05T14:18:00Z"),
		msFromTime(t, "2026-04-05T16:30:00Z"),
		msFromTime(t, "2026-04-05T16:35:00Z"),
	}

	hours := ActiveHours(timestamps, 30*time.Minute)

	expectedMinutes := 23.0
	gotMinutes := hours * 60
	if diff := gotMinutes - expectedMinutes; diff < -0.01 || diff > 0.01 {
		t.Errorf("expected ~%.1f minutes, got %.1f", expectedMinutes, gotMinutes)
	}
}

func TestActiveHoursSingleEvent(t *testing.T) {
	timestamps := []string{
		msFromTime(t, "2026-04-05T14:00:00Z"),
	}

	hours := ActiveHours(timestamps, 30*time.Minute)

	// Single event has zero duration
	if hours != 0 {
		t.Errorf("expected 0 hours for single event, got %f", hours)
	}
}

func TestActiveHoursEmpty(t *testing.T) {
	hours := ActiveHours(nil, 30*time.Minute)
	if hours != 0 {
		t.Errorf("expected 0, got %f", hours)
	}
}

func TestActiveHoursAllWithinOneSession(t *testing.T) {
	timestamps := []string{
		msFromTime(t, "2026-04-05T14:00:00Z"),
		msFromTime(t, "2026-04-05T14:10:00Z"),
		msFromTime(t, "2026-04-05T14:20:00Z"),
		msFromTime(t, "2026-04-05T14:45:00Z"),
	}

	hours := ActiveHours(timestamps, 30*time.Minute)

	// One session: 14:00 - 14:45 = 45 min = 0.75 hours
	if diff := hours - 0.75; diff < -0.01 || diff > 0.01 {
		t.Errorf("expected 0.75 hours, got %f", hours)
	}
}

func TestActiveHoursUnsortedInput(t *testing.T) {
	// Same as basic test but shuffled order
	timestamps := []string{
		msFromTime(t, "2026-04-05T16:35:00Z"),
		msFromTime(t, "2026-04-05T14:15:00Z"),
		msFromTime(t, "2026-04-05T14:00:00Z"),
		msFromTime(t, "2026-04-05T16:30:00Z"),
		msFromTime(t, "2026-04-05T14:18:00Z"),
		msFromTime(t, "2026-04-05T14:03:00Z"),
	}

	hours := ActiveHours(timestamps, 30*time.Minute)

	expectedMinutes := 23.0
	gotMinutes := hours * 60
	if diff := gotMinutes - expectedMinutes; diff < -0.01 || diff > 0.01 {
		t.Errorf("expected ~%.1f minutes, got %.1f", expectedMinutes, gotMinutes)
	}
}

func TestActiveHoursCustomGap(t *testing.T) {
	// With a 10-minute gap, the basic test becomes 3 sessions:
	// Session 1: 2:00, 2:03 (3 min)
	// Session 2: 2:15, 2:18 (3 min)
	// Session 3: 4:30, 4:35 (5 min)
	// Total: 11 min
	timestamps := []string{
		msFromTime(t, "2026-04-05T14:00:00Z"),
		msFromTime(t, "2026-04-05T14:03:00Z"),
		msFromTime(t, "2026-04-05T14:15:00Z"),
		msFromTime(t, "2026-04-05T14:18:00Z"),
		msFromTime(t, "2026-04-05T16:30:00Z"),
		msFromTime(t, "2026-04-05T16:35:00Z"),
	}

	hours := ActiveHours(timestamps, 10*time.Minute)

	expectedMinutes := 11.0
	gotMinutes := hours * 60
	if diff := gotMinutes - expectedMinutes; diff < -0.01 || diff > 0.01 {
		t.Errorf("expected ~%.1f minutes, got %.1f", expectedMinutes, gotMinutes)
	}
}

func msFromTime(t *testing.T, iso string) string {
	t.Helper()
	parsed, err := time.Parse(time.RFC3339, iso)
	if err != nil {
		t.Fatalf("bad test timestamp %q: %v", iso, err)
	}
	return formatInt64(parsed.UnixMilli())
}

func formatInt64(n int64) string {
	return fmt.Sprintf("%d", n)
}
