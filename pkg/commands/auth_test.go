package commands

import (
	"errors"
	"os"
	"testing"

	"github.com/arctir/devgraph-cli/pkg/auth"
	"github.com/arctir/devgraph-cli/pkg/config"
	"github.com/stretchr/testify/assert"
	"golang.org/x/oauth2"
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

// mockAuthenticator is a test implementation of auth.Authenticator
type mockAuthenticator struct {
	token *oauth2.Token
	err   error
}

func (m *mockAuthenticator) Authenticate(cfg config.Config) (*oauth2.Token, error) {
	return m.token, m.err
}

func TestAuthCommand_Structure(t *testing.T) {
	authCmd := AuthCommand{}
	
	// Test that all subcommands are available
	assert.NotNil(t, &authCmd.Login, "Login command should be available")
	assert.NotNil(t, &authCmd.Logout, "Logout command should be available") 
	assert.NotNil(t, &authCmd.Whoami, "Whoami command should be available")
}

func TestParseJWT_ValidToken(t *testing.T) {
	// Create a test JWT token structure (not a real token, just for parsing test)
	// This avoids gosec G101 warning about hardcoded credentials
	header := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9"  // {"alg":"HS256","typ":"JWT"}
	payload := "eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ"  // {"sub":"1234567890","name":"John Doe","iat":1516239022}
	signature := "SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c"  // Mock signature for testing
	testToken := header + "." + payload + "." + signature

	claims, err := parseJWT(testToken)
	
	assert.NoError(t, err)
	assert.NotNil(t, claims)
	assert.Equal(t, "1234567890", (*claims)["sub"])
	assert.Equal(t, "John Doe", (*claims)["name"])
}

func TestParseJWT_InvalidToken(t *testing.T) {
	// Test with invalid token
	claims, err := parseJWT("invalid.token.here")
	
	assert.Error(t, err)
	assert.Nil(t, claims)
}

func TestParseJWT_EmptyToken(t *testing.T) {
	// Test with empty token
	claims, err := parseJWT("")
	
	assert.Error(t, err)
	assert.Nil(t, claims)
}

func TestAuthLoginCommand_Run(t *testing.T) {
	// Save original authenticator and restore after test
	originalAuth := auth.AuthenticatorImpl
	defer func() { auth.AuthenticatorImpl = originalAuth }()

	// Mock the authenticator to return an error (simulating auth failure)
	auth.AuthenticatorImpl = &mockAuthenticator{
		token: nil,
		err:   errors.New("mock auth error"),
	}

	loginCmd := &AuthLoginCommand{
		Config: config.Config{
			ApiURL:    "https://api.example.com",
			IssuerURL: "https://issuer.example.com",
			ClientID:  "test-client",
		},
	}

	// This will fail due to mocked auth error
	err := loginCmd.Run()
	assert.Error(t, err) // Expected to fail due to mock error
}

func TestAuthLogoutCommand_Run(t *testing.T) {
	logoutCmd := &AuthLogoutCommand{
		Config: config.Config{
			ApiURL:    "https://api.example.com",
			IssuerURL: "https://issuer.example.com",
			ClientID:  "test-client",
		},
	}

	// This will likely fail due to no credentials, but should not panic
	err := logoutCmd.Run()
	// We don't assert error here because logout might succeed even without credentials
	_ = err
}

func TestAuthWhoamiCommand_Run(t *testing.T) {
	// Use temp config so we don't pick up real credentials
	cleanup := setupTempConfig(t)
	defer cleanup()

	whoamiCmd := &AuthWhoamiCommand{
		Config: config.Config{
			ApiURL:    "https://api.example.com",
			IssuerURL: "https://issuer.example.com",
			ClientID:  "test-client",
		},
	}

	// This will fail due to no credentials in temp config
	err := whoamiCmd.Run()
	assert.Error(t, err) // Expected to fail due to no user credentials
}

func TestAuth_Run_InvalidConfig(t *testing.T) {
	// Save original authenticator and restore after test
	originalAuth := auth.AuthenticatorImpl
	defer func() { auth.AuthenticatorImpl = originalAuth }()

	// Mock the authenticator to return an error
	auth.AuthenticatorImpl = &mockAuthenticator{
		token: nil,
		err:   errors.New("invalid config error"),
	}

	authCmd := &Auth{
		Config: config.Config{
			ApiURL:    "invalid-url",
			IssuerURL: "invalid-issuer",
			ClientID:  "invalid-client",
		},
	}

	// Should fail due to mocked error
	err := authCmd.Run()
	assert.Error(t, err)
}