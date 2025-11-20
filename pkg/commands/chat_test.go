package commands

import (
	"testing"
	"github.com/arctir/devgraph-cli/pkg/config"
	"github.com/stretchr/testify/assert"
)

func TestChatCommand_Structure(t *testing.T) {
	chatCmd := &Chat{}
	
	// Test that the command has expected embedded config
	assert.IsType(t, EnvWrapperCommand{}, chatCmd.EnvWrapperCommand)
}

func TestChatCommand_SlashCommands(t *testing.T) {
	chatCmd := &Chat{
		EnvWrapperCommand: EnvWrapperCommand{
			Config: config.Config{},
		},
		Model: "test-model",
	}
	
	// Test help command
	err := chatCmd.handleSlashCommand("/help")
	assert.NoError(t, err)
	
	// Test unknown command
	err = chatCmd.handleSlashCommand("/unknown")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unknown command")
	
	// Test model command (will fail due to no API, but should not panic)
	err = chatCmd.handleSlashCommand("/model")
	assert.Error(t, err) // Expected to fail due to no mock API
}