package util

import (
	"net/http"

	"github.com/arctir/devgraph-cli/pkg/auth"
	"github.com/arctir/devgraph-cli/pkg/config"
	devgraphv1 "github.com/arctir/go-devgraph/pkg/apis/devgraph/v1"
)

func GetAuthenticatedHTTPClient(config config.Config) (*http.Client, error) {
	creds, err := auth.LoadCredentials()
	if err != nil {
		return nil, err
	}

	return devgraphv1.NewAuthHTTPClient(
		config.ApiURL,
		"https://primary-ghoul-65.clerk.accounts.dev/oauth/token",
		config.ClientID,
		creds.IDToken,
		creds.RefreshToken,
		config.Environment,
	)
}

func GetAuthenticatedClient(config config.Config) (*devgraphv1.ClientWithResponses, error) {
	creds, err := auth.LoadCredentials()
	if err != nil {
		return nil, err
	}

	return devgraphv1.NewAuthClient(
		config.ApiURL,
		"https://primary-ghoul-65.clerk.accounts.dev/oauth/token",
		config.ClientID,
		creds.IDToken,
		creds.RefreshToken,
		config.Environment,
	)
}
