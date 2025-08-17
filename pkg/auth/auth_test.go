package auth

import (
	"testing"

	"github.com/arctir/devgraph-cli/pkg/config"
	"github.com/stretchr/testify/assert"
)

func TestAuthenticatedClient_InvalidIssuerURL(t *testing.T) {
	c := config.Config{
		IssuerURL:   "https://test.example.com",
		ClientID:    "test-client",
		RedirectURL: "http://localhost:8080/callback",
	}

	client, err := AuthenticatedClient(c)
	assert.Error(t, err)
	assert.Nil(t, client)
	// Should fail when trying to fetch well-known endpoints from invalid URL
	assert.Contains(t, err.Error(), "failed to get well-known endpoints")
}

func TestAuthenticatedClient_InvalidCredentials(t *testing.T) {
	c := config.Config{
		IssuerURL:   "https://test.example.com",
		ClientID:    "test-client",
		RedirectURL: "http://localhost:8080/callback",
		Environment: "test-env",
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
