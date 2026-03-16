package cmd

import (
	"fmt"

	"github.com/fusemomo/fusemomo-cli/internal/output"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print build version, commit, and build date",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Print(logo)
		return output.JSON(map[string]string{
			"version":  Version,
			"commit":   Commit,
			"built_at": BuiltAt,
		})
	},
}
