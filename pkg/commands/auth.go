// Package commands provides command-line command implementations for Devgraph CLI.
package commands

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/arctir/devgraph-cli/pkg/auth"
	"github.com/arctir/devgraph-cli/pkg/config"
	"github.com/arctir/devgraph-cli/pkg/util"
	"github.com/golang-jwt/jwt/v5"
)

// AuthLoginCommand handles user authentication with Devgraph.
type AuthLoginCommand struct {
	config.Config
	Cluster      string `flag:"cluster" help:"API server URL (e.g., https://api.devgraph.ai, http://localhost:8000)"`
	Context      string `flag:"context" help:"Name for the context to create (defaults to auto-generated from API URL)"`
	SetAsCurrent bool   `flag:"set-current" default:"true" help:"Set as current context after login"`
	Relogin      bool   `flag:"relogin" help:"Re-authenticate to the current context's cluster instead of production"`
}

// AuthLogoutCommand handles user logout and credential cleanup.
type AuthLogoutCommand struct {
	config.Config
}

// AuthWhoamiCommand displays information about the currently authenticated user.
type AuthWhoamiCommand struct {
	config.Config
}

// AuthTokenCommand prints the user's authentication token to stdout.
type AuthTokenCommand struct {
	config.Config
}

// AuthCommand is the parent command for all authentication-related subcommands.
type AuthCommand struct {
	Login  *AuthLoginCommand  `cmd:"login" help:"Authenticate with your Devgraph account"`
	Logout *AuthLogoutCommand `cmd:"logout" help:"Log out and clear authentication credentials"`
	Whoami *AuthWhoamiCommand `cmd:"whoami" help:"Show information about the currently authenticated user"`
	Token  *AuthTokenCommand  `cmd:"token" help:"Print the authentication token to stdout"`
}

// Keep the old Auth struct for backward compatibility
type Auth struct {
	config.Config
}

func parseJWT(tokenString string) (*jwt.MapClaims, error) {
	token, _, err := new(jwt.Parser).ParseUnverified(tokenString, jwt.MapClaims{})
	if err != nil {
		fmt.Println("Error parsing token:", err)
		return nil, err
	}

	// Extract claims
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("failed to extract claims from token")
	}
	return &claims, nil
}

func (a *Auth) Run() error {
	// Step 1: Authenticate with OIDC
	token, err := auth.AuthenticatorImpl.Authenticate(a.Config)
	if err != nil {
		return err
	}

	claims, err := parseJWT(token.Extra("id_token").(string))
	if err != nil {
		fmt.Println("Error parsing JWT:", err)
	}

	creds := config.Credentials{
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
		IDToken:      token.Extra("id_token").(string),
		Claims:       claims,
	}

	// Extract username from email
	username := "user"
	var email, orgSlug string
	if claims != nil {
		if e, ok := (*claims)["email"].(string); ok && e != "" {
			email = e
			// Use email prefix as username
			if atIndex := len(email); atIndex > 0 {
				for i, c := range email {
					if c == '@' {
						username = email[:i]
						break
					}
				}
			}
		}
		if org, ok := (*claims)["org_slug"].(string); ok && org != "" {
			orgSlug = org
		}
	}

	fmt.Println("Authentication successful")
	if email != "" {
		fmt.Printf("Logged in as: %s\n", email)
	}
	if orgSlug != "" {
		fmt.Printf("Organization: %s\n", orgSlug)
	}

	return createOrUpdateContext(username, creds, a.Config, "", "", true)
}

