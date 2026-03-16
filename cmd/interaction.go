package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/fusemomo/fusemomo-cli/internal/api"
	"github.com/fusemomo/fusemomo-cli/internal/output"
	"github.com/fusemomo/fusemomo-cli/internal/tui"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(interactionCmd)
	interactionCmd.AddCommand(interactionLogCmd)
	interactionCmd.AddCommand(interactionBatchCmd)

	// interaction log flags.
	interactionLogCmd.Flags().String("entity-id", "", "UUID of the entity (required)")
	interactionLogCmd.Flags().String("api", "", "API source e.g. sendgrid, stripe (required)")
	interactionLogCmd.Flags().String("action-type", "", "Action category e.g. send_email (required)")
	interactionLogCmd.Flags().String("action", "", "Specific action e.g. payment_reminder_v3 (required)")
	interactionLogCmd.Flags().String("outcome", "", "Result: success|failed|pending|ignored|unknown (required)")
	interactionLogCmd.Flags().String("intent", "", "Business goal e.g. payment_recovery")
	interactionLogCmd.Flags().String("agent-id", "", "Agent identifier")
	interactionLogCmd.Flags().String("external-ref", "", "External reference ID")
	interactionLogCmd.Flags().String("meta", "", "Additional metadata as JSON string")
	interactionLogCmd.Flags().String("occurred-at", "", "RFC3339 timestamp (defaults to now)")
	interactionLogCmd.MarkFlagRequired("entity-id")
	interactionLogCmd.MarkFlagRequired("api")
	interactionLogCmd.MarkFlagRequired("action-type")
	interactionLogCmd.MarkFlagRequired("action")
	interactionLogCmd.MarkFlagRequired("outcome")

	// interaction batch flags.
	interactionBatchCmd.Flags().String("file", "", "Path to JSON file with interactions array (required)")
	interactionBatchCmd.MarkFlagRequired("file")
}

var interactionCmd = &cobra.Command{
	Use:   "interaction",
	Short: "L2 Behavioral Graph commands",
	Long:  "Log behavioral interaction events to the FuseMomo immutable interaction graph.",
}

//  interaction log

var interactionLogCmd = &cobra.Command{
	Use:   "log",
	Short: "Log a single interaction event (POST /v1/core/interactions/log)",
	RunE: func(cmd *cobra.Command, args []string) error {
		entityID, _ := cmd.Flags().GetString("entity-id")
		if !isValidUUID(entityID) {
			output.WriteErrorMsg(1, "validation_error", "--entity-id must be a valid UUID")
			os.Exit(1)
		}

		outcome, _ := cmd.Flags().GetString("outcome")
		validOutcomes := map[string]bool{"success": true, "failed": true, "pending": true, "ignored": true, "unknown": true}
		if !validOutcomes[outcome] {
			output.WriteErrorMsg(1, "validation_error", "--outcome must be one of: success, failed, pending, ignored, unknown")
			os.Exit(1)
		}

		req := api.InteractionLogRequest{
			EntityID:   entityID,
			API:        mustFlag(cmd, "api"),
			ActionType: mustFlag(cmd, "action-type"),
			Action:     mustFlag(cmd, "action"),
			Outcome:    outcome,
		}

		if v, _ := cmd.Flags().GetString("intent"); v != "" {
			req.Intent = &v
		}
		if v, _ := cmd.Flags().GetString("agent-id"); v != "" {
			req.AgentID = &v
		}
		if v, _ := cmd.Flags().GetString("external-ref"); v != "" {
			req.ExternalRef = &v
		}
		if metaStr, _ := cmd.Flags().GetString("meta"); metaStr != "" {
			var meta map[string]interface{}
			if err := json.Unmarshal([]byte(metaStr), &meta); err != nil {
				output.WriteErrorMsg(1, "validation_error", fmt.Sprintf("--meta contains invalid JSON: %v", err))
				os.Exit(1)
			}
			req.Metadata = meta
		}
		if v, _ := cmd.Flags().GetString("occurred-at"); v != "" {
			t, err := time.Parse(time.RFC3339, v)
			if err != nil {
				output.WriteErrorMsg(1, "validation_error", fmt.Sprintf("--occurred-at must be RFC3339 format: %v", err))
				os.Exit(1)
			}
			req.OccurredAt = &t
		}

		if isDryRun(cmd) {
			return output.JSON(req)
		}

		result, err := tui.RunWithSpinner("Logging interaction", func() (*api.InteractionLogResponse, error) {
			return apiClient.LogInteraction(context.Background(), req)
		})
		if err != nil {
			handleCLIError(err)
		}

		return output.Print(getOutputFormat(cmd), "interaction", result, isNoColor(cmd))
	},
}

//  interaction batch

var interactionBatchCmd = &cobra.Command{
	Use:   "batch",
	Short: "Log multiple interactions from a JSON file (POST /v1/core/interactions/batch)",
	RunE: func(cmd *cobra.Command, args []string) error {
		filePath, _ := cmd.Flags().GetString("file")

		data, err := os.ReadFile(filePath)
		if err != nil {
			output.WriteErrorMsg(1, "file_error", fmt.Sprintf("cannot read file %q: %v", filePath, err))
			os.Exit(1)
		}

		var interactions []api.InteractionLogRequest
		if err := json.Unmarshal(data, &interactions); err != nil {
			output.WriteErrorMsg(1, "validation_error", fmt.Sprintf("file contains invalid JSON: %v", err))
			os.Exit(1)
		}
		if len(interactions) == 0 {
			output.WriteErrorMsg(1, "validation_error", "file must contain at least 1 interaction")
			os.Exit(1)
		}
		if len(interactions) > 100 {
			output.WriteErrorMsg(1, "validation_error", "batch maximum is 100 interactions")
			os.Exit(1)
		}

		req := api.BatchInteractionLogRequest{Interactions: interactions}

		if isDryRun(cmd) {
			return output.JSON(req)
		}

		result, err := tui.RunBatchWithProgress(func() (*api.BatchInteractionLogResponse, error) {
			return apiClient.BatchInteractions(context.Background(), req)
		})
		if err != nil {
			handleCLIError(err)
		}

		return output.Print(getOutputFormat(cmd), "interaction-batch", result, isNoColor(cmd))
	},
}

// mustFlag gets a string flag value — safe to use after MarkFlagRequired.
func mustFlag(cmd *cobra.Command, name string) string {
	v, _ := cmd.Flags().GetString(name)
	return v
}
