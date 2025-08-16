package commands

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/arctir/devgraph-cli/pkg/config"
	"github.com/arctir/devgraph-cli/pkg/util"
	devgraphv1 "github.com/arctir/go-devgraph/pkg/apis/devgraph/v1"
	"k8s.io/utils/ptr"
)

type TokenCommand struct {
	Create TokenCreate `cmd:"create" help:"Create a new opaque token."`
	List   TokenList   `cmd:"list" help:"List all opaque tokens."`
}

type TokenCreate struct {
	config.Config
	Name   string   `arg:"" name:"name" help:"Name of the opaque token to create"`
	Scopes []string `arg:"" name:"scopes" help:"Scopes for the opaque token"`
}

type TokenList struct {
	config.Config
}

var allowedScopes = []string{
	"create:entities",
	"create:entitydefinitions",
	"create:entityrelations",
}

func checkScopeInput(list []string) bool {
	refMap := make(map[string]bool)
	for _, item := range allowedScopes {
		refMap[item] = true
	}

	// Check if each item in list exists in reference
	for _, item := range list {
		if !refMap[item] {
			return false
		}
	}
	return true
}

func (a *TokenCreate) Run() error {
	ok := checkScopeInput(a.Scopes)
	if !ok {
		return fmt.Errorf("one or more scopes are invalid. Allowed scopes are: %v", allowedScopes)
	}

	client, err := util.GetAuthenticatedClient(a.Config)
	if err != nil {
		return fmt.Errorf("failed to create authenticated client: %w", err)
	}

	response, err := client.CreateTokenWithResponse(context.Background(), devgraphv1.CreateTokenJSONRequestBody{
		Name:      a.Name,
		Scopes:    ptr.To(a.Scopes),
		ExpiresAt: ptr.To(""), // Empty string means no expiration
	})
	if err != nil {
		return fmt.Errorf("error creating token: %v", err)
	}
	if response.StatusCode() != http.StatusCreated {
		return fmt.Errorf("failed to create token, status code: %d", response.StatusCode())
	}

	token := response.JSON201
	tokens := []devgraphv1.ApiTokenResponse{*token}
	displayTokens(&tokens)
	return nil
}

func (a *TokenList) Run() error {
	client, err := util.GetAuthenticatedClient(a.Config)
	if err != nil {
		return fmt.Errorf("failed to create authenticated client: %w", err)
	}

	response, err := client.GetTokensWithResponse(context.Background())
	if err != nil {
		return fmt.Errorf("error listing tokens: %v", err)
	}
	if response.StatusCode() != http.StatusOK {
		return fmt.Errorf("failed to list tokens, status code: %d", response.StatusCode())
	}

	tokens := response.JSON200
	displayTokens(tokens)

	return nil
}

func displayTokens(tokens *[]devgraphv1.ApiTokenResponse) {
	headers := []string{"ID", "Name", "Scopes", "Token", "Expires At"}

	data := make([]map[string]interface{}, 0, len(*tokens))
	for _, token := range *tokens {
		expiresAt := "Never"
		if token.ExpiresAt != nil && *token.ExpiresAt != "" {
			expiresAt = *token.ExpiresAt
		}
		data = append(data, map[string]interface{}{
			"ID":         token.Id,
			"Name":       token.Name,
			"Scopes":     strings.Join(*token.Scopes, ", "),
			"Token":     token.Token,
			"Expires At": expiresAt,
		})
	}

	util.DisplayTable(data, headers)

}
