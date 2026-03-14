package cmd

import (
	"context"
	"errors"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/joshuap/frontapp-cli/internal/config"
	"github.com/joshuap/frontapp-cli/internal/envelope"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// Version is set at build time via -ldflags:
//
//	go build -ldflags "-X github.com/joshuap/frontapp-cli/cmd.Version=1.2.3"
var Version = "dev"

var rootCmd = &cobra.Command{
	Use:           "front",
	Short:         "CLI for the Front API",
	Long:          "CLI for the Front API. Returns JSON envelopes with result data and next_actions for agent-friendly navigation.",
	SilenceUsage:  true,
	SilenceErrors: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		var actions []envelope.Action
		collectLeafCommands(cmd, "front", &actions)

		envelope.FprintResult(cmd.OutOrStdout(), "front", RootResult{Version: Version}, actions)
		return nil
	},
}

// collectLeafCommands walks the command tree and collects leaf commands (those with RunE).
func collectLeafCommands(cmd *cobra.Command, prefix string, actions *[]envelope.Action) {
	for _, c := range cmd.Commands() {
		if c.Hidden {
			continue
		}
		fullName := prefix + " " + c.Name()
		if c.HasSubCommands() {
			collectLeafCommands(c, fullName, actions)
		} else if c.RunE != nil {
			*actions = append(*actions, buildAction(c, fullName))
		}
	}
}

// buildAction creates a fully-documented Action for a leaf command,
// including positional args and all flags as structured params.
func buildAction(c *cobra.Command, fullName string) envelope.Action {
	cmd, params := buildCommandUsageAndParams(c, fullName)
	action := envelope.Action{
		Command:     cmd,
		Description: c.Short,
	}
	if len(params) > 0 {
		action.Params = params
	}
	return action
}

// buildCommandUsageAndParams returns the command string with positional arg
// placeholders (no flag hints) and a params map documenting all flags and args.
func buildCommandUsageAndParams(c *cobra.Command, fullName string) (string, map[string]envelope.ParamSpec) {
	usage := fullName
	params := map[string]envelope.ParamSpec{}

	// Add positional arg placeholders (both <required> and [optional])
	parts := strings.Fields(c.Use)
	for _, p := range parts[1:] {
		if strings.HasPrefix(p, "<") && strings.HasSuffix(p, ">") {
			usage += " " + p
			name := strings.Trim(p, "<>")
			params[name] = envelope.ParamSpec{
				Description: name + " (required)",
				Required:    true,
			}
		} else if strings.HasPrefix(p, "[") && strings.HasSuffix(p, "]") {
			usage += " " + p
			name := strings.Trim(p, "[]")
			params[name] = envelope.ParamSpec{
				Description: name + " (optional)",
			}
		}
	}

	// Add all non-internal flags as structured params
	c.Flags().VisitAll(func(f *pflag.Flag) {
		if vals, ok := f.Annotations["internal"]; ok && len(vals) > 0 && vals[0] == "true" {
			return
		}
		spec := envelope.ParamSpec{
			Description: f.Usage,
		}
		zeroDefaults := map[string]string{"int": "0", "bool": "false", "uint": "0", "int64": "0", "float64": "0"}
		if f.DefValue != "" && f.DefValue != zeroDefaults[f.Value.Type()] {
			spec.Default = f.DefValue
		}
		if vals, ok := f.Annotations["enum"]; ok && len(vals) > 0 {
			spec.Enum = vals
		}
		params["--"+f.Name] = spec
	})

	if len(params) == 0 {
		return usage, nil
	}
	return usage, params
}


// commandRegistry maps "resource action" keys to their cobra.Command pointers.
// Populated in init() after all commands are registered.
var commandRegistry = map[string]*cobra.Command{}

// registerCommand adds a command to the registry for cross-command references.
func registerCommand(key string, c *cobra.Command) {
	commandRegistry[key] = c
}

// actionFor builds a next_action referencing a command by registry key (e.g. "conversations list"),
// with full flag documentation.
func actionFor(key string) envelope.Action {
	c, ok := commandRegistry[key]
	if !ok {
		return envelope.Action{Command: "front " + key, Description: key}
	}
	return buildAction(c, "front "+key)
}


func init() {
	rootCmd.CompletionOptions.HiddenDefaultCmd = true
	rootCmd.SetHelpCommand(&cobra.Command{Hidden: true})
	rootCmd.Version = Version
	rootCmd.SetVersionTemplate("{{.Version}}\n")
}

// ErrPrinted is returned by commands that have already written error JSON to stdout.
// Execute() uses this to exit with code 1 without printing again.
var ErrPrinted = errors.New("error already printed")

func Execute() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()
	rootCmd.SetContext(ctx)

	if err := rootCmd.Execute(); err != nil {
		if !errors.Is(err, ErrPrinted) {
			var nte noTokenError
			if errors.As(err, &nte) {
				fix := "Set FRONT_API_TOKEN or configure token_command in config file"
				if p, err := config.Path(); err == nil {
					fix = "Set FRONT_API_TOKEN or configure token_command in " + p
				}
				envelope.PrintError("front", err.Error(), "UNAUTHORIZED", fix, nil)
			} else {
				msg := err.Error()
				fix := ""
				if strings.Contains(msg, "required arg") || strings.Contains(msg, "accepts") {
					fix = "Run 'front' to see available commands and their required arguments"
				} else if strings.Contains(msg, "unknown command") {
					fix = "Run 'front' to see available commands"
				}
				envelope.PrintError("front", msg, "CLI_ERROR", fix, nil)
			}
		}
		os.Exit(1)
	}
}

