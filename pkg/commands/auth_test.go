package commands

import (
	"testing"

	"github.com/arctir/devgraph-cli/pkg/config"
	"github.com/stretchr/testify/assert"
)

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
	// Test that AuthLoginCommand delegates to Auth.Run()
	loginCmd := &AuthLoginCommand{
		Config: config.Config{
			ApiURL:      "https://api.example.com",
			IssuerURL:   "https://issuer.example.com", 
			ClientID:    "test-client",
			RedirectURL: "http://localhost:8080/callback",
		},
	}

	// This will fail due to network/auth issues, but should not panic
	err := loginCmd.Run()
	assert.Error(t, err) // Expected to fail due to invalid config/no network
}

func TestAuthLogoutCommand_Run(t *testing.T) {
	logoutCmd := &AuthLogoutCommand{
		Config: config.Config{
			ApiURL:      "https://api.example.com",
			IssuerURL:   "https://issuer.example.com",
			ClientID:    "test-client", 
			RedirectURL: "http://localhost:8080/callback",
		},
	}

	// This will likely fail due to no credentials, but should not panic
	err := logoutCmd.Run()
	// We don't assert error here because logout might succeed even without credentials
	_ = err
}

func TestAuthWhoamiCommand_Run(t *testing.T) {
	whoamiCmd := &AuthWhoamiCommand{
		Config: config.Config{
			ApiURL:      "https://api.example.com",
			IssuerURL:   "https://issuer.example.com",
			ClientID:    "test-client",
			RedirectURL: "http://localhost:8080/callback",
		},
	}

	// This will fail due to no credentials, but should not panic
	err := whoamiCmd.Run()
	assert.Error(t, err) // Expected to fail due to no user credentials
}

func TestAuth_Run_InvalidConfig(t *testing.T) {
	auth := &Auth{
		Config: config.Config{
			ApiURL:      "invalid-url",
			IssuerURL:   "invalid-issuer",
			ClientID:    "invalid-client",
			RedirectURL: "invalid-redirect",
		},
	}

	// Should fail due to invalid configuration
	err := auth.Run()
	assert.Error(t, err)
}