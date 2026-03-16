package cmd

import (
	"context"
	"os"

	"github.com/fusemomo/fusemomo-cli/internal/api"
	"github.com/fusemomo/fusemomo-cli/internal/output"
	"github.com/fusemomo/fusemomo-cli/internal/tui"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(recommendCmd)
	recommendCmd.AddCommand(recommendGetCmd)
	recommendCmd.AddCommand(recommendOutcomeCmd)

	// recommend get flags.
	recommendGetCmd.Flags().String("entity-id", "", "UUID of the entity (required)")
	recommendGetCmd.Flags().String("intent", "", "Business goal e.g. payment_recovery (required)")
	recommendGetCmd.Flags().Int("lookback", 0, "Lookback window in days (0 = plan default, max 730)")
	recommendGetCmd.Flags().Int("min-sample", 0, "Min interactions required (0 = default of 2, max 100)")
	recommendGetCmd.MarkFlagRequired("entity-id")
	recommendGetCmd.MarkFlagRequired("intent")

	// recommend outcome flags.
	recommendOutcomeCmd.Flags().Bool("followed", false, "Whether the recommendation was followed (required)")
	recommendOutcomeCmd.Flags().String("interaction-id", "", "UUID of the resulting interaction (optional but recommended)")
	recommendOutcomeCmd.MarkFlagRequired("followed")
}

var recommendCmd = &cobra.Command{
	Use:   "recommend",
	Short: "L3 Behavioral Intelligence commands",
	Long:  "Query the highest-success action recommendations and close the feedback loop.",
}

//  recommend get

var recommendGetCmd = &cobra.Command{
	Use:   "get",
	Short: "Get action recommendation for an entity and intent (POST /v1/core/recommends)",
	RunE: func(cmd *cobra.Command, args []string) error {
		entityID, _ := cmd.Flags().GetString("entity-id")
		if !isValidUUID(entityID) {
			output.WriteErrorMsg(1, "validation_error", "--entity-id must be a valid UUID")
			os.Exit(1)
		}

		intent, _ := cmd.Flags().GetString("intent")
		lookback, _ := cmd.Flags().GetInt("lookback")
		minSample, _ := cmd.Flags().GetInt("min-sample")

		req := api.RecommendRequest{
			EntityID:      entityID,
			Intent:        intent,
			LookbackDays:  lookback,
			MinSampleSize: minSample,
		}

		if isDryRun(cmd) {
			return output.JSON(req)
		}

		result, err := tui.RunWithSpinner("Getting recommendation", func() (*api.RecommendResponse, error) {
			return apiClient.GetRecommendation(context.Background(), req)
		})
		if err != nil {
			handleCLIError(err)
		}

		// Insufficient data is exit code 0 — not an error.
		// recommended_action_type will be null in the JSON output.
		return output.Print(getOutputFormat(cmd), "recommend", result, isNoColor(cmd))
	},
}

//  recommend outcome

var recommendOutcomeCmd = &cobra.Command{
	Use:   "outcome <recommendation_id>",
	Short: "Record recommendation outcome — closes feedback loop (PATCH /v1/core/recommends/:id/outcomes)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		recID := args[0]
		if !isValidUUID(recID) {
			output.WriteErrorMsg(1, "validation_error", "recommendation_id must be a valid UUID")
			os.Exit(1)
		}

		followed, _ := cmd.Flags().GetBool("followed")
		req := api.RecommendOutcomeRequest{WasFollowed: followed}

		if intID, _ := cmd.Flags().GetString("interaction-id"); intID != "" {
			if !isValidUUID(intID) {
				output.WriteErrorMsg(1, "validation_error", "--interaction-id must be a valid UUID")
				os.Exit(1)
			}
			req.OutcomeInteractionID = &intID
		}

		if isDryRun(cmd) {
			return output.JSON(req)
		}

		result, err := tui.RunWithSpinner("Updating outcome", func() (*api.RecommendOutcomeResponse, error) {
			return apiClient.UpdateOutcome(context.Background(), recID, req)
		})
		if err != nil {
			handleCLIError(err)
		}

		return output.Print(getOutputFormat(cmd), "recommend-outcome", result, isNoColor(cmd))
	},
}
