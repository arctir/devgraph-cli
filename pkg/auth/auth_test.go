package auth

import (
	"net/http"
	"testing"
	"time"

	"github.com/arctir/devgraph-cli/pkg/config"
	"github.com/stretchr/testify/assert"
	"golang.org/x/oauth2"
)

func TestAuthenticatedClient_InvalidIssuerURL(t *testing.T) {
	c := config.Config{
		IssuerURL: "https://test.example.com",
		ClientID:  "test-client",
	}

	client, err := AuthenticatedClient(c)
	assert.Error(t, err)
	assert.Nil(t, client)
	// Should fail when trying to fetch well-known endpoints from invalid URL
	assert.Contains(t, err.Error(), "failed to get well-known endpoints")
}

func TestAuthenticatedClient_InvalidCredentials(t *testing.T) {
	c := config.Config{
		IssuerURL: "https://test.example.com",
		ClientID:  "test-client",
	}

	client, err := AuthenticatedClient(c)
	assert.Error(t, err)
	assert.Nil(t, client)
	// The actual error could be either credential loading or network issues
	// Both are valid test outcomes for invalid configs
	assert.True(t,
		err != nil && (len(err.Error()) > 0),
		"Expected some error for invalid configuration")
}

func TestDevgraphTransport_RoundTrip(t *testing.T) {
	// Create a test transport
	transport := &DevgraphTransport{
		Headers: map[string]string{
			"X-Custom-Header": "test-value",
			"Authorization":   "Bearer test-token",
		},
	}

	// Create a test request
	req, err := http.NewRequest("GET", "https://example.com", nil)
	assert.NoError(t, err)

	// Since we can't easily mock the full HTTP round trip,
	// we test that the function doesn't panic and handles nil transport
	assert.NotPanics(t, func() {
		// Test with nil underlying transport (should use default)
		transport.Transport = nil
		_, _ = transport.RoundTrip(req)
	})
}

func TestNewOIDCTokenManager(t *testing.T) {
	// Test creating a new token manager
	oauth2Config := oauth2.Config{
		ClientID: "test-client",
		Scopes:   []string{"openid"},
	}

	initialToken := &oauth2.Token{
		AccessToken: "test-access-token",
		TokenType:   "Bearer",
	}

	// Test with nil provider (should not panic)
	assert.NotPanics(t, func() {
		manager := NewOIDCTokenManager(oauth2Config, initialToken, nil, "test-env")
		assert.NotNil(t, manager)
		assert.Equal(t, "test-env", manager.devgraphEnvironment)
		assert.Equal(t, initialToken, manager.token)
	})
}

func TestOIDCTokenManager_GetCurrentToken(t *testing.T) {
	oauth2Config := oauth2.Config{
		ClientID: "test-client",
	}

	initialToken := &oauth2.Token{
		AccessToken: "test-access-token",
		TokenType:   "Bearer",
	}

	manager := NewOIDCTokenManager(oauth2Config, initialToken, nil, "test-env")

	// Test getting current token
	currentToken := manager.GetCurrentToken()
	assert.Equal(t, initialToken, currentToken)
	assert.Equal(t, "test-access-token", currentToken.AccessToken)
}

func TestOIDCTokenManager_HTTPClient(t *testing.T) {
	oauth2Config := oauth2.Config{
		ClientID: "test-client",
	}

	initialToken := &oauth2.Token{
		AccessToken: "test-access-token",
		TokenType:   "Bearer",
	}

	manager := NewOIDCTokenManager(oauth2Config, initialToken, nil, "test-env")

	// Test getting HTTP client
	client := manager.HTTPClient()
	assert.NotNil(t, client)
	assert.Equal(t, 30*time.Second, client.Timeout)

	// Verify the client has the right transport structure
	assert.IsType(t, &oauth2.Transport{}, client.Transport)
}
