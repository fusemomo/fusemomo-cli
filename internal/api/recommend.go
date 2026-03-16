package api

import (
	"context"
	"fmt"
	"net/http"
)

// GetRecommendation calls POST /v1/core/recommends.
func (c *Client) GetRecommendation(ctx context.Context, req RecommendRequest) (*RecommendResponse, error) {
	resp, err := c.do(ctx, http.MethodPost, "/v1/core/recommends", req)
	if err != nil {
		return nil, err
	}
	var result RecommendResponse
	return &result, decodeJSON(resp, &result)
}

// UpdateOutcome calls PATCH /v1/core/recommends/:id/outcomes.
func (c *Client) UpdateOutcome(ctx context.Context, recommendationID string, req RecommendOutcomeRequest) (*RecommendOutcomeResponse, error) {
	path := fmt.Sprintf("/v1/core/recommends/%s/outcomes", recommendationID)
	resp, err := c.do(ctx, http.MethodPatch, path, req)
	if err != nil {
		return nil, err
	}
	var result RecommendOutcomeResponse
	return &result, decodeJSON(resp, &result)
}
