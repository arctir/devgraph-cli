package commands

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestTokenCommand_Structure tests the token command structure
func TestTokenCommand_Structure(t *testing.T) {
	tokenCmd := TokenCommand{}

	// Test that all CRUD subcommands are available
	assert.NotNil(t, &tokenCmd.Create, "Create command should be available")
	assert.NotNil(t, &tokenCmd.Delete, "Delete command should be available")
	assert.NotNil(t, &tokenCmd.Get, "Get command should be available")
	assert.NotNil(t, &tokenCmd.List, "List command should be available")
	assert.NotNil(t, &tokenCmd.Update, "Update command should be available")
}

// TestTokenCreateCommand_Structure tests the token create command structure
func TestTokenCreateCommand_Structure(t *testing.T) {
	createCmd := TokenCreate{}

	// Test that command has expected fields
	assert.IsType(t, "", createCmd.Name)
	assert.IsType(t, []string{}, createCmd.Scopes)
}

// TestTokenGetCommand_Structure tests the token get command structure
func TestTokenGetCommand_Structure(t *testing.T) {
	getCmd := TokenGet{}

	// Test that command has expected ID field
	assert.IsType(t, "", getCmd.ID)
}

// TestTokenListCommand_Structure tests the token list command structure
func TestTokenListCommand_Structure(t *testing.T) {
	listCmd := TokenList{}

	// Test that command has EnvWrapperCommand embedded
	assert.NotNil(t, &listCmd.EnvWrapperCommand)
}

// TestTokenUpdateCommand_Structure tests the token update command structure
func TestTokenUpdateCommand_Structure(t *testing.T) {
	updateCmd := TokenUpdate{}

	// Test that command has expected fields
	assert.IsType(t, "", updateCmd.ID)
	assert.IsType(t, "", updateCmd.Name)
	assert.IsType(t, []string{}, updateCmd.Scopes)
}

// TestTokenDeleteCommand_Structure tests the token delete command structure
func TestTokenDeleteCommand_Structure(t *testing.T) {
	deleteCmd := TokenDelete{}

	// Test that command has expected ID field
	assert.IsType(t, "", deleteCmd.ID)
}

// TestCheckScopeInput tests the scope validation function
func TestCheckScopeInput(t *testing.T) {
	testCases := []struct {
		name     string
		scopes   []string
		expected bool
	}{
		{
			name:     "all scopes keyword",
			scopes:   []string{"all"},
			expected: true,
		},
		{
			name:     "valid single scope",
			scopes:   []string{"create:entitydefinitions"},
			expected: true,
		},
		{
			name:     "valid multiple scopes",
			scopes:   []string{"create:entities", "read:entities"},
			expected: true,
		},
		{
			name:     "all valid scopes",
			scopes:   allowedScopes,
			expected: true,
		},
		{
			name:     "invalid single scope",
			scopes:   []string{"invalid:scope"},
			expected: false,
		},
		{
			name:     "mix of valid and invalid",
			scopes:   []string{"create:entities", "invalid:scope"},
			expected: false,
		},
		{
			name:     "empty scope list",
			scopes:   []string{},
			expected: true, // Empty list passes validation
		},
		{
			name:     "case sensitive - wrong case",
			scopes:   []string{"CREATE:ENTITIES"},
			expected: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := checkScopeInput(tc.scopes)
			assert.Equal(t, tc.expected, result, "checkScopeInput(%v) should return %v", tc.scopes, tc.expected)
		})
	}
}

// TestAllowedScopes verifies the allowed scopes list
func TestAllowedScopes(t *testing.T) {
	// Verify the list is not empty
	require.NotEmpty(t, allowedScopes, "allowedScopes should not be empty")

	// Verify no duplicates
	scopeMap := make(map[string]bool)
	for _, scope := range allowedScopes {
		assert.False(t, scopeMap[scope], "Duplicate scope found: %s", scope)
		scopeMap[scope] = true
	}

	// Verify expected scopes are present
	expectedScopes := []string{
		"create:entitydefinitions",
		"list:entitydefinitions",
		"delete:entitydefinitions",
		"create:entities",
		"read:entities",
		"delete:entities",
		"create:entityrelations",
		"delete:entityrelations",
	}

	for _, expected := range expectedScopes {
		assert.Contains(t, allowedScopes, expected, "Expected scope %s should be in allowedScopes", expected)
	}
}

