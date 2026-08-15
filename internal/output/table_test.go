package output

import (
	"bytes"
	"strings"
	"testing"

	"github.com/dmwyatt/cursor-usage/internal/api"
	"github.com/dmwyatt/cursor-usage/internal/memstat"
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
	if err := RenderSummary(&buf, summary, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()

	checks := []string{
		"enterprise",
		"$20.00",
		"$23.09",
		"Billing",
	}
	for _, check := range checks {
		if !strings.Contains(output, check) {
			t.Errorf("expected output to contain %q, got:\n%s", check, output)
		}
	}
}

func TestRenderSummaryTokens(t *testing.T) {
	summary := &api.UsageSummary{
		BillingCycleStart: "2026-08-06T02:46:21.000Z",
		BillingCycleEnd:   "2026-09-06T02:46:21.000Z",
		MembershipType:    "pro",
	}
	tokens := &TokenTotals{Input: 7_200_000, Output: 771_000, CacheRead: 140_000_000, CacheWrite: 209_300, Total: 148_180_300}

	var buf bytes.Buffer
	if err := RenderSummary(&buf, summary, tokens); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got := buf.String()
	for _, check := range []string{"Tokens", "148.2M", "7.2M", "771.0K", "140.0M", "Cache R / W", "In / Out"} {
		if !strings.Contains(got, check) {
			t.Errorf("expected output to contain %q, got:\n%s", check, got)
		}
	}
}

func TestRenderSummaryUsesPercentWhenUsedIsCapped(t *testing.T) {
	summary := &api.UsageSummary{
		BillingCycleStart: "2026-08-06T02:46:21.000Z",
		BillingCycleEnd:   "2026-09-06T02:46:21.000Z",
		MembershipType:    "pro",
		IndividualUsage: api.IndividualUsage{
			Plan: api.PlanUsage{
				Used:      2000,
				Limit:     2000,
				Remaining: 0,
				Breakdown: api.PlanBreakdown{
					Included: 2000,
					Bonus:    5780,
					Total:    7780,
				},
				AutoPercentUsed:  25.933333333333337,
				APIPercentUsed:   0,
				TotalPercentUsed: 22.55072463768116,
			},
		},
	}

	var buf bytes.Buffer
	if err := RenderSummary(&buf, summary, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	checks := []string{
		"$4.51 / $20.00 (22.6%)",
		"25.9% / 0.0%",
		"$57.80",
		"$77.80",
	}
	for _, check := range checks {
		if !strings.Contains(output, check) {
			t.Errorf("expected output to contain %q, got:\n%s", check, output)
		}
	}
	if strings.Contains(output, "2000 / 2000") {
		t.Errorf("should not print capped used/limit as request counts, got:\n%s", output)
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

func TestRenderRecentEvents(t *testing.T) {
	events := []api.UsageEvent{
		{
			Timestamp:    "1786532478780",
			Model:        "default",
			Kind:         "USAGE_EVENT_KIND_INCLUDED_IN_PRO",
			ChargedCents: 6.87,
			TokenUsage: api.TokenUsage{
				InputTokens:  10992,
				OutputTokens: 867,
			},
		},
		{
			Timestamp:    "1786532122766",
			Model:        "composer-2.5-fast",
			Kind:         "USAGE_EVENT_KIND_USAGE_BASED",
			ChargedCents: 3.88,
			TokenUsage: api.TokenUsage{
				InputTokens:  1309,
				OutputTokens: 947,
			},
			IsHeadless: true,
		},
	}

	var buf bytes.Buffer
	if err := RenderRecentEvents(&buf, events); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	checks := []string{
		"Recent Events",
		"default",
		"included",
		"composer-2.5-fast",
		"6.87¢",
	}
	for _, check := range checks {
		if !strings.Contains(output, check) {
			t.Errorf("expected output to contain %q, got:\n%s", check, output)
		}
	}
	if strings.Contains(output, "Total events:") {
		t.Errorf("recent events table should not print total count, got:\n%s", output)
	}
	if strings.Contains(output, "USAGE_EVENT_KIND_INCLUDED_IN_PRO") {
		t.Errorf("expected short kind, got:\n%s", output)
	}
	if strings.Contains(output, "INPUT TOK") || strings.Contains(output, "10992") {
		t.Errorf("summary recent events should omit token columns, got:\n%s", output)
	}
}

func TestRenderMemory(t *testing.T) {
	stats := &memstat.Stats{
		Physical:   8589934592,
		Used:       7288659968,
		App:        2348810240,
		Wired:      1962934272,
		Compressed: 2976915456,
		Cached:     1254006784,
		SwapUsed:   743545242,
		Pressure:   "green",
	}

	var buf bytes.Buffer
	if err := RenderMemory(&buf, stats); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	checks := []string{
		"Used / Phys",
		"8.00 GB",
		"App/Wire/Compr",
		"Cached / Swap",
		"709.1 MB",
		"Pressure",
		"green",
	}
	for _, check := range checks {
		if !strings.Contains(output, check) {
			t.Errorf("expected output to contain %q, got:\n%s", check, output)
		}
	}
}
