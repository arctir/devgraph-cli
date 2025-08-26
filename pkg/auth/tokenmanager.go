package auth

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"
)

// DevgraphTransport is a custom HTTP transport that adds Devgraph-specific headers
// to all outgoing requests. It wraps the standard HTTP transport and injects
// required headers like Content-Type, Accept, Authorization, and Devgraph-Environment.
type DevgraphTransport struct {
	// Transport is the underlying HTTP transport to use (defaults to http.DefaultTransport if nil)
	Transport http.RoundTripper
	// Headers contains additional headers to add to every request
	Headers   map[string]string
}

// RoundTrip implements the http.RoundTripper interface.
// It adds the configured headers to the request before forwarding to the underlying transport.
func (t *DevgraphTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// Clone the request to avoid modifying the original
	newReq := req.Clone(req.Context())

	// Add custom headers
	for key, value := range t.Headers {
		newReq.Header.Set(key, value)
	}

	// Use the underlying transport or default if none specified
	transport := t.Transport
	if transport == nil {
		transport = http.DefaultTransport
	}

	return transport.RoundTrip(newReq)
}

// OIDCTokenManager manages OIDC tokens and provides automatic token refresh capabilities.
// It implements oauth2.TokenSource and handles the complexity of refreshing expired tokens,
// saving updated tokens to persistent storage, and providing authenticated HTTP clients.
type OIDCTokenManager struct {
	config              oauth2.Config        // OAuth2 configuration
	mu                  sync.Mutex           // Mutex for thread-safe token operations
	token               *oauth2.Token        // Current token (access, refresh, ID token)
	tokenSrc            oauth2.TokenSource   // Underlying token source for refresh
	oidcVerifier        *oidc.IDTokenVerifier // Optional: for ID token verification
	devgraphEnvironment string               // Devgraph environment ID for API calls
}

// NewOIDCTokenManager creates a new token manager with the provided initial token.
// It sets up automatic token refresh and optionally configures ID token verification
// if an OIDC provider is provided.
func NewOIDCTokenManager(config oauth2.Config, initialToken *oauth2.Token, provider *oidc.Provider, devgraphEnvironment string) *OIDCTokenManager {
	ctx := context.Background()
	mgr := &OIDCTokenManager{
		config:              config,
		token:               initialToken,
		tokenSrc:            oauth2.ReuseTokenSource(initialToken, config.TokenSource(ctx, initialToken)),
		devgraphEnvironment: devgraphEnvironment,
	}
	if provider != nil {
		mgr.oidcVerifier = provider.Verifier(&oidc.Config{ClientID: config.ClientID})
	}
	return mgr
}

// Token implements oauth2.TokenSource, refreshing the token as needed
func (m *OIDCTokenManager) Token() (*oauth2.Token, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Get the latest token from the underlying TokenSource
	newToken, err := m.tokenSrc.Token()
	if err != nil {
		return nil, fmt.Errorf("failed to refresh token: %v", err)
	}

	// Update stored token if it has changed
	if newToken.AccessToken != m.token.AccessToken {
		creds, err := LoadCredentials()
		if err != nil {
			fmt.Printf("Failed to load existing credentials: %v\n", err)
			panic(err)
		}

		// Update all token information
		creds.AccessToken = newToken.AccessToken
		creds.RefreshToken = newToken.RefreshToken

		// Extract and save ID token if present
		if rawIDToken, ok := newToken.Extra("id_token").(string); ok && rawIDToken != "" {
			creds.IDToken = rawIDToken

			// Optionally verify the ID token
			if m.oidcVerifier != nil {
				if _, err := m.oidcVerifier.Verify(context.Background(), rawIDToken); err != nil {
					fmt.Printf("ID token verification failed: %v\n", err)
				}
			}
		}

		err = SaveCredentials(*creds)
		if err != nil {
			fmt.Printf("Failed to save refreshed token: %v\n", err)
			panic(err)
		}
		m.token = newToken
	}
	return m.token, nil
}

// GetCurrentToken returns the current token (thread-safe)
func (m *OIDCTokenManager) GetCurrentToken() *oauth2.Token {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.token
}

// HTTPClient returns an HTTP client that auto-refreshes the token
func (m *OIDCTokenManager) HTTPClient() *http.Client {
	transport := &DevgraphTransport{
		Transport: http.DefaultTransport,
		Headers: map[string]string{
			"Content-Type":         "application/json",
			"Accept":               "application/json",
			"Devgraph-Environment": m.devgraphEnvironment,
		},
	}

	return &http.Client{
		Transport: &oauth2.Transport{
			Source: m,
			Base:   transport,
		},
		Timeout: 30 * time.Second,
	}
}
