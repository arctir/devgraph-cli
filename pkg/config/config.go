package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/golang-jwt/jwt/v5"
	"gopkg.in/yaml.v3"
)

const DefaultIssuerURL = "https://primary-ghoul-65.clerk.accounts.dev"
const DefaultClientID = "renbud3BkDcW1utM"
const DefaultRedirectURL = "http://localhost:8080/callback"

type Config struct {
	ApiURL      string `kong:"default='https://api.staging.devgraph.ai',env='DEVGRAPH_API_URL',help='Devgraph API URL'"`
	IssuerURL   string `kong:"default='https://primary-ghoul-65.clerk.accounts.dev',env='DEVGRAPH_ISSUER_URL',help='Devgraph issuer URL'"`
	ClientID    string `kong:"default='renbud3BkDcW1utM',env='DEVGRAPH_CLIENT_ID',help='Devgraph client ID'"`
	RedirectURL string `kong:"default='http://localhost:8080/callback',env='DEVGRAPH_REDIRECT_URL',help='Redirect URL'"`
	Environment string `kong:"env='DEVGRAPH_ENVIRONMENT',help='Environment (development, staging, production)'"`

	Model     string `kong:"default='gpt-4o-mini',short='m',help='Chat model to use'"`
	MaxTokens int    `kong:"default=1000,short='t',help='Maximum number of tokens in response'"`
}

// UserConfig represents the unified user configuration file
type UserConfig struct {
	// User preferences
	Settings UserSettings `yaml:"settings,omitempty"`

	// Authentication credentials
	Credentials Credentials `yaml:"credentials,omitempty"`
}

// UserSettings represents persistent user preferences
type UserSettings struct {
	DefaultEnvironment string `yaml:"default_environment,omitempty"`
	DefaultModel       string `yaml:"default_model,omitempty"`
	DefaultMaxTokens   int    `yaml:"default_max_tokens,omitempty"`
}

// Credentials represents authentication tokens
type Credentials struct {
	AccessToken  string         `yaml:"access_token,omitempty"`
	RefreshToken string         `yaml:"refresh_token,omitempty"`
	IDToken      string         `yaml:"id_token,omitempty"`
	Claims       *jwt.MapClaims `yaml:"claims,omitempty"`
}

// LoadConfig reads and unmarshals a YAML file into a Config struct
func LoadConfig(filePath string) (*Config, error) {
	// Ensure the file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("config file does not exist: %s", filePath)
	}

	// Read the file
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %v", err)
	}

	// Unmarshal YAML into struct
	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal YAML: %v", err)
	}

	return &config, nil
}

// SaveConfig marshals a Config struct to YAML and writes it to a file
func SaveConfig(filePath string, config *Config) error {
	// Marshal struct to YAML
	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config to YAML: %v", err)
	}

	// Ensure directory exists
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %v", err)
	}

	// Write to file
	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %v", err)
	}

	return nil
}

// GetUserConfigDir returns the path to the user's config directory
func GetUserConfigDir() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user config dir: %v", err)
	}
	return filepath.Join(configDir, "devgraph"), nil
}

// GetUserConfigPath returns the full path to the unified user config file
func GetUserConfigPath() (string, error) {
	configDir, err := GetUserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(configDir, "config.yaml"), nil
}

// LoadUserConfig loads the unified user configuration
func LoadUserConfig() (*UserConfig, error) {
	configPath, err := GetUserConfigPath()
	if err != nil {
		return nil, err
	}

	// If file doesn't exist, return empty config
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return &UserConfig{}, nil
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read user config: %v", err)
	}

	var userConfig UserConfig
	if err := yaml.Unmarshal(data, &userConfig); err != nil {
		return nil, fmt.Errorf("failed to unmarshal user config: %v", err)
	}

	return &userConfig, nil
}

// SaveUserConfig saves the unified user configuration
func SaveUserConfig(userConfig *UserConfig) error {
	configPath, err := GetUserConfigPath()
	if err != nil {
		return err
	}

	data, err := yaml.Marshal(userConfig)
	if err != nil {
		return fmt.Errorf("failed to marshal user config: %v", err)
	}

	// Ensure directory exists
	configDir := filepath.Dir(configPath)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %v", err)
	}

	if err := os.WriteFile(configPath, data, 0600); err != nil {
		return fmt.Errorf("failed to write user config: %v", err)
	}

	return nil
}

// ApplyUserSettings merges user settings into config, only if config values are defaults
func (c *Config) ApplyUserSettings(settings *UserSettings) {
	if settings.DefaultEnvironment != "" && c.Environment == "" {
		c.Environment = settings.DefaultEnvironment
	}
	if settings.DefaultModel != "" && c.Model == "gpt-4o-mini" {
		c.Model = settings.DefaultModel
	}
	if settings.DefaultMaxTokens > 0 && c.MaxTokens == 1000 {
		c.MaxTokens = settings.DefaultMaxTokens
	}
}

// LoadCredentials loads credentials from the unified config (for backward compatibility)
func LoadCredentials() (*Credentials, error) {
	userConfig, err := LoadUserConfig()
	if err != nil {
		return nil, err
	}
	return &userConfig.Credentials, nil
}

// SaveCredentials saves credentials to the unified config (for backward compatibility)
func SaveCredentials(creds Credentials) error {
	userConfig, err := LoadUserConfig()
	if err != nil {
		return err
	}
	userConfig.Credentials = creds
	return SaveUserConfig(userConfig)
}

// IsFirstTimeSetup checks if this is the first time the user is running devgraph
func IsFirstTimeSetup() bool {
	configPath, err := GetUserConfigPath()
	if err != nil {
		return true
	}

	// Check if config file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return true
	}

	// Check if config has any meaningful settings
	userConfig, err := LoadUserConfig()
	if err != nil {
		return true
	}

	// Consider it first-time if no settings and no credentials
	hasSettings := userConfig.Settings.DefaultEnvironment != "" ||
		userConfig.Settings.DefaultModel != "" ||
		userConfig.Settings.DefaultMaxTokens > 0

	hasCredentials := userConfig.Credentials.AccessToken != "" ||
		userConfig.Credentials.RefreshToken != "" ||
		userConfig.Credentials.IDToken != ""

	return !hasSettings && !hasCredentials
}
