package api_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/fusemomo/fusemomo-cli/internal/api"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newTestClient creates a Client pointed at the given test server URL.
func newTestClient(apiURL string) *api.Client {
	return api.NewClient("fm_live_test_key", apiURL, 5)
}

func TestClient_AuthHeader(t *testing.T) {
	var gotAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(api.EntitiesListResponse{Entities: []api.EntityResponse{}, Total: 0, Limit: 20, Offset: 0})
	}))
	defer srv.Close()

	client := newTestClient(srv.URL)
	_, err := client.ListEntities(context.Background(), 20, 0, "")
	require.NoError(t, err)
	assert.Equal(t, "Bearer fm_live_test_key", gotAuth)
}

func TestClient_ContentTypeOnPost(t *testing.T) {
	var gotCT string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotCT = r.Header.Get("Content-Type")
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(api.ResolveEntityResponse{EntityID: "ent_01", CreatedAt: time.Now()})
	}))
	defer srv.Close()

	client := newTestClient(srv.URL)
	_, err := client.ResolveEntity(context.Background(), api.ResolveEntityRequest{
		Identifiers: map[string]string{"email": "test@example.com"},
	})
	require.NoError(t, err)
	assert.True(t, strings.HasPrefix(gotCT, "application/json"))
}

func TestClient_XRequestIDPresent(t *testing.T) {
	var gotID string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotID = r.Header.Get("X-Request-ID")
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(api.EntitiesListResponse{})
	}))
	defer srv.Close()

	client := newTestClient(srv.URL)
	_, err := client.ListEntities(context.Background(), 20, 0, "")
	require.NoError(t, err)
	assert.NotEmpty(t, gotID)
}

func TestClient_RetryOn5xx(t *testing.T) {
	attempts := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts <= 2 {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "server error", "code": "server_error"})
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(api.EntitiesListResponse{})
	}))
	defer srv.Close()

	client := newTestClient(srv.URL)
	_, err := client.ListEntities(context.Background(), 20, 0, "")
	require.NoError(t, err)
	assert.Equal(t, 3, attempts, "expected 2 retries + 1 success = 3 total attempts")
}

func TestClient_NoRetryOn4xx(t *testing.T) {
	attempts := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "not found", "code": "not_found"})
	}))
	defer srv.Close()

	client := newTestClient(srv.URL)
	_, err := client.ListEntities(context.Background(), 20, 0, "")
	require.Error(t, err)
	assert.Equal(t, 1, attempts, "expected exactly 1 attempt, no retries on 4xx")
}

func TestClient_ExitCode_401(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "unauthorized", "code": "authentication_error"})
	}))
	defer srv.Close()

	client := newTestClient(srv.URL)
	_, err := client.ListEntities(context.Background(), 20, 0, "")
	require.Error(t, err)
	cliErr, ok := err.(*api.CLIError)
	require.True(t, ok)
	assert.Equal(t, 3, cliErr.ExitCode)
}

func TestClient_ExitCode_402(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusPaymentRequired)
		json.NewEncoder(w).Encode(map[string]string{"error": "plan error", "code": "plan_error"})
	}))
	defer srv.Close()

	client := newTestClient(srv.URL)
	_, err := client.GetRecommendation(context.Background(), api.RecommendRequest{EntityID: "ent", Intent: "test"})
	require.Error(t, err)
	cliErr, ok := err.(*api.CLIError)
	require.True(t, ok)
	assert.Equal(t, 3, cliErr.ExitCode)
}

func TestClient_ExitCode_404(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "not found", "code": "not_found"})
	}))
	defer srv.Close()

	client := newTestClient(srv.URL)
	_, err := client.GetEntity(context.Background(), "00000000-0000-0000-0000-000000000001")
	require.Error(t, err)
	cliErr, ok := err.(*api.CLIError)
	require.True(t, ok)
	assert.Equal(t, 1, cliErr.ExitCode)
}

func TestClient_UserAgentHeader(t *testing.T) {
	var gotUA string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotUA = r.Header.Get("User-Agent")
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(api.EntitiesListResponse{})
	}))
	defer srv.Close()

	api.Version = "v1.0.0-test"
	client := newTestClient(srv.URL)
	_, _ = client.ListEntities(context.Background(), 20, 0, "")
	assert.True(t, strings.HasPrefix(gotUA, "fusemomo-cli/"), "User-Agent should start with fusemomo-cli/")
}
