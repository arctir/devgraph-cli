// Package auth provides authentication functionality for Devgraph CLI.
// It handles OIDC (OpenID Connect) authentication flow, token management,
// and provides authenticated HTTP clients for API communication.
package auth

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/arctir/devgraph-cli/pkg/config"
	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"
)

// DefaultRedirectURL is the OAuth callback URL used during authentication.
// The oauth2cli library dynamically selects from ports 40000-40005.
const DefaultRedirectURL = "http://localhost:40000/callback"

// AuthenticatedClient creates an HTTP client configured with authentication
// for making requests to Devgraph API. The client automatically handles
// token refresh and includes required headers.
func AuthenticatedClient(c config.Config) (*http.Client, error) {
	// Get the default environment UUID from user settings
	environment := ""
	userConfig, err := config.LoadUserConfig()
	if err == nil {
		environment = userConfig.Settings.DefaultEnvironment
	}
	// Note: For some operations like listing environments, environment may be empty
	// We'll pass empty string if not set

	creds, err := LoadCredentials()
	if err != nil {
		return nil, fmt.Errorf("failed to load credentials: %w", err)
	}

	endpoints, err := getWellKnownEndpoints(c.IssuerURL)
	if err != nil {
		return nil, fmt.Errorf("failed to get well-known endpoints: %w", err)
	}
	oauth2Config := oauth2.Config{
		ClientID:    c.ClientID,
		RedirectURL: DefaultRedirectURL,
		Endpoint: oauth2.Endpoint{
			AuthURL:  endpoints.AuthorizationEndpoint,
			TokenURL: endpoints.TokenEndpoint,
		},
		Scopes: []string{"openid", "profile", "email", "public_metadata", "org:read"},
	}

	var exp float64
	if creds.Claims != nil {
		exp, _ = (*creds.Claims)["exp"].(float64)
	}

	expTime := time.Unix(int64(exp), 0)
	if exp > 0 {
		expTime = expTime.Add(-30 * time.Second)
	}

	token := &oauth2.Token{
		AccessToken:  creds.IDToken,
		RefreshToken: creds.RefreshToken,
		TokenType:    "Bearer",
		Expiry:       expTime,
	}

	provider, err := oidc.NewProvider(context.Background(), c.IssuerURL)
	if err != nil {
		return nil, fmt.Errorf("failed to create OIDC provider: %w", err)
	}

	tokenManager := NewOIDCTokenManager(
		oauth2Config,
		token,
		provider,
		environment,
	)
	httpClient := tokenManager.HTTPClient()
	return httpClient, nil
}
