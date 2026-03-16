package api

import (
	"context"
	"fmt"
	"net/http"
)

// ResolveEntity calls POST /v1/core/entities/resolve.
func (c *Client) ResolveEntity(ctx context.Context, req ResolveEntityRequest) (*ResolveEntityResponse, error) {
	resp, err := c.do(ctx, http.MethodPost, "/v1/core/entities/resolve", req)
	if err != nil {
		return nil, err
	}
	var result ResolveEntityResponse
	return &result, decodeJSON(resp, &result)
}

// GetEntity calls GET /v1/core/entities/:id.
func (c *Client) GetEntity(ctx context.Context, entityID string) (*EntityDetailResponse, error) {
	resp, err := c.do(ctx, http.MethodGet, fmt.Sprintf("/v1/core/entities/%s", entityID), nil)
	if err != nil {
		return nil, err
	}
	var result EntityDetailResponse
	return &result, decodeJSON(resp, &result)
}

// ListEntities calls GET /v1/core/entities with optional query params.
func (c *Client) ListEntities(ctx context.Context, limit, offset int, entityType string) (*EntitiesListResponse, error) {
	path := fmt.Sprintf("/v1/core/entities?limit=%d&offset=%d", limit, offset)
	if entityType != "" {
		path += "&entity_type=" + entityType
	}
	resp, err := c.do(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}
	var result EntitiesListResponse
	return &result, decodeJSON(resp, &result)
}

// DeleteEntity calls DELETE /v1/core/entities/:id.
func (c *Client) DeleteEntity(ctx context.Context, entityID string) (*EntityDeleteResponse, error) {
	resp, err := c.do(ctx, http.MethodDelete, fmt.Sprintf("/v1/core/entities/%s", entityID), nil)
	if err != nil {
		return nil, err
	}
	var result EntityDeleteResponse
	return &result, decodeJSON(resp, &result)
}

// LinkEntity calls POST /v1/core/entities/:id/link.
func (c *Client) LinkEntity(ctx context.Context, entityID string, req LinkIdentifiersRequest) (*LinkIdentifiersResponse, error) {
	resp, err := c.do(ctx, http.MethodPost, fmt.Sprintf("/v1/core/entities/%s/link", entityID), req)
	if err != nil {
		return nil, err
	}
	var result LinkIdentifiersResponse
	return &result, decodeJSON(resp, &result)
}
