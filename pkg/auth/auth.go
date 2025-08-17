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

func AuthenticatedClient(c config.Config) (*http.Client, error) {
	environment := c.Environment
	// Note: For some operations like listing environments, environment may be empty
	// We'll pass empty string if not set

	creds, err := LoadCredentials()
	if err != nil {
		return nil, fmt.Errorf("failed to load credentials: %v", err)
	}

	endpoints, err := getWellKnownEndpoints(c.IssuerURL)
	if err != nil {
		return nil, fmt.Errorf("failed to get well-known endpoints: %v", err)
	}
	oauth2Config := oauth2.Config{
		ClientID:    c.ClientID,
		RedirectURL: c.RedirectURL,
		Endpoint: oauth2.Endpoint{
			AuthURL:  endpoints.AuthorizationEndpoint,
			TokenURL: endpoints.TokenEndpoint,
		},
		Scopes: []string{"openid", "profile", "email"},
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
		return nil, fmt.Errorf("failed to create OIDC provider: %v", err)
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
