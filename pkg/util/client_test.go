package util

import (
	"os"
	"testing"

	"github.com/arctir/devgraph-cli/pkg/config"
	"github.com/stretchr/testify/assert"
)

// setupTempConfig sets XDG_CONFIG_HOME to a temp directory for testing
// Returns a cleanup function to restore the original value
func setupTempConfig(t *testing.T) func() {
	t.Helper()
	originalXDG := os.Getenv("XDG_CONFIG_HOME")
	tempDir := t.TempDir()
	os.Setenv("XDG_CONFIG_HOME", tempDir)
	return func() {
		if originalXDG == "" {
			os.Unsetenv("XDG_CONFIG_HOME")
		} else {
			os.Setenv("XDG_CONFIG_HOME", originalXDG)
		}
	}
}

func TestGetAuthenticatedHTTPClient_InvalidConfig(t *testing.T) {
	// Use temp config so we don't pick up real credentials
	cleanup := setupTempConfig(t)
	defer cleanup()

	invalidConfig := config.Config{
		ApiURL:    "invalid-url",
		IssuerURL: "invalid-issuer",
		ClientID:  "invalid-client",
	}

	client, err := GetAuthenticatedHTTPClient(invalidConfig)

	// Should return an error for invalid configuration (no credentials)
	assert.Error(t, err)
	assert.Nil(t, client)
}

func TestGetAuthenticatedClient_InvalidConfig(t *testing.T) {
	// Use temp config so we don't pick up real credentials
	cleanup := setupTempConfig(t)
	defer cleanup()

	invalidConfig := config.Config{
		ApiURL:    "invalid-url",
		IssuerURL: "invalid-issuer",
		ClientID:  "invalid-client",
	}

	client, err := GetAuthenticatedClient(invalidConfig)

	// Should return an error for invalid configuration (no credentials)
	assert.Error(t, err)
	assert.Nil(t, client)
}

func TestGetAuthenticatedClient_ValidURL_NoAuth(t *testing.T) {
	// Use temp config so we don't pick up real credentials
	cleanup := setupTempConfig(t)
	defer cleanup()

	// Test with valid URL structure but no authentication
	testConfig := config.Config{
		ApiURL:    "https://api.example.com",
		IssuerURL: "https://issuer.example.com",
		ClientID:  "test-client",
	}

	client, err := GetAuthenticatedClient(testConfig)

	// Should fail due to authentication issues (no credentials)
	assert.Error(t, err)
	assert.Nil(t, client)
}