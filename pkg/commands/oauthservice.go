package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/arctir/devgraph-cli/pkg/util"
	api "github.com/arctir/go-devgraph/pkg/apis/devgraph/v1"
	"github.com/google/uuid"
	"github.com/int128/oauth2cli"
	"golang.org/x/oauth2"
)

type OAuthServiceCommand struct {
	Create    OAuthServiceCreateCommand    `cmd:"create" help:"Create a new OAuth service."`
	Get       OAuthServiceGetCommand       `cmd:"get" help:"Retrieve an OAuth service by ID."`
	List      OAuthServiceListCommand      `cmd:"" help:"List OAuth services."`
	Delete    OAuthServiceDeleteCommand    `cmd:"delete" help:"Delete an OAuth service by ID."`
	Update    OAuthServiceUpdateCommand    `cmd:"update" help:"Update an OAuth service by ID."`
	Authorize OAuthServiceAuthorizeCommand `cmd:"authorize" help:"Authorize against an OAuth provider."`
}

type OAuthServiceCreateCommand struct {
	EnvWrapperCommand
	Name                string   `arg:"" required:"" help:"Unique name identifier for the OAuth service."`
	DisplayName         string   `arg:"" required:"" help:"Human-readable display name for the OAuth service."`
	OAuthClientID       string   `arg:"" required:"" help:"OAuth client ID."`
	OAuthClientSecret   string   `arg:"" required:"" help:"OAuth client secret."`
	AuthorizationURL    string   `arg:"" required:"" help:"OAuth authorization endpoint URL."`
	TokenURL            string   `arg:"" required:"" help:"OAuth token endpoint URL."`
	SupportedGrantTypes []string `arg:"" optional:"" help:"Supported OAuth grant types (defaults to: authorization_code)."`
	Description         *string  `flag:"description" optional:"" help:"Optional description of the OAuth service."`
	UserinfoURL         *string  `flag:"userinfo-url" optional:"" help:"Optional userinfo endpoint URL."`
	DefaultScopes       []string `flag:"default-scopes" optional:"" help:"Optional default OAuth scopes."`
	IsActive            *bool    `flag:"is-active" optional:"" help:"Whether the OAuth service is active."`
	IconURL             *string  `flag:"icon-url" optional:"" help:"Optional icon URL."`
	HomepageURL         *string  `flag:"homepage-url" optional:"" help:"Optional homepage URL."`
}

type OAuthServiceListCommand struct {
	EnvWrapperCommand
	ActiveOnly *bool `flag:"active-only" optional:"" help:"Only return active services."`
}

type OAuthServiceGetCommand struct {
	EnvWrapperCommand
	ID string `arg:"" required:"" help:"ID of the OAuth service to retrieve."`
}

type OAuthServiceDeleteCommand struct {
	EnvWrapperCommand
	ID string `arg:"" required:"" help:"ID of the OAuth service to delete."`
}

type OAuthServiceUpdateCommand struct {
	EnvWrapperCommand
	ID                  string   `arg:"" required:"" help:"ID of the OAuth service to update."`
	DisplayName         *string  `flag:"update-display-name" optional:"" help:"Human-readable display name for the OAuth service."`
	OAuthClientID       *string  `flag:"update-oauth-client-id" optional:"" help:"OAuth client ID."`
	OAuthClientSecret   *string  `flag:"update-oauth-client-secret" optional:"" help:"OAuth client secret."`
	AuthorizationURL    *string  `flag:"update-authorization-url" optional:"" help:"OAuth authorization endpoint URL."`
	TokenURL            *string  `flag:"update-token-url" optional:"" help:"OAuth token endpoint URL."`
	SupportedGrantTypes []string `flag:"update-supported-grant-types" optional:"" help:"Supported OAuth grant types (e.g., authorization_code)."`
	Description         *string  `flag:"update-description" optional:"" help:"Description of the OAuth service."`
	UserinfoURL         *string  `flag:"update-userinfo-url" optional:"" help:"Userinfo endpoint URL."`
	DefaultScopes       []string `flag:"update-default-scopes" optional:"" help:"Default OAuth scopes."`
	IsActive            *bool    `flag:"update-is-active" optional:"" help:"Whether the OAuth service is active."`
	IconURL             *string  `flag:"update-icon-url" optional:"" help:"Icon URL."`
	HomepageURL         *string  `flag:"update-homepage-url" optional:"" help:"Homepage URL."`
}

