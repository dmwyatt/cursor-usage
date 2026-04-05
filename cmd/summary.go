package cmd

import (
	"github.com/dmwyatt/cursor-usage/internal/output"
	"github.com/spf13/cobra"
)

var summaryCmd = &cobra.Command{
	Use:   "summary",
	Short: "Show billing cycle summary and usage totals",
	RunE: func(cmd *cobra.Command, args []string) error {
		summary, err := apiClient.GetUsageSummary()
		if err != nil {
			return err
		}

		w := cmd.OutOrStdout()
		if jsonOutput {
			return output.RenderJSON(w, summary)
		}
		return output.RenderSummary(w, summary)
	},
}

func init() {
	rootCmd.AddCommand(summaryCmd)
}
