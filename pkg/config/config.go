// Package config provides configuration management for Devgraph CLI.
// It handles loading and saving user configuration, credentials management,
// and provides default values for CLI operations.
package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"gopkg.in/yaml.v3"
)

// EnvironmentConfig holds all configuration for an environment
type EnvironmentConfig struct {
	ApiURL    string
	IssuerURL string
	ClientID  string
}

// EnvironmentConfigMap maps environment names to their configurations
// TODO: Replace with API endpoint call to fetch these dynamically
var EnvironmentConfigMap = map[string]EnvironmentConfig{
	"staging": {
		ApiURL:    "https://api.staging.devgraph.ai",
		IssuerURL: "https://primary-ghoul-65.clerk.accounts.dev",
		ClientID:  "I97zD0IQmSFr5pql",
	},
	"production": {
		ApiURL:    "https://api.devgraph.ai",
		IssuerURL: "https://clerk.devgraph.ai",
		ClientID:  "mbcVQVf8ZF0Q2Dki",
	},
}

// Config represents the runtime configuration for Devgraph CLI operations.
// It combines command-line flags, environment variables, and user settings.
type Config struct {
	// These fields are populated from the current context's cluster (hidden from CLI)
	ApiURL    string `kong:"-"`
	IssuerURL string `kong:"-"`
	ClientID  string `kong:"-"`

	// Debug enables verbose HTTP request/response logging
	Debug bool `kong:"short='d',help='Enable debug logging (HTTP requests/responses)'"`
}

// ApplyDefaults populates the API/OAuth fields from the current context's cluster
// Falls back to staging environment config if no context is configured
func (c *Config) ApplyDefaults() {
	// Try to load from current context's cluster
	userConfig, err := LoadUserConfig()
	if err == nil && userConfig.CurrentContext != "" {
		_, cluster, _, err := userConfig.GetCurrentContext()
		if err == nil && cluster != nil {
			c.ApiURL = cluster.Server
			c.IssuerURL = cluster.IssuerURL
			c.ClientID = cluster.ClientID
			return
		}
	}

	// Fallback to production defaults for first-time setup
	envConfig := EnvironmentConfigMap["production"]
	c.ApiURL = envConfig.ApiURL
	c.IssuerURL = envConfig.IssuerURL
	c.ClientID = envConfig.ClientID
}

// UserConfig represents the unified user configuration file
type UserConfig struct {
	// User preferences
	Settings UserSettings `yaml:"settings,omitempty"`

	// Authentication credentials (deprecated - use contexts instead)
	Credentials Credentials `yaml:"credentials,omitempty"`

	// Contexts (kubectl-style)
	Contexts       map[string]*Context `yaml:"contexts,omitempty"`
	Clusters       map[string]*Cluster `yaml:"clusters,omitempty"`
	Users          map[string]*User    `yaml:"users,omitempty"`
	CurrentContext string              `yaml:"current-context,omitempty"`
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

// Context defines a context (cluster + user + environment)
type Context struct {
	Cluster     string `yaml:"cluster"`
	User        string `yaml:"user"`
	Environment string `yaml:"environment,omitempty"` // UUID of the environment
}

// Cluster defines an API server/cluster
type Cluster struct {
	Server    string `yaml:"server"`     // API URL
	IssuerURL string `yaml:"issuer-url"` // OIDC issuer URL
	ClientID  string `yaml:"client-id"`  // OAuth client ID
}

// User defines authentication credentials for a user
type User struct {
	AccessToken  string         `yaml:"access-token,omitempty"`
	RefreshToken string         `yaml:"refresh-token,omitempty"`
	IDToken      string         `yaml:"id-token,omitempty"`
	Claims       *jwt.MapClaims `yaml:"claims,omitempty"`
}

// LoadConfig reads and unmarshals a YAML file into a Config struct
// validateConfigPath ensures the file path is safe to read
func validateConfigPath(filePath string) error {
	// Clean the path to prevent directory traversal
	cleanPath := filepath.Clean(filePath)

	// Ensure no directory traversal attempts
	if strings.Contains(cleanPath, "..") {
		return fmt.Errorf("invalid file path: directory traversal detected")
	}

	// Ensure it's a reasonable config file extension
	ext := filepath.Ext(cleanPath)
	if ext != ".yaml" && ext != ".yml" && ext != ".json" {
		return fmt.Errorf("invalid file path: unsupported config file type")
	}

	return nil
}

func LoadConfig(filePath string) (*Config, error) {
	// Validate the file path for security
	if err := validateConfigPath(filePath); err != nil {
		return nil, err
	}

	// Ensure the file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("config file does not exist: %s", filePath)
	}

	// Read the file
	data, err := os.ReadFile(filePath) // #nosec G304 - path validated above
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Unmarshal YAML into struct
	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal YAML: %w", err)
	}

	return &config, nil
}

// SaveConfig marshals a Config struct to YAML and writes it to a file
func SaveConfig(filePath string, config *Config) error {
	// Marshal struct to YAML
	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config to YAML: %w", err)
	}

	// Ensure directory exists with secure permissions
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0750); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Write to file with secure permissions
	if err := os.WriteFile(filePath, data, 0600); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// GetUserConfigDir returns the path to the user's config directory
