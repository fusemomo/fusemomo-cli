package api

import (
	"context"
	"net/http"
)

// LogInteraction calls POST /v1/core/interactions/log.
func (c *Client) LogInteraction(ctx context.Context, req InteractionLogRequest) (*InteractionLogResponse, error) {
	resp, err := c.do(ctx, http.MethodPost, "/v1/core/interactions/log", req)
	if err != nil {
		return nil, err
	}
	var result InteractionLogResponse
	return &result, decodeJSON(resp, &result)
}

// BatchInteractions calls POST /v1/core/interactions/batch.
func (c *Client) BatchInteractions(ctx context.Context, req BatchInteractionLogRequest) (*BatchInteractionLogResponse, error) {
	resp, err := c.do(ctx, http.MethodPost, "/v1/core/interactions/batch", req)
	if err != nil {
		return nil, err
	}
	var result BatchInteractionLogResponse
	return &result, decodeJSON(resp, &result)
}
