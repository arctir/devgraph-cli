package auth

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/golang-jwt/jwt/v5"
	"gopkg.in/yaml.v3"
)

type Credentials struct {
	AccessToken  string         `yaml:"access_token,omitempty"`
	RefreshToken string         `yaml:"refresh_token,omitempty"`
	IDToken      string         `yaml:"id_token,omitempty"`
	Claims       *jwt.MapClaims `yaml:"claims,omitempty"`
}

func getCredentialsPath(appName string) (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user config dir: %v", err)
	}
	appDir := filepath.Join(configDir, appName)
	if err := os.MkdirAll(appDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create app dir: %v", err)
	}
	return filepath.Join(appDir, "credentials.yaml"), nil
}

func SaveCredentials(creds Credentials) error {
	filePath, err := getCredentialsPath("devgraph")
	if err != nil {
		return err
	}

	data, err := yaml.Marshal(creds)
	if err != nil {
		return err
	}
	return os.WriteFile(filePath, data, 0600)
}

func LoadCredentials() (*Credentials, error) {
	filepath, err := getCredentialsPath("devgraph")
	if err != nil {
		return nil, err
	}
	// Check if the file exists
	if _, err := os.Stat(filepath); os.IsNotExist(err) {
		return nil, fmt.Errorf("credentials file does not exist: %s", filepath)
	}

	var creds Credentials
	// Read the file
	data, err := os.ReadFile(filepath)
	if err != nil {
		return nil, fmt.Errorf("error reading YAML file: %v", err)
	}

	// Unmarshal into config
	if err := yaml.Unmarshal(data, &creds); err != nil {
		return nil, fmt.Errorf("error unmarshaling YAML: %v", err)
	}
	return &creds, nil
}
