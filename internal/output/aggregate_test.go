package output

import (
	"bytes"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/dmwyatt/cursor-usage/internal/api"
)

// makeTestEvents returns 4 events across two sessions:
//   Session 1: 14:00, 14:05, 14:20 (20 min)
//   Session 2: 16:00 (single event, 0 min)
//   Total active: 20 min = 1/3 hour
func makeTestEvents() *api.EventsResponse {
	return &api.EventsResponse{
		TotalUsageEventsCount: 4,
		UsageEventsDisplay: []api.UsageEvent{
			{
				Timestamp:      fmt.Sprintf("%d", time.Date(2026, 4, 5, 14, 0, 0, 0, time.UTC).UnixMilli()),
				Model:          "claude-4.6-opus",
				Kind:           "USAGE_EVENT_KIND_USAGE_BASED",
				ChargedCents:   100.0,
				UsageBasedCosts: "$1.00",
				TokenUsage:     api.TokenUsage{InputTokens: 500, OutputTokens: 2000, CacheWriteTokens: 1000, TotalCents: 95.0},
				CursorTokenFee: 5.0,
				IsHeadless:     false,
			},
			{
				Timestamp:      fmt.Sprintf("%d", time.Date(2026, 4, 5, 14, 5, 0, 0, time.UTC).UnixMilli()),
				Model:          "claude-4.6-opus",
				Kind:           "USAGE_EVENT_KIND_USAGE_BASED",
				ChargedCents:   50.0,
				UsageBasedCosts: "$0.50",
				TokenUsage:     api.TokenUsage{InputTokens: 300, OutputTokens: 1000, CacheWriteTokens: 500, TotalCents: 47.0},
				CursorTokenFee: 3.0,
				IsHeadless:     true,
			},
			{
				Timestamp:      fmt.Sprintf("%d", time.Date(2026, 4, 5, 14, 20, 0, 0, time.UTC).UnixMilli()),
				Model:          "claude-4.6-sonnet",
				Kind:           "USAGE_EVENT_KIND_INCLUDED_IN_BUSINESS",
				ChargedCents:   0,
				UsageBasedCosts: "$0.00",
				TokenUsage:     api.TokenUsage{InputTokens: 200, OutputTokens: 800, CacheWriteTokens: 0, TotalCents: 0},
				CursorTokenFee: 0,
				IsHeadless:     false,
			},
			{
				Timestamp:      fmt.Sprintf("%d", time.Date(2026, 4, 5, 16, 0, 0, 0, time.UTC).UnixMilli()),
				Model:          "claude-4.6-sonnet",
				Kind:           "USAGE_EVENT_KIND_USAGE_BASED",
				ChargedCents:   25.0,
				UsageBasedCosts: "$0.25",
				TokenUsage:     api.TokenUsage{InputTokens: 100, OutputTokens: 400, CacheWriteTokens: 200, TotalCents: 23.0},
				CursorTokenFee: 2.0,
				IsHeadless:     false,
			},
		},
	}
}

func TestAggregateTotals(t *testing.T) {
	agg := Aggregate(makeTestEvents(), 30*time.Minute)

	if agg.TotalEvents != 4 {
		t.Errorf("expected 4 total events, got %d", agg.TotalEvents)
	}
	if agg.TotalChargedCents != 175.0 {
		t.Errorf("expected 175.0 total charged cents, got %f", agg.TotalChargedCents)
	}
	if agg.TotalInputTokens != 1100 {
		t.Errorf("expected 1100 total input tokens, got %d", agg.TotalInputTokens)
	}
	if agg.TotalOutputTokens != 4200 {
		t.Errorf("expected 4200 total output tokens, got %d", agg.TotalOutputTokens)
	}
	if agg.TotalCacheWriteTokens != 1700 {
		t.Errorf("expected 1700 total cache write tokens, got %d", agg.TotalCacheWriteTokens)
	}
}