type OAuthServiceAuthorizeCommand struct {
	EnvWrapperCommand
	ServiceID    string   `arg:"" required:"" help:"ID of the OAuth service to authorize against."`
	ClientID     string   `arg:"" required:"" help:"OAuth client ID (not returned from API for security)."`
	ClientSecret string   `arg:"" required:"" help:"OAuth client secret (not stored for security)."`
	Scopes       []string `flag:"scopes" optional:"" help:"OAuth scopes to request (uses service defaults if not specified)."`
	RedirectPort *int     `flag:"redirect-port" default:"40000" help:"Local port for OAuth callback (default: 40000)."`
}

func (c *OAuthServiceCreateCommand) Run() error {
	// Set default grant type if none provided
	if len(c.SupportedGrantTypes) == 0 {
		c.SupportedGrantTypes = []string{"authorization_code"}
	}

	client, err := util.GetAuthenticatedClient(c.Config)
	if err != nil {
		return fmt.Errorf("failed to create authenticated client: %w", err)
	}

	// Parse required URLs
	authURL, err := url.Parse(c.AuthorizationURL)
	if err != nil {
		return fmt.Errorf("invalid authorization URL: %w", err)
	}

	tokenURL, err := url.Parse(c.TokenURL)
	if err != nil {
		return fmt.Errorf("invalid token URL: %w", err)
	}

	// Create the OAuth service request
	oauthService := api.OAuthServiceCreate{
		Name:                c.Name,
		DisplayName:         c.DisplayName,
		ClientID:            c.OAuthClientID,
		ClientSecret:        c.OAuthClientSecret,
		AuthorizationURL:    *authURL,
		TokenURL:            *tokenURL,
		SupportedGrantTypes: c.SupportedGrantTypes,
	}

	// Set optional fields
	if c.Description != nil {
		oauthService.SetDescription(api.NewOptNilString(*c.Description))
	}

	if c.UserinfoURL != nil {
		userinfoURL, err := url.Parse(*c.UserinfoURL)
		if err != nil {
			return fmt.Errorf("invalid userinfo URL: %w", err)
		}
		oauthService.SetUserinfoURL(api.NewOptNilURI(*userinfoURL))
	}

	if c.DefaultScopes != nil {
		oauthService.SetDefaultScopes(api.NewOptNilStringArray(c.DefaultScopes))
	}

	if c.IsActive != nil {
		oauthService.SetIsActive(api.NewOptBool(*c.IsActive))
	}

	if c.IconURL != nil {
		iconURL, err := url.Parse(*c.IconURL)
		if err != nil {
			return fmt.Errorf("invalid icon URL: %w", err)
		}
		oauthService.SetIconURL(api.NewOptNilURI(*iconURL))
	}

	if c.HomepageURL != nil {
		homepageURL, err := url.Parse(*c.HomepageURL)
		if err != nil {
			return fmt.Errorf("invalid homepage URL: %w", err)
		}
		oauthService.SetHomepageURL(api.NewOptNilURI(*homepageURL))
	}

	// Make the API call
	response, err := client.CreateOAuthService(context.TODO(), &oauthService)
	if err != nil {
		return fmt.Errorf("failed to create oauth service: %w", err)
	}

	// Handle response
	switch r := response.(type) {
	case *api.OAuthServiceResponse:
		fmt.Printf("✅ OAuth service '%s' created successfully with ID: %s\n", c.Name, r.GetID())
	default:
		return fmt.Errorf("failed to create oauth service")
	}

	return nil
}

