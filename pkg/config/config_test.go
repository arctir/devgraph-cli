package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadConfig(t *testing.T) {
	tests := []struct {
		name          string
		configContent string
		expected      *Config
		expectError   bool
	}{
		{
			name: "valid config",
			configContent: `apiurl: https://api.example.com
issuerurl: https://issuer.example.com
clientid: test-client-id
debug: true`,
			expected: &Config{
				ApiURL:    "https://api.example.com",
				IssuerURL: "https://issuer.example.com",
				ClientID:  "test-client-id",
				Debug:     true,
			},
			expectError: false,
		},
		{
			name: "partial config",
			configContent: `apiurl: https://api.example.com`,
			expected: &Config{
				ApiURL: "https://api.example.com",
			},
			expectError: false,
		},
		{
			name:          "invalid yaml",
			configContent: "invalid: yaml: content: [unclosed",
			expected:      nil,
			expectError:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			configFile := filepath.Join(tmpDir, "config.yaml")

			err := os.WriteFile(configFile, []byte(tt.configContent), 0600)
			require.NoError(t, err)

			config, err := LoadConfig(configFile)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, config)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, config)
			}
		})
	}
}

func TestLoadConfig_NonExistentFile(t *testing.T) {
	config, err := LoadConfig("/non/existent/file.yaml")
	assert.Error(t, err)
	assert.Nil(t, config)
	assert.Contains(t, err.Error(), "config file does not exist")
}

func TestSaveConfig(t *testing.T) {
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "test-config.yaml")

	config := &Config{
		ApiURL:    "https://api.test.com",
		IssuerURL: "https://issuer.test.com",
		ClientID:  "test-client",
		Debug:     true,
	}

	err := SaveConfig(configFile, config)
	assert.NoError(t, err)

	assert.FileExists(t, configFile)

	loadedConfig, err := LoadConfig(configFile)
	assert.NoError(t, err)
	assert.Equal(t, config, loadedConfig)
}

func TestSaveConfig_CreateDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "subdir", "nested", "config.yaml")

	config := &Config{
		ApiURL: "https://api.test.com",
	}

	err := SaveConfig(configFile, config)
	assert.NoError(t, err)

	assert.FileExists(t, configFile)

	loadedConfig, err := LoadConfig(configFile)
	assert.NoError(t, err)
	assert.Equal(t, config, loadedConfig)
}

func TestUserSettings(t *testing.T) {
	// Test UserSettings struct
	settings := UserSettings{
		DefaultEnvironment: "production",
		DefaultModel:       "gpt-4",
		DefaultMaxTokens:   2000,
	}

	assert.Equal(t, "production", settings.DefaultEnvironment)
	assert.Equal(t, "gpt-4", settings.DefaultModel)
	assert.Equal(t, 2000, settings.DefaultMaxTokens)
}