func TestAggregateActiveHours(t *testing.T) {
	agg := Aggregate(makeTestEvents(), 30*time.Minute)

	// Session 1: 14:00 to 14:20 = 20 min. Session 2: 16:00 alone = 0.
	// Total: 20 min = 1/3 hour
	expectedHours := 20.0 / 60.0
	if diff := agg.ActiveHours - expectedHours; diff < -0.01 || diff > 0.01 {
		t.Errorf("expected ~%.4f active hours, got %.4f", expectedHours, agg.ActiveHours)
	}

	// Cost/hr: $1.75 total / (1/3 hr) = $5.25/hr
	expectedCostPerHr := 1.75 / expectedHours
	if diff := agg.CostPerActiveHour - expectedCostPerHr; diff < -0.1 || diff > 0.1 {
		t.Errorf("expected ~$%.2f/hr, got $%.2f/hr", expectedCostPerHr, agg.CostPerActiveHour)
	}

	if agg.SessionGapMinutes != 30 {
		t.Errorf("expected session gap 30, got %d", agg.SessionGapMinutes)
	}
}

func TestAggregateByModel(t *testing.T) {
	agg := Aggregate(makeTestEvents(), 30*time.Minute)

	if len(agg.ByModel) != 2 {
		t.Fatalf("expected 2 models, got %d", len(agg.ByModel))
	}

	var opus, sonnet *ModelAggregate
	for i := range agg.ByModel {
		switch agg.ByModel[i].Model {
		case "claude-4.6-opus":
			opus = &agg.ByModel[i]
		case "claude-4.6-sonnet":
			sonnet = &agg.ByModel[i]
		}
	}

	if opus == nil || sonnet == nil {
		t.Fatal("missing expected model in aggregation")
	}

	if opus.Events != 2 {
		t.Errorf("opus: expected 2 events, got %d", opus.Events)
	}
	if opus.ChargedCents != 150.0 {
		t.Errorf("opus: expected 150.0 charged cents, got %f", opus.ChargedCents)
	}
	if opus.OutputTokens != 3000 {
		t.Errorf("opus: expected 3000 output tokens, got %d", opus.OutputTokens)
	}
	if opus.HeadlessEvents != 1 {
		t.Errorf("opus: expected 1 headless event, got %d", opus.HeadlessEvents)
	}

	if sonnet.Events != 2 {
		t.Errorf("sonnet: expected 2 events, got %d", sonnet.Events)
	}
	if sonnet.ChargedCents != 25.0 {
		t.Errorf("sonnet: expected 25.0 charged cents, got %f", sonnet.ChargedCents)
	}
}

func TestAggregateByKind(t *testing.T) {
	agg := Aggregate(makeTestEvents(), 30*time.Minute)

	if agg.UsageBasedEvents != 3 {
		t.Errorf("expected 3 usage-based events, got %d", agg.UsageBasedEvents)
	}
	if agg.IncludedEvents != 1 {
		t.Errorf("expected 1 included event, got %d", agg.IncludedEvents)
	}
	if agg.HeadlessEvents != 1 {
		t.Errorf("expected 1 headless event, got %d", agg.HeadlessEvents)
	}
}

func TestAggregateEmpty(t *testing.T) {
	agg := Aggregate(&api.EventsResponse{
		TotalUsageEventsCount: 0,
		UsageEventsDisplay:    nil,
	}, 30*time.Minute)

	if agg.TotalEvents != 0 {
		t.Errorf("expected 0 events, got %d", agg.TotalEvents)
	}
	if len(agg.ByModel) != 0 {
		t.Errorf("expected 0 models, got %d", len(agg.ByModel))
	}
	if agg.ActiveHours != 0 {
		t.Errorf("expected 0 active hours, got %f", agg.ActiveHours)
	}
	if agg.CostPerActiveHour != 0 {
		t.Errorf("expected 0 cost/hr, got %f", agg.CostPerActiveHour)
	}
}

func TestRenderAggregate(t *testing.T) {
	agg := Aggregate(makeTestEvents(), 30*time.Minute)

	var buf bytes.Buffer
	if err := RenderAggregate(&buf, agg); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := buf.String()
	checks := []string{
		"claude-4.6-opus",
		"claude-4.6-sonnet",
		"$1.75",
		"4200",
		"Active time:",
		"/hr",
	}
	for _, check := range checks {
		if !strings.Contains(out, check) {
			t.Errorf("expected output to contain %q, got:\n%s", check, out)
		}
	}
}
