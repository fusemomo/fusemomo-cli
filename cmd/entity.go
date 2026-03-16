package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/fusemomo/fusemomo-cli/internal/api"
	"github.com/fusemomo/fusemomo-cli/internal/output"
	"github.com/fusemomo/fusemomo-cli/internal/tui"
	"github.com/spf13/cobra"
)

func init() {
	// entity is the parent command.
	rootCmd.AddCommand(entityCmd)

	// Subcommands.
	entityCmd.AddCommand(entityResolveCmd)
	entityCmd.AddCommand(entityGetCmd)
	entityCmd.AddCommand(entityListCmd)
	entityCmd.AddCommand(entityDeleteCmd)
	entityCmd.AddCommand(entityLinkCmd)

	// entity resolve flags.
	entityResolveCmd.Flags().StringArray("id", nil, "Identifier pair key=value (repeatable, min 1)")
	entityResolveCmd.Flags().String("type", "", "Entity type e.g. contact, order")
	entityResolveCmd.Flags().String("name", "", "Human-readable display name")
	entityResolveCmd.Flags().String("meta", "", "Additional metadata as JSON string")
	entityResolveCmd.MarkFlagRequired("id")

	// entity get flags.
	// No extra flags: entity_id is a positional arg.

	// entity list flags.
	entityListCmd.Flags().Int("limit", 20, "Number of results (max 100)")
	entityListCmd.Flags().Int("offset", 0, "Pagination offset")
	entityListCmd.Flags().String("type", "", "Filter by entity_type")

	// entity delete flags.
	entityDeleteCmd.Flags().Bool("confirm", false, "Must be set to confirm irreversible deletion")

	// entity link flags.
	entityLinkCmd.Flags().StringArray("id", nil, "Identifier pair key=value (repeatable, min 1)")
	entityLinkCmd.Flags().String("strategy", "deterministic", "Link strategy: deterministic or probabilistic")
	entityLinkCmd.Flags().Float64("confidence", 1.0, "Confidence score 0.0–1.0")
	entityLinkCmd.MarkFlagRequired("id")
}

var entityCmd = &cobra.Command{
	Use:   "entity",
	Short: "L1 Identity Resolution commands",
	Long:  "Resolve, retrieve, list, delete, and link entities via the FuseMomo REST API.",
}

//  entity resolve

var entityResolveCmd = &cobra.Command{
	Use:   "resolve",
	Short: "Resolve identifiers into a canonical entity (POST /v1/core/entities/resolve)",
	RunE: func(cmd *cobra.Command, args []string) error {
		idPairs, _ := cmd.Flags().GetStringArray("id")
		identifiers, err := parseIdentifierPairs(idPairs)
		if err != nil {
			output.WriteErrorMsg(1, "validation_error", err.Error())
			os.Exit(1)
		}

		typeParm, _ := cmd.Flags().GetString("type")
		nameParm, _ := cmd.Flags().GetString("name")
		metaStr, _ := cmd.Flags().GetString("meta")

		req := api.ResolveEntityRequest{Identifiers: identifiers}
		if typeParm != "" {
			req.EntityType = &typeParm
		}
		if nameParm != "" {
			req.DisplayName = &nameParm
		}
		if metaStr != "" {
			var meta map[string]any
			if err := json.Unmarshal([]byte(metaStr), &meta); err != nil {
				output.WriteErrorMsg(1, "validation_error", fmt.Sprintf("--meta contains invalid JSON: %v", err))
				os.Exit(1)
			}
			req.Metadata = meta
		}

		if isDryRun(cmd) {
			return output.JSON(req)
		}

		result, err := tui.RunWithSpinner("Resolving entity", func() (*api.ResolveEntityResponse, error) {
			return apiClient.ResolveEntity(context.Background(), req)
		})
		if err != nil {
			handleCLIError(err)
		}

		return output.Print(getOutputFormat(cmd), "entity", result, isNoColor(cmd))
	},
}

//  entity get

var entityGetCmd = &cobra.Command{
	Use:   "get <entity_id>",
	Short: "Get full entity profile (GET /v1/core/entities/:id)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		entityID := args[0]
		if !isValidUUID(entityID) {
			output.WriteErrorMsg(1, "validation_error", "entity_id must be a valid UUID")
			os.Exit(1)
		}

		result, err := tui.RunWithSpinner("Fetching entity", func() (*api.EntityDetailResponse, error) {
			return apiClient.GetEntity(context.Background(), entityID)
		})
		if err != nil {
			handleCLIError(err)
		}

		return output.Print(getOutputFormat(cmd), "entity", result, isNoColor(cmd))
	},
}

