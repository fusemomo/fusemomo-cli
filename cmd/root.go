// Package cmd contains all Cobra command definitions.
// Commands stay thin — all HTTP calls in internal/api, all output in internal/output.
package cmd

import (
	"fmt"
	"os"

	"github.com/fusemomo/fusemomo-cli/internal/api"
	"github.com/fusemomo/fusemomo-cli/internal/logging"
	"github.com/fusemomo/fusemomo-cli/internal/output"
	"github.com/fusemomo/fusemomo-cli/pkg/config"
	"github.com/spf13/cobra"
)

// buildVersion vars are set by main.go (injected from ldflags).
var (
	Version = "dev"
	Commit  = "none"
	BuiltAt = "unknown"
)

// cfg is the resolved, validated configuration — available to all subcommands.
var cfg *config.Config

// apiClient is the shared HTTP client — built once after config loads.
var apiClient *api.Client

// Commands that skip config validation.
var skipValidation = map[string]bool{
	"setup":      true,
	"version":    true,
	"completion": true,
	"help":       true,
}

var logo = `
███████╗██╗   ██╗███████╗███████╗███╗   ███╗ ██████╗ ███╗   ███╗ ██████╗ 
██╔════╝██║   ██║██╔════╝██╔════╝████╗ ████║██╔═══██╗████╗ ████║██╔═══██╗
█████╗  ██║   ██║███████╗█████╗  ██╔████╔██║██║   ██║██╔████╔██║██║   ██║
██╔══╝  ██║   ██║╚════██║██╔══╝  ██║╚██╔╝██║██║   ██║██║╚██╔╝██║██║   ██║
██║     ╚██████╔╝███████║███████╗██║ ╚═╝ ██║╚██████╔╝██║ ╚═╝ ██║╚██████╔╝
╚═╝      ╚═════╝ ╚══════╝╚══════╝╚═╝     ╚═╝ ╚═════╝ ╚═╝     ╚═╝ ╚═════╝ 

`

// rootCmd is the base `fusemomo` command.
var rootCmd = &cobra.Command{
	Use:   "fusemomo",
	Short: "FuseMomo CLI — Behavioral Entity Graph for AI agents",
	Long: logo + `Fusemomo is the command-line interface for the Fusemomo REST API.
Get started:
  fusemomo setup       		# Set up your API key
  fusemomo entity --help   	# Explore entity commands
  fusemomo version         	# Print build info`,
	SilenceErrors: true,
	SilenceUsage:  true,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// Initialise logger first.
		debug, _ := cmd.Flags().GetBool("debug")
		logging.Init(debug)

		// Skip config validation for utility commands.
		if skipValidation[cmd.Name()] {
			return nil
		}

		// Load and validate config.
		var err error
		cfg, err = config.Load(cmd.Root())
		if err != nil {
			output.WriteErrorMsg(3, "config_error", err.Error())
			os.Exit(3)
		}
		if err := cfg.Validate(); err != nil {
			if ve, ok := err.(*config.ValidationError); ok {
				output.WriteErrorMsg(ve.ExitCode, "config_error", ve.Message)
				os.Exit(ve.ExitCode)
			}
			output.WriteErrorMsg(3, "config_error", err.Error())
			os.Exit(3)
		}

		// Warn about non-TLS URLs.
		if cfg.IsHTTP() {
			fmt.Fprintln(os.Stderr, `{"warning":"api_url uses http:// instead of https:// — your API key may be exposed in transit"}`)
		}

		// Build the shared API client.
		api.Version = Version
		apiClient = api.NewClient(cfg.APIKey, cfg.APIURL, cfg.Timeout)
		return nil
	},
}

// Execute is called by main.go. It binds build-time variables and runs the root command.
func Execute(version, commit, builtAt string) {
	Version = version
	Commit = commit
	BuiltAt = builtAt

	if err := rootCmd.Execute(); err != nil {
		// Cobra errors are usually flag/usage errors.
		output.WriteErrorMsg(1, "usage_error", err.Error())
		os.Exit(1)
	}
}

func init() {
	// Global flags available on every subcommand.
	rootCmd.PersistentFlags().String("api-key", "", "FuseMomo API key (overrides FUSEMOMO_API_KEY and config file)")
	rootCmd.PersistentFlags().String("api-url", "https://api.fusemomo.com", "API base URL")
	rootCmd.PersistentFlags().String("output", "json", "Output format: json or table")
	rootCmd.PersistentFlags().Int("timeout", 30, "HTTP request timeout in seconds")
	rootCmd.PersistentFlags().Bool("debug", false, "Enable debug logging to stderr")
	rootCmd.PersistentFlags().Bool("dry-run", false, "Print request body without making the API call (mutating commands only)")
	rootCmd.PersistentFlags().Bool("no-color", false, "Disable ANSI color in table output")
}

// getOutputFormat returns the resolved --output flag value.
func getOutputFormat(cmd *cobra.Command) string {
	f, _ := cmd.Flags().GetString("output")
	if f == "" {
		f = "json"
	}
	return f
}

// isDryRun returns true when --dry-run is set.
func isDryRun(cmd *cobra.Command) bool {
	v, _ := cmd.Flags().GetBool("dry-run")
	return v
}

// isNoColor returns true when --no-color is set.
func isNoColor(cmd *cobra.Command) bool {
	v, _ := cmd.Flags().GetBool("no-color")
	return v
}

// handleCLIError writes a CLIError to stderr and exits with the correct code.
func handleCLIError(err error) {
	if cliErr, ok := err.(*api.CLIError); ok {
		code := output.WriteError(cliErr, false)
		os.Exit(code)
	}
	output.WriteErrorMsg(2, "unexpected_error", err.Error())
	os.Exit(2)
}
