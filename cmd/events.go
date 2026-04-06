package cmd

import (
	"fmt"
	"strconv"
	"time"

	"github.com/dmwyatt/cursor-usage/internal/api"
	"github.com/dmwyatt/cursor-usage/internal/dateparse"
	"github.com/dmwyatt/cursor-usage/internal/output"
	"github.com/spf13/cobra"
)

var (
	eventsSince     string
	eventsUntil     string
	eventsStartDate string
	eventsEndDate   string
	eventsModel     string
	eventsPage      int
	eventsPageSize  int
	eventsAll       bool
	eventsAllTime   bool
	eventsAggregate bool
)

var eventsCmd = &cobra.Command{
	Use:   "events",
	Short: "List usage events with filtering and pagination",
	Long: `List usage events from the Cursor dashboard API.

By default, events are scoped to the current billing cycle. Use --all-time
to query across all billing periods, or --since/--until to set a custom range.

Date filtering supports human-friendly formats:
  --since 7d          (7 days ago)
  --since yesterday
  --since 2026-04-01
  --until today

Raw millisecond timestamps are also supported for scripting:
  --start-date 1774846800000
  --end-date 1775451599999

If --since or --start-date is provided, it overrides the billing cycle default.

Use --aggregate to show cost and token totals grouped by model.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		req, billingCycle, err := buildEventsRequest()
		if err != nil {
			return err
		}

		if eventsAll || eventsAggregate || billingCycle {
			return fetchAllEvents(cmd, req)
		}

		resp, err := apiClient.GetFilteredUsageEvents(req)
		if err != nil {
			return err
		}

		w := cmd.OutOrStdout()
		if jsonOutput {
			return output.RenderJSON(w, resp)
		}
		return output.RenderEvents(w, resp)
	},
}

// buildEventsRequest returns the API request and whether billing cycle mode is active.
func buildEventsRequest() (api.EventsRequest, bool, error) {
	req := api.EventsRequest{
		Page:     eventsPage,
		PageSize: eventsPageSize,
	}
	billingCycle := false

	now := time.Now()

	// Raw flags take precedence, then human-friendly, then billing cycle default
	if eventsStartDate != "" {
		req.StartDate = eventsStartDate
	} else if eventsSince != "" {
		ms, err := dateparse.ToMillis(eventsSince, now)
		if err != nil {
			return req, false, fmt.Errorf("parsing --since: %w", err)
		}
		req.StartDate = ms
	} else if !eventsAllTime {
		ms, err := fetchBillingCycleStart()
		if err != nil {
			return req, false, fmt.Errorf("fetching billing cycle: %w", err)
		}
		req.StartDate = ms
		billingCycle = true
	}

	if eventsEndDate != "" {
		req.EndDate = eventsEndDate
	} else if eventsUntil != "" {
		ms, err := dateparse.EndOfDayMillis(eventsUntil, now)
		if err != nil {
			return req, false, fmt.Errorf("parsing --until: %w", err)
		}
		req.EndDate = ms
	}

	return req, billingCycle, nil
}

func fetchBillingCycleStart() (string, error) {
	summary, err := apiClient.GetUsageSummary()
	if err != nil {
		return "", err
	}

	t, err := time.Parse(time.RFC3339Nano, summary.BillingCycleStart)
	if err != nil {
		return "", fmt.Errorf("parsing billing cycle start %q: %w", summary.BillingCycleStart, err)
	}

	return strconv.FormatInt(t.UnixMilli(), 10), nil
}

func fetchAllEvents(cmd *cobra.Command, req api.EventsRequest) error {
	var allEvents []api.UsageEvent
	var totalCount int
	req.Page = 1

	for {
		resp, err := apiClient.GetFilteredUsageEvents(req)
		if err != nil {
			return err
		}

		totalCount = resp.TotalUsageEventsCount
		allEvents = append(allEvents, resp.UsageEventsDisplay...)

		if len(allEvents) >= totalCount || len(resp.UsageEventsDisplay) == 0 {
			break
		}

		req.Page++
		time.Sleep(200 * time.Millisecond)
	}

	combined := &api.EventsResponse{
		TotalUsageEventsCount: totalCount,
		UsageEventsDisplay:    allEvents,
	}

	w := cmd.OutOrStdout()

	if eventsAggregate {
		agg := output.Aggregate(combined)
		if jsonOutput {
			return output.RenderJSON(w, agg)
		}
		return output.RenderAggregate(w, agg)
	}

	if jsonOutput {
		return output.RenderJSON(w, combined)
	}
	return output.RenderEvents(w, combined)
}

func init() {
	eventsCmd.Flags().StringVar(&eventsSince, "since", "", "start date (e.g., 7d, yesterday, 2026-04-01)")
	eventsCmd.Flags().StringVar(&eventsUntil, "until", "", "end date (e.g., today, 2026-04-05)")
	eventsCmd.Flags().StringVar(&eventsStartDate, "start-date", "", "start date as Unix ms timestamp (overrides --since)")
	eventsCmd.Flags().StringVar(&eventsEndDate, "end-date", "", "end date as Unix ms timestamp (overrides --until)")
	eventsCmd.Flags().StringVar(&eventsModel, "model", "", "filter by model name")
	eventsCmd.Flags().IntVar(&eventsPage, "page", 1, "page number (1-based)")
	eventsCmd.Flags().IntVar(&eventsPageSize, "page-size", 50, "events per page")
	eventsCmd.Flags().BoolVar(&eventsAll, "all", false, "fetch all pages (may be slow)")
	eventsCmd.Flags().BoolVar(&eventsAllTime, "all-time", false, "query across all billing periods instead of just the current one")
	eventsCmd.Flags().BoolVar(&eventsAggregate, "aggregate", false, "show aggregated totals by model (implies --all)")
	rootCmd.AddCommand(eventsCmd)
}