func (c *OAuthServiceListCommand) Run() error {
	client, err := util.GetAuthenticatedClient(c.Config)
	if err != nil {
		return fmt.Errorf("failed to create authenticated client: %w", err)
	}

	// Set up parameters
	params := api.ListOAuthServicesParams{}
	if c.ActiveOnly != nil {
		params.ActiveOnly = api.NewOptBool(*c.ActiveOnly)
	}

	// Make the API call
	response, err := client.ListOAuthServices(context.Background(), params)
	if err != nil {
		return fmt.Errorf("failed to list oauth services: %w", err)
	}

	// Handle response
	switch r := response.(type) {
	case *api.OAuthServiceListResponse:
		if len(r.Services) == 0 {
			fmt.Println("No OAuth services found.")
			return nil
		}
		displayOAuthServices(&r.Services)
	default:
		return fmt.Errorf("failed to list oauth services")
	}

	return nil
}

func (c *OAuthServiceGetCommand) Run() error {
	client, err := util.GetAuthenticatedClient(c.Config)
	if err != nil {
		return fmt.Errorf("failed to create authenticated client: %w", err)
	}

	// Parse UUID
	serviceID, err := uuid.Parse(c.ID)
	if err != nil {
		return fmt.Errorf("invalid UUID: %w", err)
	}

	params := api.GetOAuthServiceParams{
		ServiceID: serviceID,
	}

	// Make the API call
	response, err := client.GetOAuthService(context.Background(), params)
	if err != nil {
		return fmt.Errorf("failed to get oauth service: %w", err)
	}

	// Handle response
	switch r := response.(type) {
	case *api.OAuthServiceResponse:
		services := []api.OAuthServiceResponse{*r}
		displayOAuthServices(&services)
	default:
		return fmt.Errorf("oauth service with ID '%s' not found", c.ID)
	}

	return nil
}

func (c *OAuthServiceDeleteCommand) Run() error {
	client, err := util.GetAuthenticatedClient(c.Config)
	if err != nil {
		return fmt.Errorf("failed to create authenticated client: %w", err)
	}

	// Parse UUID
	serviceID, err := uuid.Parse(c.ID)
	if err != nil {
		return fmt.Errorf("invalid UUID: %w", err)
	}

	params := api.DeleteOAuthServiceParams{
		ServiceID: serviceID,
	}

	// Make the API call
	response, err := client.DeleteOAuthService(context.Background(), params)
	if err != nil {
		return fmt.Errorf("failed to delete oauth service: %w", err)
	}

	// Handle response
	switch response.(type) {
	case *api.DeleteOAuthServiceNoContent:
		fmt.Printf("✅ OAuth service '%s' deleted successfully.\n", c.ID)
	default:
		return fmt.Errorf("failed to delete oauth service")
	}

	return nil
}

