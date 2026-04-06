package output

import (
	"bytes"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/dmwyatt/cursor-usage/internal/api"
)

func makeSessionTestEvents() []api.UsageEvent {
	return []api.UsageEvent{
		// Session 1: 14:00 - 14:18
		{
			Timestamp:    fmt.Sprintf("%d", time.Date(2026, 4, 5, 14, 0, 0, 0, time.UTC).UnixMilli()),
			Model:        "claude-4.6-opus",
			ChargedCents: 50.0,
			TokenUsage:   api.TokenUsage{InputTokens: 100, OutputTokens: 500},
		},
		{
			Timestamp:    fmt.Sprintf("%d", time.Date(2026, 4, 5, 14, 3, 0, 0, time.UTC).UnixMilli()),
			Model:        "claude-4.6-opus",
			ChargedCents: 30.0,
			TokenUsage:   api.TokenUsage{InputTokens: 80, OutputTokens: 300},
		},
		{
			Timestamp:    fmt.Sprintf("%d", time.Date(2026, 4, 5, 14, 15, 0, 0, time.UTC).UnixMilli()),
			Model:        "claude-4.6-sonnet",
			ChargedCents: 10.0,
			TokenUsage:   api.TokenUsage{InputTokens: 50, OutputTokens: 200},
		},
		{
			Timestamp:    fmt.Sprintf("%d", time.Date(2026, 4, 5, 14, 18, 0, 0, time.UTC).UnixMilli()),
			Model:        "claude-4.6-opus",
			ChargedCents: 20.0,
			TokenUsage:   api.TokenUsage{InputTokens: 60, OutputTokens: 250},
		},
		// Session 2: 16:30 - 16:35
		{
			Timestamp:    fmt.Sprintf("%d", time.Date(2026, 4, 5, 16, 30, 0, 0, time.UTC).UnixMilli()),
			Model:        "claude-4.6-sonnet",
			ChargedCents: 15.0,
			TokenUsage:   api.TokenUsage{InputTokens: 40, OutputTokens: 180},
		},
		{
			Timestamp:    fmt.Sprintf("%d", time.Date(2026, 4, 5, 16, 35, 0, 0, time.UTC).UnixMilli()),
			Model:        "claude-4.6-sonnet",
			ChargedCents: 12.0,
			TokenUsage:   api.TokenUsage{InputTokens: 30, OutputTokens: 150},
		},
	}
}

func TestGroupSessionsCount(t *testing.T) {
	sessions := GroupSessions(makeSessionTestEvents(), 30*time.Minute)

	if len(sessions) != 2 {
		t.Fatalf("expected 2 sessions, got %d", len(sessions))
	}
}

func TestGroupSessionsFirstSession(t *testing.T) {
	sessions := GroupSessions(makeSessionTestEvents(), 30*time.Minute)
	s := sessions[0]

	if len(s.Events) != 4 {
		t.Errorf("session 1: expected 4 events, got %d", len(s.Events))
	}

	expectedDuration := 18 * time.Minute
	if diff := s.Duration - expectedDuration; diff < -time.Second || diff > time.Second {
		t.Errorf("session 1: expected duration %v, got %v", expectedDuration, s.Duration)
	}

	if s.TotalChargedCents != 110.0 {
		t.Errorf("session 1: expected 110.0 cents, got %f", s.TotalChargedCents)
	}

	if len(s.ByModel) != 2 {
		t.Errorf("session 1: expected 2 models, got %d", len(s.ByModel))
	}
}

func TestGroupSessionsSecondSession(t *testing.T) {
	sessions := GroupSessions(makeSessionTestEvents(), 30*time.Minute)
	s := sessions[1]

	if len(s.Events) != 2 {
		t.Errorf("session 2: expected 2 events, got %d", len(s.Events))
	}

	expectedDuration := 5 * time.Minute
	if diff := s.Duration - expectedDuration; diff < -time.Second || diff > time.Second {
		t.Errorf("session 2: expected duration %v, got %v", expectedDuration, s.Duration)
	}

	if s.TotalChargedCents != 27.0 {
		t.Errorf("session 2: expected 27.0 cents, got %f", s.TotalChargedCents)
	}
}

func TestGroupSessionsModelBreakdown(t *testing.T) {
	sessions := GroupSessions(makeSessionTestEvents(), 30*time.Minute)
	s := sessions[0]

	modelMap := map[string]SessionModelSummary{}
	for _, m := range s.ByModel {
		modelMap[m.Model] = m
	}

	opus, ok := modelMap["claude-4.6-opus"]
	if !ok {
		t.Fatal("session 1: missing opus in model breakdown")
	}
	if opus.Events != 3 {
		t.Errorf("session 1 opus: expected 3 events, got %d", opus.Events)
	}
	if opus.ChargedCents != 100.0 {
		t.Errorf("session 1 opus: expected 100.0 cents, got %f", opus.ChargedCents)
	}
}

func TestGroupSessionsSingleEvent(t *testing.T) {
	events := []api.UsageEvent{
		{
			Timestamp:    fmt.Sprintf("%d", time.Date(2026, 4, 5, 14, 0, 0, 0, time.UTC).UnixMilli()),
			Model:        "claude-4.6-opus",
			ChargedCents: 50.0,
		},
	}

	sessions := GroupSessions(events, 30*time.Minute)

	if len(sessions) != 1 {
		t.Fatalf("expected 1 session, got %d", len(sessions))
	}
	if sessions[0].Duration != 0 {
		t.Errorf("expected 0 duration for single event, got %v", sessions[0].Duration)
	}
}

func TestGroupSessionsEmpty(t *testing.T) {
	sessions := GroupSessions(nil, 30*time.Minute)
	if len(sessions) != 0 {
		t.Errorf("expected 0 sessions, got %d", len(sessions))
	}
}

func TestGroupSessionsCustomGap(t *testing.T) {
	// With 10-minute gap, the first session splits: 14:00-14:03 and 14:15-14:18
	sessions := GroupSessions(makeSessionTestEvents(), 10*time.Minute)

	if len(sessions) != 3 {
		t.Fatalf("expected 3 sessions with 10m gap, got %d", len(sessions))
	}
}

func TestActiveHoursUsesGroupSessions(t *testing.T) {
	timestamps := make([]string, 0)
	for _, e := range makeSessionTestEvents() {
		timestamps = append(timestamps, e.Timestamp)
	}

	hours := ActiveHours(timestamps, 30*time.Minute)

	// Session 1: 18 min, Session 2: 5 min = 23 min
	expectedMinutes := 23.0
	gotMinutes := hours * 60
	if diff := gotMinutes - expectedMinutes; diff < -0.01 || diff > 0.01 {
		t.Errorf("expected ~%.1f minutes, got %.1f", expectedMinutes, gotMinutes)
	}
}

func TestActiveHoursEmpty(t *testing.T) {
	hours := ActiveHours(nil, 30*time.Minute)
	if hours != 0 {
		t.Errorf("expected 0, got %f", hours)
	}
}

func TestRenderSessions(t *testing.T) {
	sessions := GroupSessions(makeSessionTestEvents(), 30*time.Minute)

	var buf bytes.Buffer
	if err := RenderSessions(&buf, sessions); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := buf.String()
	checks := []string{
		"claude-4.6-opus",
		"claude-4.6-sonnet",
		"18m",
		"5m",
		"$1.10",
		"$0.27",
		"TOTAL",
		"2 SESSIONS",
	}
	for _, check := range checks {
		if !strings.Contains(out, check) {
			t.Errorf("expected output to contain %q, got:\n%s", check, out)
		}
	}
}