//  entity list

var entityListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all entities (GET /v1/core/entities)",
	RunE: func(cmd *cobra.Command, args []string) error {
		limit, _ := cmd.Flags().GetInt("limit")
		offset, _ := cmd.Flags().GetInt("offset")
		entityType, _ := cmd.Flags().GetString("type")

		result, err := tui.RunWithSpinner("Listing entities", func() (*api.EntitiesListResponse, error) {
			return apiClient.ListEntities(context.Background(), limit, offset, entityType)
		})
		if err != nil {
			handleCLIError(err)
		}

		return output.Print(getOutputFormat(cmd), "entity-list", result, isNoColor(cmd))
	},
}

//  entity delete

var entityDeleteCmd = &cobra.Command{
	Use:   "delete <entity_id>",
	Short: "GDPR anonymize an entity — IRREVERSIBLE (DELETE /v1/core/entities/:id)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		confirm, _ := cmd.Flags().GetBool("confirm")
		if !confirm {
			output.WriteErrorMsg(1, "confirmation_required",
				"deletion is irreversible — pass --confirm to proceed")
			os.Exit(1)
		}

		entityID := args[0]
		if !isValidUUID(entityID) {
			output.WriteErrorMsg(1, "validation_error", "entity_id must be a valid UUID")
			os.Exit(1)
		}

		if isDryRun(cmd) {
			return output.JSON(map[string]string{"entity_id": entityID, "action": "delete"})
		}

		result, err := tui.RunWithSpinner("Deleting entity", func() (*api.EntityDeleteResponse, error) {
			return apiClient.DeleteEntity(context.Background(), entityID)
		})
		if err != nil {
			handleCLIError(err)
		}

		return output.JSON(result)
	},
}

//  entity link

var entityLinkCmd = &cobra.Command{
	Use:   "link <entity_id>",
	Short: "Link additional identifiers to an entity (POST /v1/core/entities/:id/link)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		entityID := args[0]
		if !isValidUUID(entityID) {
			output.WriteErrorMsg(1, "validation_error", "entity_id must be a valid UUID")
			os.Exit(1)
		}

		idPairs, _ := cmd.Flags().GetStringArray("id")
		identifiers, err := parseIdentifierPairs(idPairs)
		if err != nil {
			output.WriteErrorMsg(1, "validation_error", err.Error())
			os.Exit(1)
		}

		strategy, _ := cmd.Flags().GetString("strategy")
		confidence, _ := cmd.Flags().GetFloat64("confidence")

		req := api.LinkIdentifiersRequest{
			Identifiers:  identifiers,
			LinkStrategy: &strategy,
			Confidence:   &confidence,
		}

		if isDryRun(cmd) {
			return output.JSON(req)
		}

		result, err := tui.RunWithSpinner("Linking identifiers", func() (*api.LinkIdentifiersResponse, error) {
			return apiClient.LinkEntity(context.Background(), entityID, req)
		})
		if err != nil {
			handleCLIError(err)
		}

		return output.JSON(result)
	},
}

//  Helpers

// parseIdentifierPairs converts ["key=value", "key2=value2"] to a map.
func parseIdentifierPairs(pairs []string) (map[string]string, error) {
	result := make(map[string]string, len(pairs))
	for _, pair := range pairs {
		idx := strings.Index(pair, "=")
		if idx <= 0 {
			return nil, fmt.Errorf("invalid identifier pair %q — expected format key=value", pair)
		}
		key := pair[:idx]
		val := pair[idx+1:]
		if val == "" {
			return nil, fmt.Errorf("identifier value for key %q cannot be empty", key)
		}
		result[key] = val
	}
	if len(result) == 0 {
		return nil, fmt.Errorf("at least one --id pair is required")
	}
	return result, nil
}

// isValidUUID does a lightweight UUID format check.
func isValidUUID(s string) bool {
	if len(s) != 36 {
		return false
	}
	for i, c := range s {
		if i == 8 || i == 13 || i == 18 || i == 23 {
			if c != '-' {
				return false
			}
		} else if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')) {
			return false
		}
	}
	return true
}
