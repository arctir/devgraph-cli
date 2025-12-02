package commands

import (
	"context"
	"fmt"
	"strings"

	"github.com/arctir/devgraph-cli/pkg/util"
	api "github.com/arctir/go-devgraph/pkg/apis/devgraph/v1"
	"github.com/google/uuid"
)

type TokenCommand struct {
	Create TokenCreate `cmd:"create" help:"Create a new opaque token."`
	Delete TokenDelete `cmd:"delete" help:"Delete an opaque token."`
	Get    TokenGet    `cmd:"get" help:"Get an opaque token by ID."`
	List   TokenList   `cmd:"list" help:"List all opaque tokens."`
	Update TokenUpdate `cmd:"update" help:"Update an opaque token."`
}

type TokenCreate struct {
	EnvWrapperCommand
	Name   string   `arg:"" name:"name" help:"Name of the opaque token to create"`
	Scopes []string `arg:"" name:"scopes" help:"Scopes for the opaque token"`
}

type TokenGet struct {
	EnvWrapperCommand
	ID string `arg:"" name:"id" help:"ID of the opaque token to get"`
}

type TokenList struct {
	EnvWrapperCommand
	Output string `short:"o" help:"Output format: table, json, yaml" default:"table"`
}

type TokenUpdate struct {
	EnvWrapperCommand
	ID     string   `arg:"" name:"id" help:"ID of the opaque token to update"`
	Name   string   `flag:"name" help:"New name for the token"`
	Scopes []string `flag:"scopes" help:"New scopes for the token (comma-separated or 'all')"`
}

type TokenDelete struct {
	EnvWrapperCommand
	ID string `arg:"" name:"id" help:"ID of the opaque token to delete"`
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
	}
	// Set ExpiresAt to null (no expiration) by creating an explicitly null OptNilString
	tokenCreate.ExpiresAt.SetToNull()
	response, err := client.CreateToken(context.Background(), &tokenCreate)
	if err != nil {
		return fmt.Errorf("failed to create token: %w", err)
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
		return fmt.Errorf("failed to list tokens: %w", err)
	}
	// Check the response type
	switch r := response.(type) {
	case *api.GetTokensOKApplicationJSON:
		tokens := []api.ApiTokenResponse(*r)
		if len(tokens) == 0 {
			fmt.Println("No tokens found.")
			return nil
		}

		type tokenOutput struct {
			ID        string   `json:"id" yaml:"id"`
			Name      string   `json:"name" yaml:"name"`
			Scopes    []string `json:"scopes" yaml:"scopes"`
			Token     string   `json:"token" yaml:"token"`
			ExpiresAt string   `json:"expires_at,omitempty" yaml:"expires_at,omitempty"`
		}

		structured := make([]tokenOutput, len(tokens))
		tableData := make([]map[string]any, len(tokens))
		for i, token := range tokens {
			expiresAt := "Never"
			if expires, ok := token.ExpiresAt.Get(); ok && expires != "" {
				expiresAt = expires
			}

			scopes := []string{}
			scopesStr := "None"
			if scopesArray, ok := token.Scopes.Get(); ok && len(scopesArray) > 0 {
				scopes = scopesArray
				scopesStr = strings.Join(scopesArray, ", ")
			}

			structured[i] = tokenOutput{
				ID:        token.ID.String(),
				Name:      token.Name,
				Scopes:    scopes,
				Token:     token.Token,
				ExpiresAt: expiresAt,
			}
			tableData[i] = map[string]any{
				"ID":         token.ID.String(),
				"Name":       token.Name,
				"Scopes":     scopesStr,
				"Token":      token.Token,
				"Expires At": expiresAt,
			}
		}

		headers := []string{"ID", "Name", "Scopes", "Token", "Expires At"}
		return util.FormatOutput(a.Output, structured, headers, tableData)
	default:
		return fmt.Errorf("failed to list tokens")
	}
}

