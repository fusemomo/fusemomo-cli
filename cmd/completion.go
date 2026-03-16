package cmd

import (
	"github.com/carapace-sh/carapace"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(completionCmd)
	// Register Carapace completions on the root command.
	carapace.Gen(rootCmd).FlagCompletion(carapace.ActionMap{})
	registerCompletions()
}

var completionCmd = &cobra.Command{
	Use:   "completion [bash|zsh|fish|powershell]",
	Short: "Generate shell completion script",
	Long: `Generate shell completion script for fusemomo.

Usage:
  # Zsh
  fusemomo completion zsh > ~/.zshrc.d/fusemomo.zsh

  # Bash
  fusemomo completion bash > /etc/bash_completion.d/fusemomo

  # Fish
  fusemomo completion fish > ~/.config/fish/completions/fusemomo.fish

  # PowerShell
  fusemomo completion powershell > fusemomo.ps1`,
	ValidArgs: []string{"bash", "zsh", "fish", "powershell"},
	Args:      cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		switch args[0] {
		case "bash":
			return rootCmd.GenBashCompletion(cmd.OutOrStdout())
		case "zsh":
			return rootCmd.GenZshCompletion(cmd.OutOrStdout())
		case "fish":
			return rootCmd.GenFishCompletion(cmd.OutOrStdout(), true)
		case "powershell":
			return rootCmd.GenPowerShellCompletionWithDesc(cmd.OutOrStdout())
		}
		return nil
	},
}

// registerCompletions wires Carapace value completions for specific flags.
func registerCompletions() {
	// interaction log --outcome
	carapace.Gen(interactionLogCmd).FlagCompletion(carapace.ActionMap{
		"outcome": carapace.ActionValues("success", "failed", "pending", "ignored", "unknown"),
	})

	// entity link --strategy
	carapace.Gen(entityLinkCmd).FlagCompletion(carapace.ActionMap{
		"strategy": carapace.ActionValues("deterministic", "probabilistic"),
	})

	// Global --output flag on root.
	carapace.Gen(rootCmd).FlagCompletion(carapace.ActionMap{
		"output": carapace.ActionValues("json", "table"),
	})
}

