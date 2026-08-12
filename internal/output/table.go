package output

import (
	"fmt"
	"io"
	"strconv"
	"time"

	"github.com/dmwyatt/cursor-usage/internal/api"
	"github.com/jedib0t/go-pretty/v6/table"
)

// RenderSummary writes a human-readable usage summary table to w.
func RenderSummary(w io.Writer, s *api.UsageSummary) error {
	t := table.NewWriter()
	t.SetOutputMirror(w)
	t.SetStyle(table.StyleLight)

	t.SetTitle("Cursor Usage Summary")
	t.AppendHeader(table.Row{"Field", "Value"})

	t.AppendRow(table.Row{"Membership", s.MembershipType})
	t.AppendRow(table.Row{"Billing Start", formatTimestamp(s.BillingCycleStart)})
	t.AppendRow(table.Row{"Billing End", formatTimestamp(s.BillingCycleEnd)})
	t.AppendSeparator()

	plan := s.IndividualUsage.Plan
	t.AppendRow(table.Row{"Plan Used / Limit", fmt.Sprintf("%s / %s (%.1f%%)", formatCents(planConsumedCents(plan)), formatCents(float64(plan.Limit)), plan.TotalPercentUsed)})
	t.AppendRow(table.Row{"Auto", fmt.Sprintf("%.1f%%", plan.AutoPercentUsed)})
	t.AppendRow(table.Row{"API", fmt.Sprintf("%.1f%%", plan.APIPercentUsed)})
	t.AppendRow(table.Row{"Plan Included", formatCents(float64(plan.Breakdown.Included))})
	t.AppendRow(table.Row{"Plan Bonus", formatCents(float64(plan.Breakdown.Bonus))})
	t.AppendRow(table.Row{"Plan Total Allowance", formatCents(float64(plan.Breakdown.Total))})
	t.AppendSeparator()

	od := s.IndividualUsage.OnDemand
	if od.Enabled {
		limitStr := "unlimited"
		if od.Limit != nil {
			limitStr = formatCents(float64(*od.Limit))
		}
		t.AppendRow(table.Row{"On-Demand Used / Limit", fmt.Sprintf("%s / %s", formatCents(float64(od.Used)), limitStr)})
	} else {
		t.AppendRow(table.Row{"On-Demand", "disabled"})
	}

	t.Render()
	return nil
}

// RenderEvents writes a human-readable events table to w.
func RenderEvents(w io.Writer, resp *api.EventsResponse) error {
	fmt.Fprintf(w, "Total events: %d\n\n", resp.TotalUsageEventsCount)

	t := table.NewWriter()
	t.SetOutputMirror(w)
	t.SetStyle(table.StyleLight)

	t.AppendHeader(table.Row{
		"Time", "Model", "Kind", "Input Tok", "Output Tok", "Cost (cents)", "Headless",
	})

	for _, e := range resp.UsageEventsDisplay {
		t.AppendRow(table.Row{
			formatMsTimestamp(e.Timestamp),
			e.Model,
			shortKind(e.Kind),
			e.TokenUsage.InputTokens,
			e.TokenUsage.OutputTokens,
			fmt.Sprintf("%.2f", e.ChargedCents),
			boolStr(e.IsHeadless),
		})
	}

	t.Render()
	return nil
}

func formatTimestamp(iso string) string {
	t, err := time.Parse(time.RFC3339Nano, iso)
	if err != nil {
		return iso
	}
	return t.Local().Format("2006-01-02 15:04")
}

func formatMsTimestamp(ms string) string {
	n, err := strconv.ParseInt(ms, 10, 64)
	if err != nil {
		return ms
	}
	t := time.UnixMilli(n)
	return t.Local().Format("2006-01-02 15:04")
}

func shortKind(kind string) string {
	switch kind {
	case "USAGE_EVENT_KIND_USAGE_BASED":
		return "usage-based"
	case "USAGE_EVENT_KIND_INCLUDED_IN_BUSINESS":
		return "included"
	default:
		return kind
	}
}

func boolStr(b bool) string {
	if b {
		return "yes"
	}
	return "no"
}

// Plan used/limit/breakdown values are cents. The API often reports used==limit
// even when totalPercentUsed is well below 100, so consume from the percentage.
func planConsumedCents(plan api.PlanUsage) float64 {
	if plan.Limit > 0 {
		return float64(plan.Limit) * plan.TotalPercentUsed / 100
	}
	return float64(plan.Used)
}

func formatCents(cents float64) string {
	return fmt.Sprintf("$%.2f", cents/100)
}