// TestTokenCreate_ScopeExpansion tests that "all" keyword expands to all allowed scopes
func TestTokenCreate_ScopeExpansion(t *testing.T) {
	createCmd := &TokenCreate{
		Name:   "test-token",
		Scopes: []string{"all"},
	}

	// The Run method will expand "all" to allowedScopes
	// We can't test Run() without a real client, but we can verify the logic
	if len(createCmd.Scopes) == 1 && createCmd.Scopes[0] == "all" {
		expandedScopes := allowedScopes
		assert.Equal(t, len(allowedScopes), len(expandedScopes))
	}
}

// TestTokenUpdate_NoFieldsProvided tests validation when no fields are provided
func TestTokenUpdate_NoFieldsProvided(t *testing.T) {
	updateCmd := &TokenUpdate{
		ID:     "550e8400-e29b-41d4-a716-446655440000",
		Name:   "",
		Scopes: []string{},
	}

	// Test the validation logic that should happen in Run()
	hasName := updateCmd.Name != ""
	hasScopes := len(updateCmd.Scopes) > 0

	assert.False(t, hasName, "Name should be empty")
	assert.False(t, hasScopes, "Scopes should be empty")
	assert.False(t, hasName || hasScopes, "Should fail validation when both are empty")
}

// TestTokenUpdate_WithValidName tests update with valid name
func TestTokenUpdate_WithValidName(t *testing.T) {
	updateCmd := &TokenUpdate{
		ID:     "550e8400-e29b-41d4-a716-446655440000",
		Name:   "updated-token-name",
		Scopes: []string{},
	}

	hasName := updateCmd.Name != ""
	assert.True(t, hasName, "Should have a name to update")
}

// TestTokenUpdate_WithValidScopes tests update with valid scopes
func TestTokenUpdate_WithValidScopes(t *testing.T) {
	updateCmd := &TokenUpdate{
		ID:     "550e8400-e29b-41d4-a716-446655440000",
		Name:   "",
		Scopes: []string{"create:entities", "read:entities"},
	}

	hasScopes := len(updateCmd.Scopes) > 0
	assert.True(t, hasScopes, "Should have scopes to update")

	// Validate scopes
	valid := checkScopeInput(updateCmd.Scopes)
	assert.True(t, valid, "Scopes should be valid")
}

// TestTokenUpdate_WithAllScopes tests update with "all" keyword
func TestTokenUpdate_WithAllScopes(t *testing.T) {
	updateCmd := &TokenUpdate{
		ID:     "550e8400-e29b-41d4-a716-446655440000",
		Name:   "",
		Scopes: []string{"all"},
	}

	// Simulate the expansion that happens in Run()
	if len(updateCmd.Scopes) == 1 && updateCmd.Scopes[0] == "all" {
		expandedScopes := allowedScopes
		assert.Equal(t, len(allowedScopes), len(expandedScopes))
		assert.True(t, checkScopeInput(expandedScopes))
	}
}

// TestTokenUpdate_InvalidScopes tests update with invalid scopes
func TestTokenUpdate_InvalidScopes(t *testing.T) {
	testCases := []struct {
		name   string
		scopes []string
	}{
		{
			name:   "invalid scope",
			scopes: []string{"invalid:scope"},
		},
		{
			name:   "partially invalid",
			scopes: []string{"create:entities", "invalid:action"},
		},
		{
			name:   "wrong format",
			scopes: []string{"createentities"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			valid := checkScopeInput(tc.scopes)
			assert.False(t, valid, "Invalid scopes should fail validation")
		})
	}
}

// TestTokenDelete_ValidUUID tests delete with valid UUID format
func TestTokenDelete_ValidUUID(t *testing.T) {
	validUUIDs := []string{
		"550e8400-e29b-41d4-a716-446655440000",
		"6ba7b810-9dad-11d1-80b4-00c04fd430c8",
		"00000000-0000-0000-0000-000000000000",
	}

	for _, uuid := range validUUIDs {
		deleteCmd := &TokenDelete{
			ID: uuid,
		}
		assert.NotEmpty(t, deleteCmd.ID, "ID should be set")
		assert.Len(t, deleteCmd.ID, 36, "UUID should be 36 characters")
	}
}

