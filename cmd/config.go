package cmd

import (
	"fmt"

	"github.com/dmwyatt/cursor-usage/internal/config"
	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage CLI configuration",
}

var configSetTokenCmd = &cobra.Command{
	Use:   "set-token TOKEN",
	Short: "Save your Cursor session token",
	Long:  "Save the WorkosCursorSessionToken cookie value for API authentication.",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return fmt.Errorf("loading config: %w", err)
		}

		cfg.SessionToken = args[0]
		if err := config.Save(cfg); err != nil {
			return fmt.Errorf("saving config: %w", err)
		}

		fmt.Fprintf(cmd.OutOrStdout(), "Token saved to %s\n", config.DefaultPath())
		return nil
	},
}

var configShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show current configuration",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return fmt.Errorf("loading config: %w", err)
		}

		w := cmd.OutOrStdout()
		fmt.Fprintf(w, "Config file: %s\n", config.DefaultPath())

		if cfg.SessionToken == "" {
			fmt.Fprintln(w, "Token:       (not set)")
		} else {
			redacted := cfg.SessionToken[:4] + "..." + cfg.SessionToken[len(cfg.SessionToken)-4:]
			fmt.Fprintf(w, "Token:       %s\n", redacted)
		}

		if cfg.BaseURL != "" {
			fmt.Fprintf(w, "Base URL:    %s\n", cfg.BaseURL)
		}

		return nil
	},
}

var configPathCmd = &cobra.Command{
	Use:   "path",
	Short: "Print the config file path",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Fprintln(cmd.OutOrStdout(), config.DefaultPath())
	},
}

func init() {
	configCmd.AddCommand(configSetTokenCmd)
	configCmd.AddCommand(configShowCmd)
	configCmd.AddCommand(configPathCmd)
	rootCmd.AddCommand(configCmd)
}