func (c *OAuthServiceUpdateCommand) Run() error {
	client, err := util.GetAuthenticatedClient(c.Config)
	if err != nil {
		return fmt.Errorf("failed to create authenticated client: %w", err)
	}

	// Parse UUID
	serviceID, err := uuid.Parse(c.ID)
	if err != nil {
		return fmt.Errorf("invalid UUID: %w", err)
	}

	// Create update request with only set fields
	oauthUpdate := api.OAuthServiceUpdate{}

	if c.DisplayName != nil {
		oauthUpdate.SetDisplayName(api.NewOptNilString(*c.DisplayName))
	}
	if c.OAuthClientID != nil {
		oauthUpdate.SetClientID(api.NewOptNilString(*c.OAuthClientID))
	}
	if c.OAuthClientSecret != nil {
		oauthUpdate.SetClientSecret(api.NewOptNilString(*c.OAuthClientSecret))
	}
	if c.AuthorizationURL != nil {
		authURL, err := url.Parse(*c.AuthorizationURL)
		if err != nil {
			return fmt.Errorf("invalid authorization URL: %w", err)
		}
		oauthUpdate.SetAuthorizationURL(api.NewOptNilURI(*authURL))
	}
	if c.TokenURL != nil {
		tokenURL, err := url.Parse(*c.TokenURL)
		if err != nil {
			return fmt.Errorf("invalid token URL: %w", err)
		}
		oauthUpdate.SetTokenURL(api.NewOptNilURI(*tokenURL))
	}
	if c.SupportedGrantTypes != nil {
		oauthUpdate.SetSupportedGrantTypes(api.NewOptNilStringArray(c.SupportedGrantTypes))
	}
	if c.Description != nil {
		oauthUpdate.SetDescription(api.NewOptNilString(*c.Description))
	}
	if c.UserinfoURL != nil {
		userinfoURL, err := url.Parse(*c.UserinfoURL)
		if err != nil {
			return fmt.Errorf("invalid userinfo URL: %w", err)
		}
		oauthUpdate.SetUserinfoURL(api.NewOptNilURI(*userinfoURL))
	}
	if c.DefaultScopes != nil {
		oauthUpdate.SetDefaultScopes(api.NewOptNilStringArray(c.DefaultScopes))
	}
	if c.IsActive != nil {
		oauthUpdate.SetIsActive(api.NewOptNilBool(*c.IsActive))
	}
	if c.IconURL != nil {
		iconURL, err := url.Parse(*c.IconURL)
		if err != nil {
			return fmt.Errorf("invalid icon URL: %w", err)
		}
		oauthUpdate.SetIconURL(api.NewOptNilURI(*iconURL))
	}
	if c.HomepageURL != nil {
		homepageURL, err := url.Parse(*c.HomepageURL)
		if err != nil {
			return fmt.Errorf("invalid homepage URL: %w", err)
		}
		oauthUpdate.SetHomepageURL(api.NewOptNilURI(*homepageURL))
	}

	params := api.UpdateOAuthServiceParams{
		ServiceID: serviceID,
	}

	// Make the API call
	response, err := client.UpdateOAuthService(context.TODO(), &oauthUpdate, params)
	if err != nil {
		return fmt.Errorf("failed to update oauth service: %w", err)
	}

	// Handle response
	switch response.(type) {
	case *api.OAuthServiceResponse:
		fmt.Printf("✅ OAuth service '%s' updated successfully.\n", c.ID)
	default:
		return fmt.Errorf("failed to update oauth service")
	}

	return nil
}

