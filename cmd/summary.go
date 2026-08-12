package cmd

import (
	"fmt"

	"github.com/dmwyatt/cursor-usage/internal/api"
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

		eventsResp, err := apiClient.GetFilteredUsageEvents(api.EventsRequest{
			Page:     1,
			PageSize: recentEventCount,
		})
		if err != nil {
			return err
		}

		w := cmd.OutOrStdout()
		if jsonOutput {
			return output.RenderJSON(w, struct {
				*api.UsageSummary
				RecentEvents []api.UsageEvent `json:"recentEvents"`
			}{summary, eventsResp.UsageEventsDisplay})
		}

		if err := output.RenderSummary(w, summary); err != nil {
			return err
		}
		if len(eventsResp.UsageEventsDisplay) == 0 {
			return nil
		}
		fmt.Fprintln(w)
		return output.RenderRecentEvents(w, eventsResp.UsageEventsDisplay)
	},
}

func init() {
	rootCmd.AddCommand(summaryCmd)
}
