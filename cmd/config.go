package cmd

import (
	"github.com/joshuap/frontapp-cli/internal/config"
	"github.com/joshuap/frontapp-cli/internal/envelope"
	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Show CLI configuration",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			envelope.FprintError(cmd.OutOrStdout(), "front config", err.Error(), "CONFIG_ERROR", "Check config file syntax", nil)
			return ErrPrinted
		}

		p, _ := config.Path()

		tokenStatus := "(not configured)"
		if len(cfg.TokenCommand) > 0 {
			tokenStatus = "(configured)"
		}

		envelope.FprintResult(cmd.OutOrStdout(), "front config", ConfigResult{
			Path:         p,
			TokenCommand: tokenStatus,
			User:         cfg.User,
		}, nil)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(configCmd)
}
