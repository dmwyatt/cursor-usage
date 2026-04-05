package cmd

import (
	"fmt"

	"github.com/dmwyatt/cursor-usage/internal/api"
	"github.com/dmwyatt/cursor-usage/internal/config"
	"github.com/spf13/cobra"
)

var (
	jsonOutput bool
	apiClient  *api.Client
)

var rootCmd = &cobra.Command{
	Use:   "cursor-usage",
	Short: "Track Cursor IDE usage and costs",
	Long:  "A CLI tool for querying Cursor IDE usage data via the unofficial dashboard API.",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// Skip client initialization for config commands
		if cmd.Parent() != nil && cmd.Parent().Name() == "config" {
			return nil
		}
		if cmd.Name() == "config" {
			return nil
		}

		cfg, err := config.Load()
		if err != nil {
			return fmt.Errorf("loading config: %w", err)
		}

		if cfg.SessionToken == "" {
			return fmt.Errorf("no session token configured; run: cursor-usage config set-token TOKEN")
		}

		opts := []api.Option{}
		if cfg.BaseURL != "" {
			opts = append(opts, api.WithBaseURL(cfg.BaseURL))
		}

		apiClient = api.NewClient(cfg.SessionToken, opts...)
		return nil
	},
}

func init() {
	rootCmd.PersistentFlags().BoolVar(&jsonOutput, "json", false, "output as JSON instead of a table")
}

func Execute() error {
	return rootCmd.Execute()
}
