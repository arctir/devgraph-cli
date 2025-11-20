package util

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/arctir/devgraph-cli/pkg/auth"
	"github.com/arctir/devgraph-cli/pkg/config"
	api "github.com/arctir/go-devgraph/pkg/apis/devgraph/v1"
)

// debugTransport wraps an http.RoundTripper and logs requests/responses when debug is enabled
type debugTransport struct {
	transport http.RoundTripper
	debug     bool
}

func (t *debugTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if t.debug {
		fmt.Printf("\n--- HTTP Request ---\n")
		fmt.Printf("%s %s\n", req.Method, req.URL.String())
		fmt.Printf("Headers:\n")
		for k, v := range req.Header {
			// Don't log sensitive headers
			if k == "Authorization" {
				fmt.Printf("  %s: [REDACTED]\n", k)
			} else {
				fmt.Printf("  %s: %v\n", k, v)
			}
		}

		if req.Body != nil {
			bodyBytes, _ := io.ReadAll(req.Body)
			req.Body.Close()
			req.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
			if len(bodyBytes) > 0 {
				fmt.Printf("Body: %s\n", string(bodyBytes))
			}
		}
	}

	resp, err := t.transport.RoundTrip(req)

	if t.debug && resp != nil {
		fmt.Printf("\n--- HTTP Response ---\n")
		fmt.Printf("Status: %s\n", resp.Status)
		fmt.Printf("Headers:\n")
		for k, v := range resp.Header {
			fmt.Printf("  %s: %v\n", k, v)
		}

		if resp.Body != nil {
			bodyBytes, _ := io.ReadAll(resp.Body)
			resp.Body.Close()
			resp.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
			if len(bodyBytes) > 0 && len(bodyBytes) < 10000 { // Only log if not too large
				fmt.Printf("Body: %s\n", string(bodyBytes))
			} else if len(bodyBytes) > 0 {
				fmt.Printf("Body: [%d bytes]\n", len(bodyBytes))
			}
		}
		fmt.Printf("---\n\n")
	}

	return resp, err
}

// GetAuthenticatedHTTPClient returns an HTTP client configured with authentication
// for making requests to Devgraph API endpoints. The client automatically handles
// token refresh and includes necessary headers for API communication.
func GetAuthenticatedHTTPClient(cfg config.Config) (*http.Client, error) {
	// Use the token manager for automatic refresh
	client, err := auth.AuthenticatedClient(cfg)
	if err != nil {
		return nil, err
	}

	// Wrap transport with debug logging if enabled
	if cfg.Debug {
		if client.Transport == nil {
			client.Transport = http.DefaultTransport
		}
		client.Transport = &debugTransport{
			transport: client.Transport,
			debug:     true,
		}
	}

	return client, nil
}

// DevgraphSecuritySource implements the SecuritySource interface for Devgraph API
type DevgraphSecuritySource struct {
	config config.Config
}

// OAuth2PasswordBearer provides the OAuth2 bearer token for API requests
func (s *DevgraphSecuritySource) OAuth2PasswordBearer(ctx context.Context, operationName api.OperationName) (api.OAuth2PasswordBearer, error) {
	creds, err := auth.LoadCredentials()
	if err != nil {
		return api.OAuth2PasswordBearer{}, err
	}
	
	return api.OAuth2PasswordBearer{
		Token:  creds.IDToken,
		Scopes: []string{}, // Scopes are handled by the token itself
	}, nil
}

// GetAuthenticatedClient returns a Devgraph API client with authentication configured.
// This is a higher-level client that wraps the HTTP client and provides typed
// methods for interacting with Devgraph API endpoints.
// It supports both context-based and legacy configuration.
func GetAuthenticatedClient(cfg config.Config) (*api.Client, error) {
	// Load user config to check for contexts
	userConfig, err := config.LoadUserConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load user config: %w", err)
	}

	// If using contexts, override config with context settings
	if userConfig.CurrentContext != "" {
		context, cluster, user, err := userConfig.GetCurrentContext()
		if err == nil {
			// Override API URL from cluster
			if cluster.Server != "" {
				cfg.ApiURL = cluster.Server
			}
			if cluster.IssuerURL != "" {
				cfg.IssuerURL = cluster.IssuerURL
			}
			if cluster.ClientID != "" {
				cfg.ClientID = cluster.ClientID
			}

			// Note: user credentials are loaded separately via LoadCredentials
			// Environment UUID is loaded from userConfig.Settings.DefaultEnvironment
			_ = user
			_ = context
		}
	}

	httpClient, err := GetAuthenticatedHTTPClient(cfg)
	if err != nil {
		return nil, err
	}

	securitySource := &DevgraphSecuritySource{config: cfg}
	return api.NewClient(cfg.ApiURL, securitySource, api.WithClient(httpClient))
}

// IsAuthenticated checks if the user has valid authentication credentials.
// It returns true if valid, unexpired credentials are available, false otherwise.
func IsAuthenticated() bool {
	creds, err := auth.LoadCredentials()
	if err != nil || creds.IDToken == "" || creds.AccessToken == "" {
		return false
	}

	// Check if tokens are expired
	if creds.Claims != nil {
		if exp, ok := (*creds.Claims)["exp"].(float64); ok {
			if time.Now().Unix() > int64(exp) {
				return false
			}
		}
	}

	return true
}
