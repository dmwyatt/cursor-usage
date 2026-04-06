package output

import (
	"fmt"
	"io"
	"sort"
	"time"

	"github.com/dmwyatt/cursor-usage/internal/api"
	"github.com/jedib0t/go-pretty/v6/table"
)

// AggregateResult holds computed totals for a set of usage events.
type AggregateResult struct {
	TotalEvents           int              `json:"totalEvents"`
	TotalChargedCents     float64          `json:"totalChargedCents"`
	TotalInputTokens      int              `json:"totalInputTokens"`
	TotalOutputTokens     int              `json:"totalOutputTokens"`
	TotalCacheWriteTokens int              `json:"totalCacheWriteTokens"`
	UsageBasedEvents      int              `json:"usageBasedEvents"`
	IncludedEvents        int              `json:"includedEvents"`
	HeadlessEvents        int              `json:"headlessEvents"`
	ActiveHours           float64          `json:"activeHours"`
	CostPerActiveHour     float64          `json:"costPerActiveHour"`
	SessionGapMinutes     int              `json:"sessionGapMinutes"`
	ByModel               []ModelAggregate `json:"byModel"`
}

// ModelAggregate holds computed totals for a single model.
type ModelAggregate struct {
	Model            string  `json:"model"`
	Events           int     `json:"events"`
	ChargedCents     float64 `json:"chargedCents"`
	InputTokens      int     `json:"inputTokens"`
	OutputTokens     int     `json:"outputTokens"`
	CacheWriteTokens int     `json:"cacheWriteTokens"`
	HeadlessEvents   int     `json:"headlessEvents"`
}

// Aggregate computes totals from a set of usage events.
func Aggregate(resp *api.EventsResponse, sessionGap time.Duration) *AggregateResult {
	result := &AggregateResult{}
	models := map[string]*ModelAggregate{}
	timestamps := make([]string, 0, len(resp.UsageEventsDisplay))

	for _, e := range resp.UsageEventsDisplay {
		result.TotalEvents++
		result.TotalChargedCents += e.ChargedCents
		result.TotalInputTokens += e.TokenUsage.InputTokens
		result.TotalOutputTokens += e.TokenUsage.OutputTokens
		result.TotalCacheWriteTokens += e.TokenUsage.CacheWriteTokens
		timestamps = append(timestamps, e.Timestamp)

		switch e.Kind {
		case "USAGE_EVENT_KIND_USAGE_BASED":
			result.UsageBasedEvents++
		case "USAGE_EVENT_KIND_INCLUDED_IN_BUSINESS":
			result.IncludedEvents++
		}

		if e.IsHeadless {
			result.HeadlessEvents++
		}

		m, ok := models[e.Model]
		if !ok {
			m = &ModelAggregate{Model: e.Model}
			models[e.Model] = m
		}
		m.Events++
		m.ChargedCents += e.ChargedCents
		m.InputTokens += e.TokenUsage.InputTokens
		m.OutputTokens += e.TokenUsage.OutputTokens
		m.CacheWriteTokens += e.TokenUsage.CacheWriteTokens
		if e.IsHeadless {
			m.HeadlessEvents++
		}
	}

	for _, m := range models {
		result.ByModel = append(result.ByModel, *m)
	}
	sort.Slice(result.ByModel, func(i, j int) bool {
		return result.ByModel[i].ChargedCents > result.ByModel[j].ChargedCents
	})

	result.SessionGapMinutes = int(sessionGap.Minutes())
	result.ActiveHours = ActiveHours(timestamps, sessionGap)
	if result.ActiveHours > 0 {
		result.CostPerActiveHour = (result.TotalChargedCents / 100) / result.ActiveHours
	}

	return result
}

// RenderAggregate writes a human-readable aggregate summary to w.
func RenderAggregate(w io.Writer, agg *AggregateResult) error {
	fmt.Fprintf(w, "Total events: %d (usage-based: %d, included: %d, headless: %d)\n",
		agg.TotalEvents, agg.UsageBasedEvents, agg.IncludedEvents, agg.HeadlessEvents)
	fmt.Fprintf(w, "Total cost:   $%.2f\n", agg.TotalChargedCents/100)
	fmt.Fprintf(w, "Total tokens: %d input, %d output, %d cache write\n",
		agg.TotalInputTokens, agg.TotalOutputTokens, agg.TotalCacheWriteTokens)

	if agg.ActiveHours > 0 {
		fmt.Fprintf(w, "Active time:  %.1fh ($%.2f/hr, sessions split by %dm+ gaps)\n",
			agg.ActiveHours, agg.CostPerActiveHour, agg.SessionGapMinutes)
	}
	fmt.Fprintln(w)

	t := table.NewWriter()
	t.SetOutputMirror(w)
	t.SetStyle(table.StyleLight)

	t.AppendHeader(table.Row{
		"Model", "Events", "Cost", "Input Tok", "Output Tok", "Cache Write Tok", "Headless",
	})

	for _, m := range agg.ByModel {
		t.AppendRow(table.Row{
			m.Model,
			m.Events,
			fmt.Sprintf("$%.2f", m.ChargedCents/100),
			m.InputTokens,
			m.OutputTokens,
			m.CacheWriteTokens,
			m.HeadlessEvents,
		})
	}

	t.Render()
	return nil
}
