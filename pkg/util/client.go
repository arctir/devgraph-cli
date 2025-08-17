package util

import (
	"net/http"

	"github.com/arctir/devgraph-cli/pkg/auth"
	"github.com/arctir/devgraph-cli/pkg/config"
	devgraphv1 "github.com/arctir/go-devgraph/pkg/apis/devgraph/v1"
)

func GetAuthenticatedHTTPClient(config config.Config) (*http.Client, error) {
	// Use the token manager for automatic refresh
	return auth.AuthenticatedClient(config)
}

func GetAuthenticatedClient(config config.Config) (*devgraphv1.ClientWithResponses, error) {
	httpClient, err := GetAuthenticatedHTTPClient(config)
	if err != nil {
		return nil, err
	}

	return devgraphv1.NewClientWithResponses(config.ApiURL, devgraphv1.WithHTTPClient(httpClient))
}
