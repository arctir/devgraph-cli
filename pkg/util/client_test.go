package util

import (
	"testing"

	"github.com/arctir/devgraph-cli/pkg/config"
	"github.com/stretchr/testify/assert"
)

func TestGetAuthenticatedHTTPClient_InvalidConfig(t *testing.T) {
	invalidConfig := config.Config{
		ApiURL:    "invalid-url",
		IssuerURL: "invalid-issuer",
		ClientID:  "invalid-client",
	}

	client, err := GetAuthenticatedHTTPClient(invalidConfig)
	
	// Should return an error for invalid configuration
	assert.Error(t, err)
	assert.Nil(t, client)
}

func TestGetAuthenticatedClient_InvalidConfig(t *testing.T) {
	invalidConfig := config.Config{
		ApiURL:    "invalid-url", 
		IssuerURL: "invalid-issuer",
		ClientID:  "invalid-client",
	}

	client, err := GetAuthenticatedClient(invalidConfig)
	
	// Should return an error for invalid configuration  
	assert.Error(t, err)
	assert.Nil(t, client)
}

func TestGetAuthenticatedClient_ValidURL_NoAuth(t *testing.T) {
	// Test with valid URL structure but no authentication
	testConfig := config.Config{
		ApiURL:    "https://api.example.com",
		IssuerURL: "https://issuer.example.com", 
		ClientID:  "test-client",
	}

	client, err := GetAuthenticatedClient(testConfig)
	
	// Should fail due to authentication issues, not URL parsing
	assert.Error(t, err)
	assert.Nil(t, client)
	// Error should be related to authentication, not URL parsing
	assert.NotContains(t, err.Error(), "parsing")
}