func GetUserConfigDir() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user config dir: %w", err)
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

	data, err := os.ReadFile(configPath) // #nosec G304 - path from GetUserConfigPath() is safe
	if err != nil {
		return nil, fmt.Errorf("failed to read user config: %w", err)
	}

	var userConfig UserConfig
	if err := yaml.Unmarshal(data, &userConfig); err != nil {
		return nil, fmt.Errorf("failed to unmarshal user config: %w", err)
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
		return fmt.Errorf("failed to marshal user config: %w", err)
	}

	// Ensure directory exists with secure permissions
	configDir := filepath.Dir(configPath)
	if err := os.MkdirAll(configDir, 0750); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0600); err != nil {
		return fmt.Errorf("failed to write user config: %w", err)
	}

	return nil
}


// LoadCredentials loads credentials from the unified config (for backward compatibility)
// Now supports loading from contexts
func LoadCredentials() (*Credentials, error) {
	userConfig, err := LoadUserConfig()
	if err != nil {
		return nil, err
	}
	return userConfig.GetCredentialsFromContext()
}

// SaveCredentials saves credentials to the unified config (for backward compatibility)
// If a context is active, updates the user in that context; otherwise updates legacy credentials
func SaveCredentials(creds Credentials) error {
	userConfig, err := LoadUserConfig()
	if err != nil {
		return err
	}

	// If using contexts, update the current context's user
	if userConfig.CurrentContext != "" {
		context, _, _, err := userConfig.GetCurrentContext()
		if err == nil {
			// Update the user associated with this context
			userConfig.SetUser(context.User, creds.AccessToken, creds.RefreshToken, creds.IDToken, creds.Claims)
		}
	}

	// Also save to legacy credentials for backward compatibility
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

	hasContexts := len(userConfig.Contexts) > 0 && len(userConfig.Users) > 0

	return !hasSettings && !hasCredentials && !hasContexts
}

// GetCurrentContext returns the current context configuration
func (uc *UserConfig) GetCurrentContext() (*Context, *Cluster, *User, error) {
	if uc.CurrentContext == "" {
		return nil, nil, nil, fmt.Errorf("no current context set")
	}

	context, ok := uc.Contexts[uc.CurrentContext]
	if !ok {
		return nil, nil, nil, fmt.Errorf("current context '%s' not found", uc.CurrentContext)
	}

	cluster, ok := uc.Clusters[context.Cluster]
	if !ok {
		return nil, nil, nil, fmt.Errorf("cluster '%s' not found for context '%s'", context.Cluster, uc.CurrentContext)
	}

	user, ok := uc.Users[context.User]
	if !ok {
		return nil, nil, nil, fmt.Errorf("user '%s' not found for context '%s'", context.User, uc.CurrentContext)
	}

	return context, cluster, user, nil
}

// SetContext creates or updates a context
func (uc *UserConfig) SetContext(name string, cluster, user, environment string) {
	if uc.Contexts == nil {
		uc.Contexts = make(map[string]*Context)
	}
	uc.Contexts[name] = &Context{
		Cluster:     cluster,
		User:        user,
		Environment: environment,
	}
}

// SetCluster creates or updates a cluster
func (uc *UserConfig) SetCluster(name string, server, issuerURL, clientID string) {
	if uc.Clusters == nil {
		uc.Clusters = make(map[string]*Cluster)
	}
	uc.Clusters[name] = &Cluster{
		Server:    server,
		IssuerURL: issuerURL,
		ClientID:  clientID,
	}
}

// SetUser creates or updates a user with credentials
func (uc *UserConfig) SetUser(name string, accessToken, refreshToken, idToken string, claims *jwt.MapClaims) {
	if uc.Users == nil {
		uc.Users = make(map[string]*User)
	}
	uc.Users[name] = &User{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		IDToken:      idToken,
		Claims:       claims,
	}
}

// UseContext sets the current context
func (uc *UserConfig) UseContext(name string) error {
	if _, ok := uc.Contexts[name]; !ok {
		return fmt.Errorf("context '%s' not found", name)
	}
	uc.CurrentContext = name
	return nil
}

// DeleteContext removes a context
func (uc *UserConfig) DeleteContext(name string) error {
	if _, ok := uc.Contexts[name]; !ok {
		return fmt.Errorf("context '%s' not found", name)
	}
	delete(uc.Contexts, name)
	if uc.CurrentContext == name {
		uc.CurrentContext = ""
	}
	return nil
}

// GetCredentialsFromContext returns credentials from the current context
// Falls back to legacy credentials if no context is set
func (uc *UserConfig) GetCredentialsFromContext() (*Credentials, error) {
	// Try to get credentials from current context
	if uc.CurrentContext != "" {
		_, _, user, err := uc.GetCurrentContext()
		if err == nil && user != nil {
			return &Credentials{
				AccessToken:  user.AccessToken,
				RefreshToken: user.RefreshToken,
				IDToken:      user.IDToken,
				Claims:       user.Claims,
			}, nil
		}
	}

	// Fall back to legacy credentials
	if uc.Credentials.AccessToken != "" || uc.Credentials.IDToken != "" {
		return &uc.Credentials, nil
	}

	return nil, fmt.Errorf("no credentials found")
}
