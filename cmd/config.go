package cmd

import (
	"fmt"

	"github.com/joshuap/frontapp-cli/internal/config"
	"github.com/joshuap/frontapp-cli/internal/envelope"
	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage CLI configuration",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			envelope.FprintError(cmd.OutOrStdout(), "front config", err.Error(), "CONFIG_ERROR", "Check config file syntax", nil)
			return ErrPrinted
		}

		p, _ := config.Path()

		envelope.FprintResult(cmd.OutOrStdout(), "front config", ConfigResult{
			Path:         p,
			TokenCommand: cfg.TokenCommand,
			User:         cfg.User,
		}, []envelope.Action{
			{
				Command:     "front config set <key> <value>",
				Description: "Set a config value",
				Params: map[string]envelope.ParamSpec{
					"key":   {Description: "Config key (token_command, user)", Required: true},
					"value": {Description: "Config value", Required: true},
				},
			},
			{
				Command:     "front config path",
				Description: "Print config file path",
			},
		})
		return nil
	},
}

var configSetCmd = &cobra.Command{
	Use:   "set <key> <value>",
	Short: "Set a config value",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		key, value := args[0], args[1]

		validKeys := map[string]bool{"token_command": true, "user": true}
		if !validKeys[key] {
			envelope.FprintError(cmd.OutOrStdout(), "front config set", fmt.Sprintf("unknown config key %q", key), "CLI_ERROR", "Valid keys: token_command, user", nil)
			return ErrPrinted
		}

		cfg, err := config.Load()
		if err != nil {
			envelope.FprintError(cmd.OutOrStdout(), "front config set", err.Error(), "CONFIG_ERROR", "Check config file syntax", nil)
			return ErrPrinted
		}

		switch key {
		case "token_command":
			cfg.TokenCommand = value
		case "user":
			cfg.User = value
		}

		if err := config.Save(cfg); err != nil {
			envelope.FprintError(cmd.OutOrStdout(), "front config set", err.Error(), "CONFIG_ERROR", "Check file permissions", nil)
			return ErrPrinted
		}

		p, _ := config.Path()
		envelope.FprintResult(cmd.OutOrStdout(), "front config set", ConfigSetResult{
			Key:   key,
			Value: value,
			Path:  p,
		}, nil)
		return nil
	},
}

var configPathCmd = &cobra.Command{
	Use:   "path",
	Short: "Print config file path",
	RunE: func(cmd *cobra.Command, args []string) error {
		p, err := config.Path()
		if err != nil {
			envelope.FprintError(cmd.OutOrStdout(), "front config path", err.Error(), "CONFIG_ERROR", "", nil)
			return ErrPrinted
		}

		envelope.FprintResult(cmd.OutOrStdout(), "front config path", ConfigPathResult{
			Path: p,
		}, nil)
		return nil
	},
}

func init() {
	configCmd.AddCommand(configSetCmd)
	configCmd.AddCommand(configPathCmd)
	rootCmd.AddCommand(configCmd)
}