// createOrUpdateContext handles context creation/update logic
// clusterURL is the API URL (or empty to use cfg.ApiURL)
func createOrUpdateContext(username string, creds config.Credentials, cfg config.Config, clusterURL, contextName string, setAsCurrent bool) error {
	// Step 2: Create or update context
	userConfig, err := config.LoadUserConfig()
	if err != nil {
		userConfig = &config.UserConfig{}
	}

	// Use provided API URL or fall back to config
	apiURL := cfg.ApiURL
	if clusterURL != "" {
		apiURL = clusterURL
	}

	// Derive cluster name from API URL using EnvironmentConfigMap
	clusterName := "default"
	knownURL := false
	if apiURL != "" {
		for envName, envConfig := range config.EnvironmentConfigMap {
			if envConfig.ApiURL == apiURL {
				clusterName = envName
				knownURL = true
				break
			}
		}

		// For unknown URLs, use provided context name or prompt user
		if !knownURL {
			if contextName != "" {
				// User provided a context name, use it as cluster name too
				clusterName = contextName
			} else {
				// Prompt user for a name
				reader := bufio.NewReader(os.Stdin)
				fmt.Printf("\nUnknown API URL: %s\n", apiURL)
				fmt.Print("Enter a name for this context: ")
				input, _ := reader.ReadString('\n')
				input = strings.TrimSpace(input)
				if input == "" {
					// Fall back to sanitized URL
					clusterName = sanitizeURLForName(apiURL)
				} else {
					clusterName = input
					contextName = input // Use the same name for context
				}
			}
		}
	}

	// Create cluster if it doesn't exist
	if _, exists := userConfig.Clusters[clusterName]; !exists {
		userConfig.SetCluster(clusterName, apiURL, cfg.IssuerURL, cfg.ClientID)
	}

	// Save user credentials
	userConfig.SetUser(username, creds.AccessToken, creds.RefreshToken, creds.IDToken, creds.Claims)

	// Create context name if not provided
	if contextName == "" {
		contextName = clusterName + "-context"
	}

	// Create or update context
	userConfig.SetContext(contextName, clusterName, username, "")

	// Set as current context based on flag
	if setAsCurrent || userConfig.CurrentContext == "" {
		userConfig.CurrentContext = contextName
		fmt.Printf("Context '%s' set as current\n", contextName)
	} else {
		fmt.Printf("Context '%s' created\n", contextName)
		fmt.Printf("Switch to it with: dg config use-context %s\n", contextName)
	}

	// Save config
	err = config.SaveUserConfig(userConfig)
	if err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	return nil
}

// sanitizeURLForName converts a URL to a safe cluster name
func sanitizeURLForName(url string) string {
	// Remove protocol
	name := url
	name = strings.TrimPrefix(name, "https://")
	name = strings.TrimPrefix(name, "http://")
	// Remove port
	if idx := strings.Index(name, ":"); idx != -1 {
		name = name[:idx]
	}
	// Replace dots and slashes with dashes
	name = strings.ReplaceAll(name, ".", "-")
	name = strings.ReplaceAll(name, "/", "-")
	name = strings.Trim(name, "-")
	return name
}

func (a *AuthLoginCommand) Run() error {
	// Determine which cluster to authenticate to
	if a.Relogin {
		// Re-authenticate to current context's cluster
		a.Config.ApplyDefaults()
		if a.Config.ApiURL == "" {
			return fmt.Errorf("no current context to relogin to. Use 'dg auth login' to authenticate")
		}
	} else if a.Cluster != "" {
		// Use explicitly specified cluster
		a.Config.ApiURL = a.Cluster
		fmt.Printf("Fetching OIDC configuration from %s...\n", a.Cluster)
		issuerURL, clientID, err := config.FetchOIDCConfig(a.Cluster)
		if err != nil {
			return fmt.Errorf("failed to fetch OIDC configuration: %w", err)
		}
		a.Config.IssuerURL = issuerURL
		a.Config.ClientID = clientID
	} else {
		// Default to production
		prodConfig := config.EnvironmentConfigMap["production"]
		a.Config.ApiURL = prodConfig.ApiURL
		fmt.Printf("Fetching OIDC configuration from %s...\n", prodConfig.ApiURL)
		issuerURL, clientID, err := config.FetchOIDCConfig(prodConfig.ApiURL)
		if err != nil {
			// Fall back to hardcoded values if API is unavailable
			fmt.Printf("⚠️  Could not fetch OIDC config from API, using defaults: %v\n", err)
			a.Config.IssuerURL = prodConfig.IssuerURL
			a.Config.ClientID = prodConfig.ClientID
		} else {
			a.Config.IssuerURL = issuerURL
			a.Config.ClientID = clientID
		}
	}

	// Step 1: Authenticate with OIDC
	token, err := auth.AuthenticatorImpl.Authenticate(a.Config)
	if err != nil {
		return err
	}

	claims, err := parseJWT(token.Extra("id_token").(string))
	if err != nil {
		fmt.Println("Error parsing JWT:", err)
	}

	creds := config.Credentials{
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
		IDToken:      token.Extra("id_token").(string),
		Claims:       claims,
	}

	// Print success banner
	fmt.Println("============================================================")
	fmt.Println("✅ Authentication Successful!")
	fmt.Println("============================================================")
	fmt.Printf("🌐 Cluster: %s\n", a.Config.ApiURL)

	// Extract username from email
	username := "user"
	if claims != nil {
		if email, ok := (*claims)["email"].(string); ok && email != "" {
			fmt.Printf("👤 Logged in as: %s\n", email)
			// Use email prefix as username
			if atIndex := len(email); atIndex > 0 {
				for i, c := range email {
					if c == '@' {
						username = email[:i]
						break
					}
				}
			}
		}
		if orgSlug, ok := (*claims)["org_slug"].(string); ok && orgSlug != "" {
			fmt.Printf("🏢 Organization: %s\n", orgSlug)
		}
	}

	// Pass API URL as clusterURL parameter (was clusterName)
	err = createOrUpdateContext(username, creds, a.Config, a.Cluster, a.Context, a.SetAsCurrent)
	if err != nil {
		return err
	}

	// Auto-configure environment after successful login
	fmt.Println("🌍 Setting up your environment...")
	if err := configureEnvironmentAfterLogin(a.Config); err != nil {
		fmt.Printf("⚠️  Could not configure environment: %v\n", err)
		fmt.Println("   You can set it later with: dg config set-context <name> --env <env>")
	}
	fmt.Println()

	return nil
}

