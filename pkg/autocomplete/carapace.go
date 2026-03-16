package autocomplete

import (
	"github.com/carapace-sh/carapace"
	"github.com/spf13/cobra"
)

// Register wires Carapace value completions onto the provided commands.
// Called from cmd/completion.go after all commands are registered.
func Register(
	rootCmd *cobra.Command,
	interactionLogCmd *cobra.Command,
	entityLinkCmd *cobra.Command,
) {
	// Global --output on root.
	carapace.Gen(rootCmd).FlagCompletion(carapace.ActionMap{
		"output": carapace.ActionValues("json", "table"),
	})

	// interaction log --outcome.
	carapace.Gen(interactionLogCmd).FlagCompletion(carapace.ActionMap{
		"outcome": carapace.ActionValues("success", "failed", "pending", "ignored", "unknown"),
	})

	// entity link --strategy.
	carapace.Gen(entityLinkCmd).FlagCompletion(carapace.ActionMap{
		"strategy": carapace.ActionValues("deterministic", "probabilistic"),
	})
}