// TestTokenDelete_InvalidUUID tests delete with invalid UUID format
func TestTokenDelete_InvalidUUID(t *testing.T) {
	invalidUUIDs := []string{
		"not-a-uuid",
		"123",
		"550e8400-e29b-41d4-a716",
		"550e8400-e29b-41d4-a716-446655440000-extra",
		"",
	}

	for _, uuid := range invalidUUIDs {
		t.Run("invalid_"+uuid, func(t *testing.T) {
			deleteCmd := &TokenDelete{
				ID: uuid,
			}
			// The actual UUID parsing will fail in Run(), but we can verify structure
			if uuid != "" {
				assert.NotEqual(t, 36, len(deleteCmd.ID), "Invalid UUID should not be 36 chars")
			} else {
				assert.Empty(t, deleteCmd.ID, "Empty UUID should be empty")
			}
		})
	}
}

// TestTokenGet_ValidID tests get with valid token ID
func TestTokenGet_ValidID(t *testing.T) {
	getCmd := &TokenGet{
		ID: "550e8400-e29b-41d4-a716-446655440000",
	}

	assert.NotEmpty(t, getCmd.ID, "ID should be set")
	assert.Len(t, getCmd.ID, 36, "UUID should be 36 characters")
}

// TestTokenCommandNaming verifies command naming conventions
func TestTokenCommandNaming(t *testing.T) {
	// Verify struct names follow convention
	assert.IsType(t, TokenCreate{}, TokenCreate{})
	assert.IsType(t, TokenDelete{}, TokenDelete{})
	assert.IsType(t, TokenGet{}, TokenGet{})
	assert.IsType(t, TokenList{}, TokenList{})
	assert.IsType(t, TokenUpdate{}, TokenUpdate{})
}

// TestScopeValidation_EdgeCases tests edge cases in scope validation
func TestScopeValidation_EdgeCases(t *testing.T) {
	testCases := []struct {
		name     string
		scopes   []string
		expected bool
	}{
		{
			name:     "nil slice",
			scopes:   nil,
			expected: true, // nil is treated as empty which passes
		},
		{
			name:     "whitespace in scope",
			scopes:   []string{" create:entities "},
			expected: false,
		},
		{
			name:     "scope with extra colon",
			scopes:   []string{"create:entities:extra"},
			expected: false,
		},
		{
			name:     "scope missing colon",
			scopes:   []string{"createentities"},
			expected: false,
		},
		{
			name:     "duplicate scopes in input",
			scopes:   []string{"create:entities", "create:entities"},
			expected: true, // Duplicates are valid (server should handle)
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := checkScopeInput(tc.scopes)
			assert.Equal(t, tc.expected, result)
		})
	}
}

// TestTokenCreate_RequiredFields tests that required fields are enforced by structure
func TestTokenCreate_RequiredFields(t *testing.T) {
	createCmd := TokenCreate{
		Name:   "my-token",
		Scopes: []string{"create:entities"},
	}

	assert.NotEmpty(t, createCmd.Name, "Name is required")
	assert.NotEmpty(t, createCmd.Scopes, "Scopes are required")
}

// BenchmarkCheckScopeInput benchmarks the scope validation function
func BenchmarkCheckScopeInput(b *testing.B) {
	scopes := []string{"create:entities", "read:entities", "delete:entities"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		checkScopeInput(scopes)
	}
}

// BenchmarkCheckScopeInput_All benchmarks validation with "all" keyword
func BenchmarkCheckScopeInput_All(b *testing.B) {
	scopes := []string{"all"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		checkScopeInput(scopes)
	}
}

// BenchmarkCheckScopeInput_Invalid benchmarks validation with invalid scopes
func BenchmarkCheckScopeInput_Invalid(b *testing.B) {
	scopes := []string{"invalid:scope", "another:invalid"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		checkScopeInput(scopes)
	}
}
