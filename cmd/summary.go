package cmd

import (
	"fmt"
	"strconv"
	"time"

	"github.com/dmwyatt/cursor-usage/internal/api"
	"github.com/dmwyatt/cursor-usage/internal/memstat"
	"github.com/dmwyatt/cursor-usage/internal/output"
	"github.com/spf13/cobra"
)

const recentEventCount = 5

var summaryCmd = &cobra.Command{
	Use:   "summary",
	Short: "Show billing cycle summary, usage totals, and recent events",
	RunE: func(cmd *cobra.Command, args []string) error {
		summary, err := apiClient.GetUsageSummary()
		if err != nil {
			return err
		}

		startMs, err := billingCycleStartMillis(summary.BillingCycleStart)
		if err != nil {
			return err
		}
		eventsResp, err := apiClient.GetAllFilteredUsageEvents(api.EventsRequest{
			StartDate: startMs,
			PageSize:  200,
		})
		if err != nil {
			return err
		}

		tokens := output.SumTokens(eventsResp.UsageEventsDisplay)
		recent := eventsResp.UsageEventsDisplay
		if len(recent) > recentEventCount {
			recent = recent[:recentEventCount]
		}

		mem, _ := memstat.Read()

		w := cmd.OutOrStdout()
		if jsonOutput {
			return output.RenderJSON(w, struct {
				*api.UsageSummary
				TokenTotals  output.TokenTotals `json:"tokenTotals"`
				RecentEvents []api.UsageEvent   `json:"recentEvents"`
				Memory       *memstat.Stats     `json:"memory,omitempty"`
			}{summary, tokens, recent, mem})
		}

		if err := output.RenderSummary(w, summary, &tokens); err != nil {
			return err
		}
		if mem != nil {
			fmt.Fprintln(w)
			if err := output.RenderMemory(w, mem); err != nil {
				return err
			}
		}
		if len(recent) == 0 {
			return nil
		}
		fmt.Fprintln(w)
		return output.RenderRecentEvents(w, recent)
	},
}

func billingCycleStartMillis(start string) (string, error) {
	t, err := time.Parse(time.RFC3339Nano, start)
	if err != nil {
		return "", fmt.Errorf("parsing billing cycle start %q: %w", start, err)
	}
	return strconv.FormatInt(t.UnixMilli(), 10), nil
}

func init() {
	rootCmd.AddCommand(summaryCmd)
}
