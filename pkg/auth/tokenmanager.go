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

type DevgraphTransport struct {
	Transport http.RoundTripper
	Headers   map[string]string
}

// RoundTrip implements the http.RoundTripper interface
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

// OIDCTokenManager manages OIDC tokens and provides a refreshing TokenSource
type OIDCTokenManager struct {
	config              oauth2.Config
	mu                  sync.Mutex
	token               *oauth2.Token // Current token (access, refresh, ID token)
	tokenSrc            oauth2.TokenSource
	oidcVerifier        *oidc.IDTokenVerifier // Optional: for ID token verification
	devgraphEnvironment string
}

// NewOIDCTokenManager creates a new token manager with initial token
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
		creds.RefreshToken = newToken.RefreshToken
		err = SaveCredentials(*creds)
		if err != nil {
			fmt.Printf("Failed to save refreshed token: %v\n", err)
			panic(err)
		}
		m.token = newToken

		// Optionally verify the ID token if present
		if m.oidcVerifier != nil {
			rawIDToken, ok := newToken.Extra("id_token").(string)
			if ok && rawIDToken != "" {
				if _, err := m.oidcVerifier.Verify(context.Background(), rawIDToken); err != nil {
					fmt.Printf("ID token verification failed: %v\n", err)
				}
			}
		}
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
			"Authorization":        fmt.Sprintf("Bearer %s", m.GetCurrentToken().AccessToken),
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