func (c *OAuthServiceAuthorizeCommand) Run() error {
	client, err := util.GetAuthenticatedClient(c.Config)
	if err != nil {
		return fmt.Errorf("failed to create authenticated client: %w", err)
	}

	// Parse UUID
	serviceID, err := uuid.Parse(c.ServiceID)
	if err != nil {
		return fmt.Errorf("invalid service ID: %w", err)
	}

	// Get the OAuth service details
	params := api.GetOAuthServiceParams{
		ServiceID: serviceID,
	}

	response, err := client.GetOAuthService(context.Background(), params)
	if err != nil {
		return fmt.Errorf("failed to get oauth service: %w", err)
	}

	var service *api.OAuthServiceResponse
	switch r := response.(type) {
	case *api.OAuthServiceResponse:
		service = r
	default:
		return fmt.Errorf("oauth service with ID '%s' not found", c.ServiceID)
	}

	// Determine scopes to use
	scopes := c.Scopes
	if len(scopes) == 0 && len(service.DefaultScopes) > 0 {
		scopes = service.DefaultScopes
		fmt.Printf("Using service default scopes: %s\n", strings.Join(scopes, ", "))
	}

	// Set up OAuth2 config with service details and provided credentials
	oauth2Config := oauth2.Config{
		ClientID:     c.ClientID,
		ClientSecret: c.ClientSecret,
		Endpoint: oauth2.Endpoint{
			AuthURL:  service.AuthorizationURL,
			TokenURL: service.TokenURL,
		},
		Scopes: scopes,
	}

	// Determine redirect port
	redirectPort := 40000
	if c.RedirectPort != nil {
		redirectPort = *c.RedirectPort
	}

	// Set up OAuth2 CLI for local callback handling
	oauth2Config.RedirectURL = fmt.Sprintf("http://localhost:%d/callback", redirectPort)

	// Start OAuth flow
	fmt.Printf("Starting OAuth authorization for service '%s'...\n", service.DisplayName)
	fmt.Printf("Service ID: %s\n", service.ID)
	fmt.Printf("Client ID: %s\n", c.ClientID)
	fmt.Printf("Authorization URL: %s\n", service.AuthorizationURL)
	fmt.Printf("Redirect URL: %s\n", oauth2Config.RedirectURL)
	if len(scopes) > 0 {
		fmt.Printf("Requested Scopes: %s\n", strings.Join(scopes, ", "))
	}

	// Use oauth2cli to handle the flow
	token, err := oauth2cli.GetToken(context.Background(), oauth2cli.Config{
		OAuth2Config: oauth2Config,
		LocalServerBindAddress: []string{
			fmt.Sprintf("127.0.0.1:%d", redirectPort),
			fmt.Sprintf("::1:%d", redirectPort),
		},
		LocalServerReadyChan: make(chan string, 1),
	})
	if err != nil {
		return fmt.Errorf("oauth authorization failed: %w", err)
	}

	fmt.Println("✅ OAuth authorization successful!")
	fmt.Printf("Access Token: %s\n", token.AccessToken)
	if token.RefreshToken != "" {
		fmt.Printf("Refresh Token: %s\n", token.RefreshToken)
	}
	if !token.Expiry.IsZero() {
		fmt.Printf("Token Expires: %s\n", token.Expiry.Format(time.RFC3339))
	}

	// If there's a userinfo endpoint, fetch user info
	if !service.UserinfoURL.Null && service.UserinfoURL.Value != "" {
		fmt.Println("\nFetching user information...")
		if err := c.fetchUserInfo(token, service.UserinfoURL.Value); err != nil {
			fmt.Printf("Warning: Failed to fetch user info: %v\n", err)
		}
	}

	return nil
}

func (c *OAuthServiceAuthorizeCommand) fetchUserInfo(token *oauth2.Token, userinfoURL string) error {
	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequest("GET", userinfoURL, nil)
	if err != nil {
		return err
	}
	
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token.AccessToken))
	req.Header.Set("Accept", "application/json")
	
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("userinfo endpoint returned status %d", resp.StatusCode)
	}
	
	var userInfo map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		return err
	}
	
	fmt.Println("User Information:")
	for key, value := range userInfo {
		fmt.Printf("  %s: %v\n", key, value)
	}
	
	return nil
}

func displayOAuthServices(services *[]api.OAuthServiceResponse) {
	if services == nil || len(*services) == 0 {
		fmt.Println("No OAuth services found.")
		return
	}

	headers := []string{"ID", "Name", "Display Name", "Active", "Grant Types"}
	data := make([]map[string]interface{}, len(*services))

	for i, service := range *services {
		isActive := "No"
		if service.IsActive {
			isActive = "Yes"
		}

		grantTypes := "None"
		if len(service.SupportedGrantTypes) > 0 {
			grantTypes = fmt.Sprintf("%v", service.SupportedGrantTypes)
		}

		data[i] = map[string]interface{}{
			"ID":           service.ID,
			"Name":         service.Name,
			"Display Name": service.DisplayName,
			"Active":       isActive,
			"Grant Types":  grantTypes,
		}
	}

	util.DisplaySimpleTable(data, headers)
}