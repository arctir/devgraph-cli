package commands

import (
	"context"
	"fmt"
	"strings"

	"github.com/arctir/devgraph-cli/pkg/util"
	api "github.com/arctir/go-devgraph/pkg/apis/devgraph/v1"
)

type TokenCommand struct {
	Create TokenCreate `cmd:"create" help:"Create a new opaque token."`
	List   TokenList   `cmd:"list" help:"List all opaque tokens."`
}

type TokenCreate struct {
	EnvWrapperCommand
	Name   string   `arg:"" name:"name" help:"Name of the opaque token to create"`
	Scopes []string `arg:"" name:"scopes" help:"Scopes for the opaque token"`
}

type TokenList struct {
	EnvWrapperCommand
}

var allowedScopes = []string{
	"create:entitydefinitions",
	"list:entitydefinitions",
	"delete:entitydefinitions",
	"create:entities",
	"read:entities",
	"delete:entities",
	"create:entityrelations",
	"delete:entityrelations",
}

func checkScopeInput(list []string) bool {
	if len(list) == 1 && list[0] == "all" {
		return true
	}

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
	if len(a.Scopes) == 1 && a.Scopes[0] == "all" {
		a.Scopes = allowedScopes
	} else {
		ok := checkScopeInput(a.Scopes)
		if !ok {
			return fmt.Errorf("one or more scopes are invalid. Allowed scopes are: %v", allowedScopes)
		}
	}
	client, err := util.GetAuthenticatedClient(a.Config)
	if err != nil {
		return fmt.Errorf("failed to create authenticated client: %w", err)
	}

	tokenCreate := api.ApiTokenCreate{
		Name:   a.Name,
		Scopes: a.Scopes,
		// Set ExpiresAt to null (no expiration)
		ExpiresAt: api.NewOptApiTokenCreateExpiresAt(api.NewNullApiTokenCreateExpiresAt(struct{}{})),
	}
	response, err := client.CreateToken(context.Background(), &tokenCreate)
	if err != nil {
		return fmt.Errorf("error creating token: %v", err)
	}
	// Check the response type
	switch r := response.(type) {
	case *api.ApiTokenResponse:
		tokens := []api.ApiTokenResponse{*r}
		displayTokens(&tokens)
	default:
		return fmt.Errorf("failed to create token")
	}
	return nil
}

func (a *TokenList) Run() error {
	client, err := util.GetAuthenticatedClient(a.Config)
	if err != nil {
		return fmt.Errorf("failed to create authenticated client: %w", err)
	}

	response, err := client.GetTokens(context.Background())
	if err != nil {
		return fmt.Errorf("error listing tokens: %v", err)
	}
	// Check the response type
	switch r := response.(type) {
	case *api.GetTokensOKApplicationJSON:
		tokens := []api.ApiTokenResponse(*r)
		displayTokens(&tokens)
	default:
		return fmt.Errorf("failed to list tokens")
	}

	return nil
}

func displayTokens(tokens *[]api.ApiTokenResponse) {
	headers := []string{"ID", "Name", "Scopes", "Token", "Expires At"}

	data := make([]map[string]interface{}, 0, len(*tokens))
	for _, token := range *tokens {
		expiresAt := "Never"
		if token.ExpiresAt.IsSet() {
			if expires, ok := token.ExpiresAt.Get(); ok {
				if expires.IsString() {
					if s, ok := expires.GetString(); ok && s != "" {
						expiresAt = s
					}
				}
			}
		}
		
		scopes := "None"
		if token.Scopes.IsSet() {
			if scopesData, ok := token.Scopes.Get(); ok {
				if scopesData.IsStringArray() {
					if scopesArray, ok := scopesData.GetStringArray(); ok {
						scopes = strings.Join(scopesArray, ", ")
					}
				}
			}
		}
		
		data = append(data, map[string]interface{}{
			"ID":         token.ID,
			"Name":       token.Name,
			"Scopes":     scopes,
			"Token":      token.Token,
			"Expires At": expiresAt,
		})
	}

	util.DisplaySimpleTable(data, headers)

}
