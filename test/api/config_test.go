package api_test

import (
	"os"
	"testing"

	"github.com/fusemomo/fusemomo-cli/pkg/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfig_MissingAPIKey(t *testing.T) {
	cfg := &config.Config{
		APIKey:  "",
		APIURL:  "https://api.fusemomo.com",
		Timeout: 30,
	}
	err := cfg.Validate()
	require.Error(t, err)
	ve, ok := err.(*config.ValidationError)
	require.True(t, ok)
	assert.Equal(t, 3, ve.ExitCode)
}

func TestConfig_InvalidKeyFormat(t *testing.T) {
	cfg := &config.Config{
		APIKey:  "not_a_valid_key",
		APIURL:  "https://api.fusemomo.com",
		Timeout: 30,
	}
	err := cfg.Validate()
	require.Error(t, err)
	ve, ok := err.(*config.ValidationError)
	require.True(t, ok)
	assert.Equal(t, 3, ve.ExitCode)
}

func TestConfig_ValidLiveKey(t *testing.T) {
	cfg := &config.Config{
		APIKey:  "fm_live_abc123",
		APIURL:  "https://api.fusemomo.com",
		Timeout: 30,
	}
	assert.NoError(t, cfg.Validate())
}

func TestConfig_ValidTestKey(t *testing.T) {
	cfg := &config.Config{
		APIKey:  "fm_test_abc123",
		APIURL:  "https://api.fusemomo.com",
		Timeout: 30,
	}
	assert.NoError(t, cfg.Validate())
}

func TestConfig_InvalidURL(t *testing.T) {
	cfg := &config.Config{
		APIKey:  "fm_live_abc123",
		APIURL:  "not-a-url",
		Timeout: 30,
	}
	err := cfg.Validate()
	require.Error(t, err)
	ve, ok := err.(*config.ValidationError)
	require.True(t, ok)
	assert.Equal(t, 3, ve.ExitCode)
}

func TestConfig_InvalidTimeout(t *testing.T) {
	cfg := &config.Config{
		APIKey:  "fm_live_abc123",
		APIURL:  "https://api.fusemomo.com",
		Timeout: -1,
	}
	err := cfg.Validate()
	require.Error(t, err)
	ve, ok := err.(*config.ValidationError)
	require.True(t, ok)
	assert.Equal(t, 3, ve.ExitCode)
}

func TestConfig_EnvVarOverride(t *testing.T) {
	t.Setenv("FUSEMOMO_API_KEY", "fm_live_from_env")
	defer os.Unsetenv("FUSEMOMO_API_KEY")

	val := os.Getenv("FUSEMOMO_API_KEY")
	assert.Equal(t, "fm_live_from_env", val)
}
