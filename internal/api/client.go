package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/fusemomo/fusemomo-cli/internal/logging"
	"go.uber.org/zap"
)

// Version is set by the linker at build time and used for the User-Agent header.
var Version = "dev"

// Client is a thin HTTP wrapper over the FuseMomo REST API.
// All business logic and retry handling lives here; Cobra commands stay thin.
type Client struct {
	apiKey  string
	apiURL  string
	timeout time.Duration
	http    *http.Client
}

// NewClient creates a new API client with the given configuration.
func NewClient(apiKey, apiURL string, timeoutSecs int) *Client {
	if timeoutSecs <= 0 {
		timeoutSecs = 30
	}
	return &Client{
		apiKey:  apiKey,
		apiURL:  apiURL,
		timeout: time.Duration(timeoutSecs) * time.Second,
		http:    &http.Client{},
	}
}

// do executes an HTTP request with retries, auth headers, and timeout.
// It returns a CLIError with the correct exit code on failure.
func (c *Client) do(ctx context.Context, method, path string, body interface{}) (*http.Response, error) {
	log := logging.Get()

	var bodyBytes []byte
	if body != nil {
		var err error
		bodyBytes, err = json.Marshal(body)
		if err != nil {
			return nil, &CLIError{ExitCode: 1, Code: "serialization_error", Message: fmt.Sprintf("failed to serialize request: %v", err)}
		}
	}

	url := c.apiURL + path

	const maxRetries = 2
	backoff := []time.Duration{500 * time.Millisecond, 1000 * time.Millisecond}

	var lastErr *CLIError
	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			log.Warn("retrying request",
				zap.String("method", method),
				zap.String("url", url),
				zap.Int("attempt", attempt),
			)
			time.Sleep(backoff[attempt-1])
		}

		reqCtx, cancel := context.WithTimeout(ctx, c.timeout)
		defer cancel()

		req, err := http.NewRequestWithContext(reqCtx, method, url, bytes.NewReader(bodyBytes))
		if err != nil {
			return nil, &CLIError{ExitCode: 2, Code: "request_error", Message: fmt.Sprintf("failed to build request: %v", err)}
		}

		// Required headers.
		req.Header.Set("Authorization", "Bearer "+c.apiKey)
		req.Header.Set("User-Agent", "fusemomo-cli/"+Version)
		req.Header.Set("X-Request-ID", generateRequestID())
		if method == http.MethodPost || method == http.MethodPatch {
			req.Header.Set("Content-Type", "application/json")
		}

		log.Debug("HTTP request",
			zap.String("method", method),
			zap.String("url", url),
			zap.String("authorization", "Bearer "+logging.RedactKey(c.apiKey)),
		)

		start := time.Now()
		resp, err := c.http.Do(req)
		elapsed := time.Since(start)

		if err != nil {
			// Network error — retry.
			log.Warn("network error", zap.String("url", url), zap.Error(err))
			lastErr = &CLIError{ExitCode: 2, Code: "connection_error",
				Message: fmt.Sprintf("failed to reach %s: %v", c.apiURL, err)}
			if attempt == maxRetries {
				return nil, lastErr
			}
			continue
		}

		log.Debug("HTTP response",
			zap.Int("status", resp.StatusCode),
			zap.Duration("elapsed", elapsed),
		)

		if elapsed > 5*time.Second {
			log.Warn("slow API response", zap.Duration("elapsed", elapsed), zap.String("url", url))
		}

		// 5xx — retry.
		if resp.StatusCode >= 500 {
			resp.Body.Close()
			lastErr = &CLIError{
				ExitCode: 2,
				Code:     "server_error",
				Status:   resp.StatusCode,
				Message:  fmt.Sprintf("server error: HTTP %d at %s", resp.StatusCode, url),
			}
			if attempt == maxRetries {
				return nil, lastErr
			}
			continue
		}

		// 4xx — do not retry; map to exit code.
		if resp.StatusCode >= 400 {
			cliErr := mapHTTPStatus(resp)
			return nil, cliErr
		}

		return resp, nil
	}

	return nil, lastErr
}

// mapHTTPStatus decodes the API error body and maps the HTTP status to a CLI exit code.
func mapHTTPStatus(resp *http.Response) *CLIError {
	defer resp.Body.Close()

	var apiErr APIErrorResponse
	_ = json.NewDecoder(resp.Body).Decode(&apiErr)

	msg := apiErr.Message
	if msg == "" {
		msg = apiErr.Error
	}
	if msg == "" {
		msg = fmt.Sprintf("HTTP %d", resp.StatusCode)
	}

	code := apiErr.Code
	if code == "" {
		code = httpStatusCode(resp.StatusCode)
	}

	exitCode := httpExitCode(resp.StatusCode)
	return &CLIError{ExitCode: exitCode, Code: code, Message: msg, Status: resp.StatusCode}
}

func httpExitCode(status int) int {
	switch status {
	case 401, 402:
		return 3
	case 400, 404, 409, 429:
		return 1
	default:
		return 2
	}
}

func httpStatusCode(status int) string {
	switch status {
	case 400:
		return "validation_error"
	case 401:
		return "authentication_error"
	case 402:
		return "plan_error"
	case 404:
		return "not_found"
	case 409:
		return "conflict"
	case 429:
		return "rate_limit_exceeded"
	default:
		return "server_error"
	}
}

// decodeJSON reads resp.Body into v and closes the body.
func decodeJSON(resp *http.Response, v interface{}) error {
	defer resp.Body.Close()
	if err := json.NewDecoder(resp.Body).Decode(v); err != nil {
		return &CLIError{ExitCode: 2, Code: "decode_error", Message: fmt.Sprintf("failed to decode response: %v", err)}
	}
	return nil
}

// generateRequestID creates a simple unique request ID for tracing.
func generateRequestID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}
