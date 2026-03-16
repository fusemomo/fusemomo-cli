package cmd

import (
	"bufio"
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/fusemomo/fusemomo-cli/pkg/config"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(setupCmd)
}

var setupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Interactive one-time setup — saves API key to ~/.fusemomo/config.yaml",
	RunE:  runSetup,
}

func runSetup(cmd *cobra.Command, args []string) error {
	// Use bufio.NewReader — NOT go-prompt — so the terminal stays in
	// cooked mode and is never corrupted after the process exits.
	reader := bufio.NewReader(os.Stdin)

	fmt.Fprintln(os.Stderr, "FuseMomo CLI Setup")

	// Prompt for API key.
	fmt.Fprint(os.Stderr, "Enter your FuseMomo API key: ")
	apiKey, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("failed to read API key: %w", err)
	}
	apiKey = strings.TrimSpace(apiKey)

	if apiKey == "" {
		fmt.Fprintln(os.Stderr, `{"error":"API key cannot be empty","code":"validation_error"}`)
		os.Exit(1)
	}
	if !strings.HasPrefix(apiKey, "fm_live_") && !strings.HasPrefix(apiKey, "fm_test_") {
		fmt.Fprintln(os.Stderr, `{"error":"API key must start with fm_live_ or fm_test_","code":"validation_error"}`)
		os.Exit(1)
	}

	// Prompt for API URL.
	fmt.Fprint(os.Stderr, "Enter API URL [https://api.fusemomo.com]: ")
	rawURL, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("failed to read API URL: %w", err)
	}
	rawURL = strings.TrimSpace(rawURL)
	if rawURL == "" {
		rawURL = "https://api.fusemomo.com"
	}

	// Verify connectivity via GET /ping.
	fmt.Fprintf(os.Stderr, "Verifying connectivity to %s... ", rawURL)
	if err := pingServer(rawURL); err != nil {
		fmt.Fprintf(os.Stderr, "FAILED\n")
		fmt.Fprintf(os.Stderr, `{"error":"%s","code":"connection_error"}`+"\n", err.Error())
		os.Exit(2)
	}
	fmt.Fprintln(os.Stderr, "OK")

	// Ensure the config directory exists with restricted permissions.
	dirPath, err := config.ConfigDirPath()
	if err != nil {
		return fmt.Errorf("cannot determine config directory: %w", err)
	}
	if err := os.MkdirAll(dirPath, 0700); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Write config file with 0600 permissions (owner read/write only).
	cfgPath := filepath.Join(dirPath, "config.yaml")
	content := fmt.Sprintf("api_key: %s\napi_url: %s\ntimeout: 30\noutput: json\ndebug: false\n",
		apiKey, rawURL)
	if err := os.WriteFile(cfgPath, []byte(content), 0600); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	fmt.Fprintf(os.Stdout, `{"message":"Configuration saved to %s"}`+"\n", cfgPath)
	return nil
}

// pingServer verifies connectivity by calling GET /ping (no auth required).
func pingServer(apiURL string) error {
	url := strings.TrimRight(apiURL, "/") + "/ping"
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	resp.Body.Close()
	if resp.StatusCode >= 500 {
		return fmt.Errorf("server returned HTTP %d", resp.StatusCode)
	}
	return nil
}
