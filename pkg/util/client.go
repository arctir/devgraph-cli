package util

import (
	"net/http"

	"github.com/arctir/devgraph-cli/pkg/auth"
	"github.com/arctir/devgraph-cli/pkg/config"
	devgraphv1 "github.com/arctir/go-devgraph/pkg/apis/devgraph/v1"
)

// GetAuthenticatedHTTPClient returns an HTTP client configured with authentication
// for making requests to Devgraph API endpoints. The client automatically handles
// token refresh and includes necessary headers for API communication.
func GetAuthenticatedHTTPClient(config config.Config) (*http.Client, error) {
	// Use the token manager for automatic refresh
	return auth.AuthenticatedClient(config)
}

// GetAuthenticatedClient returns a Devgraph API client with authentication configured.
// This is a higher-level client that wraps the HTTP client and provides typed
// methods for interacting with Devgraph API endpoints.
func GetAuthenticatedClient(config config.Config) (*devgraphv1.ClientWithResponses, error) {
	httpClient, err := GetAuthenticatedHTTPClient(config)
	if err != nil {
		return nil, err
	}

	return devgraphv1.NewClientWithResponses(config.ApiURL, devgraphv1.WithHTTPClient(httpClient))
}
