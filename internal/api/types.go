package api

import "time"

// ResolveEntityRequest is the body for POST /v1/core/entities/resolve.
type ResolveEntityRequest struct {
	Identifiers map[string]string `json:"identifiers"`
	EntityType  *string           `json:"entity_type,omitempty"`
	DisplayName *string           `json:"display_name,omitempty"`
	Metadata    map[string]any    `json:"metadata,omitempty"`
}

// EntityIdentifier describes a single identifier linked to an entity.
type EntityIdentifier struct {
	ID              string     `json:"id"`
	Source          string     `json:"source"`
	IdentifierType  string     `json:"identifier_type"`
	IdentifierValue string     `json:"identifier_value"`
	Confidence      float64    `json:"confidence"`
	LinkStrategy    string     `json:"link_strategy"`
	VerifiedAt      *time.Time `json:"verified_at,omitempty"`
}

// ResolveEntityResponse is the success body for POST /v1/core/entities/resolve.
type ResolveEntityResponse struct {
	EntityID               string             `json:"entity_id"`
	Identifiers            []EntityIdentifier `json:"identifiers"`
	EntityType             *string            `json:"entity_type,omitempty"`
	DisplayName            *string            `json:"display_name,omitempty"`
	TotalInteractions      int                `json:"total_interactions"`
	SuccessfulInteractions int                `json:"successful_interactions"`
	LastInteractionAt      *time.Time         `json:"last_interaction_at,omitempty"`
	PreferredActionType    *string            `json:"preferred_action_type,omitempty"`
	BehavioralScore        *float64           `json:"behavioral_score,omitempty"`
	Metadata               map[string]any     `json:"metadata,omitempty"`
	CreatedAt              time.Time          `json:"created_at"`
}

// InteractionSummary is an abbreviated interaction record in entity detail.
type InteractionSummary struct {
	ID         string    `json:"id"`
	API        string    `json:"api"`
	ActionType string    `json:"action_type"`
	Outcome    string    `json:"outcome"`
	OccurredAt time.Time `json:"occurred_at"`
}

// EntityResponse is a full entity record including behavioral stats.
type EntityResponse struct {
	ID                     string                 `json:"id"`
	TenantID               string                 `json:"tenant_id"`
	DisplayName            string                 `json:"display_name,omitempty"`
	EntityType             string                 `json:"entity_type,omitempty"`
	TotalInteractions      int                    `json:"total_interactions"`
	SuccessfulInteractions int                    `json:"successful_interactions"`
	LastInteractionAt      *time.Time             `json:"last_interaction_at,omitempty"`
	PreferredActionType    string                 `json:"preferred_action_type,omitempty"`
	BehavioralScore        *float64               `json:"behavioral_score,omitempty"`
	Metadata               map[string]interface{} `json:"metadata"`
	CreatedAt              time.Time              `json:"created_at"`
	UpdatedAt              time.Time              `json:"updated_at"`
}

// EntityDetailResponse embeds EntityResponse with identifiers and recent interactions.
type EntityDetailResponse struct {
	EntityResponse
	Identifiers        []EntityIdentifier   `json:"identifiers"`
	RecentInteractions []InteractionSummary `json:"recent_interactions"`
}

// EntitiesListResponse is the response for GET /v1/core/entities.
type EntitiesListResponse struct {
	Entities []EntityResponse `json:"entities"`
	Total    int              `json:"total"`
	Limit    int              `json:"limit"`
	Offset   int              `json:"offset"`
}

// EntityDeleteResponse is the response for DELETE /v1/core/entities/:id.
type EntityDeleteResponse struct {
	EntityID   string    `json:"entity_id"`
	Anonymized bool      `json:"anonymized"`
	ErasedAt   time.Time `json:"erased_at"`
}

