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

var (
	summaryWatch    bool
	summaryInterval time.Duration
)

var summaryCmd = &cobra.Command{
	Use:   "summary",
	Short: "Show billing cycle summary, usage totals, and recent events",
	RunE: func(cmd *cobra.Command, args []string) error {
		if summaryWatch && jsonOutput {
			return fmt.Errorf("--watch cannot be used with --json")
		}
		if summaryWatch {
			return runSummaryWatch(summaryInterval)
		}

		summary, tokens, recent, mem, err := fetchSummaryDashboard()
		if err != nil {
			return err
		}

		w := cmd.OutOrStdout()
		if jsonOutput {
			return output.RenderJSON(w, struct {
				*api.UsageSummary
				TokenTotals  output.TokenTotals `json:"tokenTotals"`
				RecentEvents []api.UsageEvent   `json:"recentEvents"`
				Memory       *memstat.Stats     `json:"memory,omitempty"`
			}{summary, *tokens, recent, mem})
		}

		return output.RenderSummaryScreen(w, summary, tokens, recent, mem)
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
	summaryCmd.Flags().BoolVarP(&summaryWatch, "watch", "w", false, "refresh in place (prefer over watch(1) on phones)")
	summaryCmd.Flags().DurationVar(&summaryInterval, "interval", 5*time.Second, "refresh interval for --watch")
	rootCmd.AddCommand(summaryCmd)
}