func (a *TokenGet) Run() error {
	client, err := util.GetAuthenticatedClient(a.Config)
	if err != nil {
		return fmt.Errorf("failed to create authenticated client: %w", err)
	}

	// Get all tokens and filter by ID
	response, err := client.GetTokens(context.Background())
	if err != nil {
		return fmt.Errorf("failed to get token: %w", err)
	}

	switch r := response.(type) {
	case *api.GetTokensOKApplicationJSON:
		tokens := []api.ApiTokenResponse(*r)
		// Find the token with matching ID
		for _, token := range tokens {
			if token.ID.String() == a.ID {
				displayTokens(&[]api.ApiTokenResponse{token})
				return nil
			}
		}
		return fmt.Errorf("token with ID %s not found", a.ID)
	default:
		return fmt.Errorf("failed to get token")
	}
}

func (a *TokenUpdate) Run() error {
	if a.Name == "" && len(a.Scopes) == 0 {
		return fmt.Errorf("must provide at least --name or --scopes to update")
	}

	// Validate scopes if provided
	if len(a.Scopes) > 0 {
		if len(a.Scopes) == 1 && a.Scopes[0] == "all" {
			a.Scopes = allowedScopes
		} else {
			ok := checkScopeInput(a.Scopes)
			if !ok {
				return fmt.Errorf("one or more scopes are invalid. Allowed scopes are: %v", allowedScopes)
			}
		}
	}

	client, err := util.GetAuthenticatedClient(a.Config)
	if err != nil {
		return fmt.Errorf("failed to create authenticated client: %w", err)
	}

	// Parse UUID
	tokenID, err := uuid.Parse(a.ID)
	if err != nil {
		return fmt.Errorf("invalid token ID: %w", err)
	}

	tokenUpdate := api.ApiTokenUpdate{}
	if a.Name != "" {
		tokenUpdate.Name.SetTo(a.Name)
	}
	if len(a.Scopes) > 0 {
		tokenUpdate.Scopes.SetTo(a.Scopes)
	}

	params := api.UpdateTokenParams{
		TokenID: tokenID,
	}

	response, err := client.UpdateToken(context.Background(), &tokenUpdate, params)
	if err != nil {
		return fmt.Errorf("failed to update token: %w", err)
	}

	switch r := response.(type) {
	case *api.ApiTokenResponse:
		fmt.Printf("Token %s updated successfully\n", r.ID)
		displayTokens(&[]api.ApiTokenResponse{*r})
	default:
		return fmt.Errorf("failed to update token")
	}

	return nil
}

func (a *TokenDelete) Run() error {
	client, err := util.GetAuthenticatedClient(a.Config)
	if err != nil {
		return fmt.Errorf("failed to create authenticated client: %w", err)
	}

	// Parse UUID
	tokenID, err := uuid.Parse(a.ID)
	if err != nil {
		return fmt.Errorf("invalid token ID: %w", err)
	}

	params := api.DeleteTokenParams{
		TokenID: tokenID,
	}

	response, err := client.DeleteToken(context.Background(), params)
	if err != nil {
		return fmt.Errorf("failed to delete token: %w", err)
	}

	switch response.(type) {
	case *api.DeleteTokenNoContent:
		fmt.Printf("✅ Token '%s' deleted successfully.\n", a.ID)
	case *api.DeleteTokenNotFound:
		return fmt.Errorf("token with ID %s not found", a.ID)
	default:
		return fmt.Errorf("failed to delete token")
	}

	return nil
}

func displayTokens(tokens *[]api.ApiTokenResponse) {
	headers := []string{"ID", "Name", "Scopes", "Token", "Expires At"}

	data := make([]map[string]interface{}, 0, len(*tokens))
	for _, token := range *tokens {
		expiresAt := "Never"
		if expires, ok := token.ExpiresAt.Get(); ok && expires != "" {
			expiresAt = expires
		}

		scopes := "None"
		if scopesArray, ok := token.Scopes.Get(); ok && len(scopesArray) > 0 {
			scopes = strings.Join(scopesArray, ", ")
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
