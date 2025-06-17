package main

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

const DefaultIssuerURL = "https://primary-ghoul-65.clerk.accounts.dev"
const DefaultClientID = "renbud3BkDcW1utM"
const DefaultRedirectURL = "http://localhost:8080/callback"

type Config struct {
	ApiURL      string `kong:"default='http://localhost:8000/api/v1/model',env='DEVGRAPH_API_URL',help='Devgraph API URL'"`
	IssuerURL   string `kong:"default='https://primary-ghoul-65.clerk.accounts.dev',env='DEVGRAPH_ISSUER_URL',help='Devgraph issuer URL'"`
	ClientID    string `kong:"default='renbud3BkDcW1utM',env='DEVGRAPH_CLIENT_ID',help='Devgraph client ID'"`
	RedirectURL string `kong:"default='http://localhost:8080/callback',env='DEVGRAPH_REDIRECT_URL',help='Redirect URL'"`

	Model     string `kong:"default='gpt-3.5-turbo',short='m',help='OpenAI model to use'"`
	MaxTokens int    `kong:"default=1000,short='t',help='Maximum number of tokens in response'"`
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
