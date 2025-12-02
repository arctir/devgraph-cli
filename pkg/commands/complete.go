// Package commands provides command-line command implementations for Devgraph CLI.
package commands

import (
	"context"
	"fmt"

	"github.com/arctir/devgraph-cli/pkg/config"
	"github.com/arctir/devgraph-cli/pkg/util"
	api "github.com/arctir/go-devgraph/pkg/apis/devgraph/v1"
)

// CompleteCommand is a hidden command that provides dynamic completions for shell scripts.
// It outputs resource names/IDs that can be used by shell completion functions.
type CompleteCommand struct {
	config.Config
	ResourceType string `arg:"" help:"Type of resource to complete (contexts, clusters, users, environments, tokens, providers, mcps, models, entities, entity-definitions)"`
}

// Run executes the completion lookup and prints results to stdout.
func (c *CompleteCommand) Run() error {
	c.Config.ApplyDefaults()

	switch c.ResourceType {
	// Local config resources (no API call needed)
	case "contexts":
		return c.completeContexts()
	case "clusters":
		return c.completeClusters()
	case "users":
		return c.completeUsers()

	// API resources (requires authentication)
	case "environments":
		return c.completeEnvironments()
	case "tokens":
		return c.completeTokens()
	case "providers":
		return c.completeProviders()
	case "mcps":
		return c.completeMCPs()
	case "models":
		return c.completeModels()
	case "entities":
		return c.completeEntities()
	case "entity-definitions":
		return c.completeEntityDefinitions()

	default:
		return fmt.Errorf("unknown resource type: %s", c.ResourceType)
	}
}

// Local config completions

func (c *CompleteCommand) completeContexts() error {
	userConfig, err := config.LoadUserConfig()
	if err != nil {
		return nil // Silently fail for completions
	}

	for name := range userConfig.Contexts {
		fmt.Println(name)
	}
	return nil
}

func (c *CompleteCommand) completeClusters() error {
	userConfig, err := config.LoadUserConfig()
	if err != nil {
		return nil
	}

	for name := range userConfig.Clusters {
		fmt.Println(name)
	}
	return nil
}

func (c *CompleteCommand) completeUsers() error {
	userConfig, err := config.LoadUserConfig()
	if err != nil {
		return nil
	}

	for name := range userConfig.Users {
		fmt.Println(name)
	}
	return nil
}

// API resource completions

func (c *CompleteCommand) completeEnvironments() error {
	if !util.IsAuthenticated() {
		return nil
	}

	envs, err := util.GetEnvironments(c.Config)
	if err != nil {
		return nil
	}

	for _, env := range *envs {
		// Output both name and slug for flexibility
		fmt.Println(env.Name)
		if env.Slug != env.Name {
			fmt.Println(env.Slug)
		}
	}
	return nil
}

func (c *CompleteCommand) completeTokens() error {
	client, err := c.getClient()
	if err != nil {
		return nil
	}

	resp, err := client.GetTokens(context.TODO())
	if err != nil {
		return nil
	}

	switch r := resp.(type) {
	case *api.GetTokensOKApplicationJSON:
		for _, token := range *r {
			fmt.Println(token.Name)
		}
	}
	return nil
}

func (c *CompleteCommand) completeProviders() error {
	client, err := c.getClient()
	if err != nil {
		return nil
	}

	resp, err := client.ListConfiguredProviders(context.TODO())
	if err != nil {
		return nil
	}

	switch r := resp.(type) {
	case *api.ConfiguredProvidersListResponse:
		for _, provider := range r.Providers {
			fmt.Println(provider.Name)
		}
	}
	return nil
}

func (c *CompleteCommand) completeMCPs() error {
	client, err := c.getClient()
	if err != nil {
		return nil
	}

	resp, err := client.GetMcpendpoints(context.TODO())
	if err != nil {
		return nil
	}

	switch r := resp.(type) {
	case *api.GetMcpendpointsOKApplicationJSON:
		for _, mcp := range *r {
			fmt.Println(mcp.Name)
		}
	}
	return nil
}

func (c *CompleteCommand) completeModels() error {
	client, err := c.getClient()
	if err != nil {
		return nil
	}

	resp, err := client.GetModels(context.TODO())
	if err != nil {
		return nil
	}

	switch r := resp.(type) {
	case *api.GetModelsOKApplicationJSON:
		for _, model := range *r {
			fmt.Println(model.Name)
		}
	}
	return nil
}

func (c *CompleteCommand) completeEntities() error {
	client, err := c.getClient()
	if err != nil {
		return nil
	}

	params := api.GetEntitiesParams{}
	resp, err := client.GetEntities(context.TODO(), params)
	if err != nil {
		return nil
	}

	switch r := resp.(type) {
	case *api.EntityResultSetResponse:
		for _, entity := range r.PrimaryEntities {
			fmt.Println(entity.Name)
		}
	}
	return nil
}

func (c *CompleteCommand) completeEntityDefinitions() error {
	client, err := c.getClient()
	if err != nil {
		return nil
	}

	resp, err := client.GetEntityDefinitions(context.TODO())
	if err != nil {
		return nil
	}

	switch r := resp.(type) {
	case *api.GetEntityDefinitionsOKApplicationJSON:
		for _, def := range *r {
			fmt.Println(def.Name)
		}
	}
	return nil
}

// getClient returns an authenticated API client
func (c *CompleteCommand) getClient() (*api.Client, error) {
	if !util.IsAuthenticated() {
		return nil, fmt.Errorf("not authenticated")
	}

	return util.GetAuthenticatedClient(c.Config)
}