// configureEnvironmentAfterLogin fetches environments and sets on current context
func configureEnvironmentAfterLogin(cfg config.Config) error {
	envs, err := util.GetEnvironments(cfg)
	if err != nil {
		return fmt.Errorf("failed to get environments: %w", err)
	}

	userConfig, err := config.LoadUserConfig()
	if err != nil {
		return fmt.Errorf("failed to load user config: %w", err)
	}

	if userConfig.CurrentContext == "" {
		return fmt.Errorf("no current context set")
	}

	currentCtx, ok := userConfig.Contexts[userConfig.CurrentContext]
	if !ok {
		return fmt.Errorf("current context '%s' not found", userConfig.CurrentContext)
	}

	// Clear any existing environment - it may be from a different cluster
	currentCtx.Environment = ""

	if envs == nil || len(*envs) == 0 {
		fmt.Println("No environments found. You may need to create one first.")
		return config.SaveUserConfig(userConfig)
	}

	var selectedEnvID string
	var selectedEnvName string

	// Auto-select if only one environment
	if len(*envs) == 1 {
		env := (*envs)[0]
		selectedEnvID = env.ID.String()
		selectedEnvName = env.Name
	} else {
		// Prompt user to select
		fmt.Println("Available environments:")
		for i, env := range *envs {
			fmt.Printf("  %d. %s (%s)\n", i+1, env.Name, env.Slug)
		}

		reader := bufio.NewReader(os.Stdin)
		for {
			fmt.Print("\nSelect an environment (enter number): ")
			input, _ := reader.ReadString('\n')
			input = strings.TrimSpace(input)

			choice, err := strconv.Atoi(input)
			if err != nil || choice < 1 || choice > len(*envs) {
				fmt.Printf("Invalid choice. Please enter a number between 1 and %d.\n", len(*envs))
				continue
			}

			selectedEnv := (*envs)[choice-1]
			selectedEnvID = selectedEnv.ID.String()
			selectedEnvName = selectedEnv.Name
			break
		}
	}

	// Set environment on current context
	currentCtx.Environment = selectedEnvID
	// Also set in settings for backward compatibility
	userConfig.Settings.DefaultEnvironment = selectedEnvID

	fmt.Printf("✅ Environment set to: %s\n", selectedEnvName)
	return config.SaveUserConfig(userConfig)
}

func (a *AuthLogoutCommand) Run() error {
	return auth.Logout(a.Config)
}

func (a *AuthWhoamiCommand) Run() error {
	userInfo, err := auth.GetCurrentUser()
	if err != nil {
		return fmt.Errorf("failed to get user information: %w", err)
	}

	// Display user information with a cleaner format
	if userInfo.Name != "" {
		fmt.Printf("Logged in as %s", userInfo.Name)
	} else if userInfo.GivenName != "" || userInfo.FamilyName != "" {
		fmt.Printf("Logged in as %s %s", userInfo.GivenName, userInfo.FamilyName)
	}

	if userInfo.Email != "" {
		fmt.Printf(" (%s)", userInfo.Email)
	}
	fmt.Println()

	if userInfo.OrganizationSlug != "" {
		fmt.Printf("Organization: %s", userInfo.OrganizationSlug)
		if userInfo.OrganizationID != "" {
			fmt.Printf(" (%s)", userInfo.OrganizationID)
		}
		fmt.Println()
	}

	// Show current environment
	userConfig, err := config.LoadUserConfig()
	if err == nil && userConfig.Settings.DefaultEnvironment != "" {
		fmt.Printf("Environment: %s\n", userConfig.Settings.DefaultEnvironment)
	}

	return nil
}

func (a *AuthTokenCommand) Run() error {
	creds, err := auth.LoadCredentials()
	if err != nil {
		return fmt.Errorf("failed to load credentials: %w", err)
	}

	fmt.Println(creds.IDToken)
	return nil
}
