package output

import (
	"bytes"
	"strings"
	"testing"

	"github.com/dmwyatt/cursor-usage/internal/api"
)

func TestRenderSummary(t *testing.T) {
	summary := &api.UsageSummary{
		BillingCycleStart: "2026-04-02T14:11:55.000Z",
		BillingCycleEnd:   "2026-05-02T14:11:55.000Z",
		MembershipType:    "enterprise",
		IndividualUsage: api.IndividualUsage{
			Plan: api.PlanUsage{
				Used:  2000,
				Limit: 2000,
				Breakdown: api.PlanBreakdown{
					Included: 2000,
					Bonus:    6121,
					Total:    8121,
				},
				TotalPercentUsed: 100,
			},
			OnDemand: api.OnDemandUsage{
				Enabled: true,
				Used:    2309,
			},
		},
	}

	var buf bytes.Buffer
	if err := RenderSummary(&buf, summary); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()

	checks := []string{
		"enterprise",
		"2000",
		"2309",
		"Billing",
	}
	for _, check := range checks {
		if !strings.Contains(output, check) {
			t.Errorf("expected output to contain %q, got:\n%s", check, output)
		}
	}
}

func TestRenderEvents(t *testing.T) {
	events := &api.EventsResponse{
		TotalUsageEventsCount: 2,
		UsageEventsDisplay: []api.UsageEvent{
			{
				Timestamp:    "1775418973898",
				Model:        "claude-4.6-opus",
				Kind:         "USAGE_EVENT_KIND_USAGE_BASED",
				ChargedCents: 124.73,
				TokenUsage: api.TokenUsage{
					InputTokens:  3,
					OutputTokens: 20525,
				},
				IsHeadless: false,
			},
			{
				Timestamp:    "1775418000000",
				Model:        "claude-4.6-sonnet",
				Kind:         "USAGE_EVENT_KIND_INCLUDED_IN_BUSINESS",
				ChargedCents: 0,
				TokenUsage: api.TokenUsage{
					InputTokens:  1500,
					OutputTokens: 3000,
				},
				IsHeadless: true,
			},
		},
	}

	var buf bytes.Buffer
	if err := RenderEvents(&buf, events); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()

	checks := []string{
		"claude-4.6-opus",
		"claude-4.6-sonnet",
		"20525",
		"Total events: 2",
	}
	for _, check := range checks {
		if !strings.Contains(output, check) {
			t.Errorf("expected output to contain %q, got:\n%s", check, output)
		}
	}
}