// LinkIdentifiersRequest is the body for POST /v1/core/entities/:id/link.
type LinkIdentifiersRequest struct {
	Identifiers  map[string]string `json:"identifiers"`
	LinkStrategy *string           `json:"link_strategy,omitempty"` // "deterministic" | "probabilistic"
	Confidence   *float64          `json:"confidence,omitempty"`    // 0.0–1.0
}

// LinkIdentifiersResponse is the response for POST /v1/core/entities/:id/link.
type LinkIdentifiersResponse struct {
	EntityID    string             `json:"entity_id"`
	Identifiers []EntityIdentifier `json:"identifiers"`
	LinkedCount int                `json:"linked_count"`
}

//  Interaction Types

// InteractionLogRequest is the body for POST /v1/core/interactions/log.
type InteractionLogRequest struct {
	EntityID    string                 `json:"entity_id"`
	API         string                 `json:"api"`
	ActionType  string                 `json:"action_type"`
	Action      string                 `json:"action"`
	Outcome     string                 `json:"outcome"` // success|failed|pending|ignored|unknown
	Intent      *string                `json:"intent,omitempty"`
	AgentID     *string                `json:"agent_id,omitempty"`
	ExternalRef *string                `json:"external_ref,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	OccurredAt  *time.Time             `json:"occurred_at,omitempty"`
}

// InteractionLogResponse is the response for POST /v1/core/interactions/log.
type InteractionLogResponse struct {
	InteractionID string    `json:"interaction_id"`
	EntityID      string    `json:"entity_id"`
	LoggedAt      time.Time `json:"logged_at"`
}

// BatchInteractionLogRequest is the body for POST /v1/core/interactions/batch.
type BatchInteractionLogRequest struct {
	Interactions []InteractionLogRequest `json:"interactions"`
}

// BatchInteractionLogResponse is the response for POST /v1/core/interactions/batch.
type BatchInteractionLogResponse struct {
	LoggedCount int       `json:"logged_count"`
	FirstID     string    `json:"first_id"`
	LastID      string    `json:"last_id"`
	LoggedAt    time.Time `json:"logged_at"`
}

//  Recommendation Types

// RecommendRequest is the body for POST /v1/core/recommends.
type RecommendRequest struct {
	EntityID      string `json:"entity_id"`
	Intent        string `json:"intent"`
	LookbackDays  int    `json:"lookback_days,omitempty"`
	MinSampleSize int    `json:"min_sample_size,omitempty"`
}

// RecommendResponse is the response for POST /v1/core/recommends.
type RecommendResponse struct {
	RecommendationID      *string            `json:"recommendation_id"`
	EntityID              string             `json:"entity_id"`
	Intent                string             `json:"intent"`
	RecommendedActionType *string            `json:"recommended_action_type"`
	Confidence            *float64           `json:"confidence"`
	ScoringBreakdown      map[string]float64 `json:"scoring_breakdown"`
	Reason                string             `json:"reason"`
	SampleSize            int64              `json:"sample_size"`
	LookbackDays          int                `json:"lookback_days"`
}

// RecommendOutcomeRequest is the body for PATCH /v1/core/recommends/:id/outcomes.
type RecommendOutcomeRequest struct {
	WasFollowed          bool    `json:"was_followed"`
	OutcomeInteractionID *string `json:"outcome_interaction_id,omitempty"`
}

// RecommendOutcomeResponse is the response for PATCH /v1/core/recommends/:id/outcomes.
type RecommendOutcomeResponse struct {
	RecommendationID string    `json:"recommendation_id"`
	WasFollowed      bool      `json:"was_followed"`
	Outcome          *string   `json:"outcome"`
	UpdatedAt        time.Time `json:"updated_at"`
}

//  Error Types

// APIErrorResponse represents an error returned by the FuseMomo REST API.
type APIErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
	Code    string `json:"code"`
}

// CLIError wraps an API or network error with a CLI exit code.
type CLIError struct {
	ExitCode int
	Code     string
	Message  string
	Status   int
}

func (e *CLIError) Error() string {
	return e.Message
}
