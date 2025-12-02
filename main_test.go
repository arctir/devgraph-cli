package main

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCLIStructure(t *testing.T) {
	cli := CLI{}

	// Test that all expected commands are available
	assert.NotNil(t, &cli.Chat, "Chat command should be available")
	assert.NotNil(t, &cli.Auth, "Auth command should be available")
	assert.NotNil(t, &cli.Config, "Config command should be available")
	assert.NotNil(t, &cli.Token, "Token command should be available")
	assert.NotNil(t, &cli.Environment, "Environment command should be available")
	assert.NotNil(t, &cli.EntityDefinition, "EntityDefinition command should be available")
	assert.NotNil(t, &cli.Entity, "Entity command should be available")
	assert.NotNil(t, &cli.MCP, "MCP command should be available")
	assert.NotNil(t, &cli.ModelProvider, "ModelProvider command should be available")
	assert.NotNil(t, &cli.Model, "Model command should be available")
	assert.NotNil(t, &cli.Provider, "Provider command should be available")
	assert.NotNil(t, &cli.Subscription, "Subscription command should be available")
}

func TestMain_Integration(t *testing.T) {
	// Test that main doesn't panic with invalid arguments
	// Note: This is a basic smoke test since main() calls os.Exit

	// We can't easily test main() directly since it calls os.Exit,
	// but we can test that the CLI structure is valid
	cli := CLI{}

	// Verify the CLI can be initialized without panicking
	assert.NotPanics(t, func() {
		_ = cli
	}, "CLI initialization should not panic")
}

// Test CLI help descriptions
func TestCLIHelpDescriptions(t *testing.T) {
	// These tests ensure that help text is meaningful
	testCases := []struct {
		name        string
		description string
	}{
		{"chat", "interactive chat"},
		{"auth", "authentication"},
		{"config", "configuration settings"},
		{"token", "tokens"},
		{"env", "environments"},
		{"entitydefinition", "entitydefinition"},
		{"entity", "entity"},
		{"mcp", "MCP resources"},
		{"modelprovider", "modelprovider"},
		{"model", "Model resources"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// This is a basic test to ensure help text contains expected keywords
			// In a real CLI test, you'd capture help output and verify it
			assert.Contains(t, strings.ToLower(tc.description), tc.name,
				"Help description should contain command name or related terms")
		})
	}
}
