package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"

	"github.com/arctir/devgraph-cli/pkg/config"
	oidc "github.com/coreos/go-oidc/v3/oidc"
	"github.com/int128/oauth2cli"
	"github.com/int128/oauth2cli/oauth2params"
	"github.com/pkg/browser"
	"golang.org/x/oauth2"
	"golang.org/x/sync/errgroup"
)

// WellKnownConfig represents the standard OpenID Connect discovery document
type WellKnownConfig struct {
	Issuer                string `json:"issuer"`
	AuthorizationEndpoint string `json:"authorization_endpoint"`
	TokenEndpoint         string `json:"token_endpoint"`
	UserinfoEndpoint      string `json:"userinfo_endpoint"`
	JwksURI               string `json:"jwks_uri"`
	// Add more fields as needed (e.g., "end_session_endpoint")
}

func getWellKnownEndpoints(issuerURL string) (*WellKnownConfig, error) {
	// Parse the issuer URL
	u, err := url.Parse(issuerURL)
	if err != nil {
		return nil, fmt.Errorf("invalid issuer URL: %v", err)
	}

	// Append the well-known path
	u.Path = u.Path + "/.well-known/openid-configuration"

	// Make the HTTP request
	resp, err := http.Get(u.String())
	if err != nil {
		return nil, fmt.Errorf("failed to fetch well-known config: %v", err)
	}
	defer resp.Body.Close() //nolint:errcheck

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	// Decode the JSON response
	var config WellKnownConfig
	if err := json.NewDecoder(resp.Body).Decode(&config); err != nil {
		return nil, fmt.Errorf("failed to decode response: %v", err)
	}

	return &config, nil
}

func Authenticate(a config.Config) (*oauth2.Token, error) {
	ctx := context.Background()
	ready := make(chan string, 1)
	tokenChan := make(chan *oauth2.Token, 1)
	defer close(ready)
	defer close(tokenChan)

	providerConfig, err := getWellKnownEndpoints(a.IssuerURL)
	if err != nil {
		fmt.Println("Error:", err)
		return nil, err
	}

	// Configure OAuth2 with PKCE
	oauth2Config := oauth2.Config{
		ClientID:    a.ClientID,
		RedirectURL: a.RedirectURL,
		Endpoint: oauth2.Endpoint{
			AuthURL:  providerConfig.AuthorizationEndpoint,
			TokenURL: providerConfig.TokenEndpoint,
		},
		Scopes: []string{oidc.ScopeOpenID, "profile", "email"},
	}

	pkce, err := oauth2params.NewPKCE()
	if err != nil {
		log.Fatalf("error: %s", err)
	}

	// Set up oauth2cli configuration
	cliConfig := oauth2cli.Config{
		OAuth2Config:           oauth2Config,
		LocalServerBindAddress: []string{"0.0.0.0:8080"},

		AuthCodeOptions:      pkce.AuthCodeOptions(),
		TokenRequestOptions:  pkce.TokenRequestOptions(),
		LocalServerReadyChan: ready,
		Logf:                 log.Printf,
	}

	eg, ctx := errgroup.WithContext(ctx)
	eg.Go(func() error {
		select {
		case url := <-ready:
			log.Printf("Open %s", url)
			if err := browser.OpenURL(url); err != nil {
				log.Printf("could not open the browser: %s", err)
			}
			return nil
		case <-ctx.Done():
			return fmt.Errorf("context done while waiting for authorization: %w", ctx.Err())
		}
	})
	eg.Go(func() error {
		token, err := oauth2cli.GetToken(ctx, cliConfig)
		if err != nil {
			return fmt.Errorf("could not get a token: %w", err)
		}
		tokenChan <- token
		return nil
	})
	if err := eg.Wait(); err != nil {
		log.Fatalf("authorization error: %s", err)
	}

	token := <-tokenChan
	return token, nil
}
