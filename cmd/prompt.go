package cmd

import (
	"github.com/fusemomo/fusemomo-cli/internal/prompt"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(promptCmd)
}

var promptCmd = &cobra.Command{
	Use:   "prompt",
	Short: "Launch interactive REPL for exploratory use",
	Long:  "Launches an interactive command REPL with tab completion, command history, and suggestions. Type 'exit' or press Ctrl+D to quit.",
	RunE: func(cmd *cobra.Command, args []string) error {
		prompt.Run()
		return nil
	},
}
