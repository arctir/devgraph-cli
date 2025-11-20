package commands

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/arctir/devgraph-cli/pkg/util"
	api "github.com/arctir/go-devgraph/pkg/apis/devgraph/v1"
	"github.com/google/uuid"
	"gopkg.in/yaml.v3"
)

// ProviderCommand handles discovery provider management
type ProviderCommand struct {
	List   ProviderListCommand   `cmd:"" help:"List all configured discovery providers."`
	Get    ProviderGetCommand    `cmd:"get" help:"Get a specific configured discovery provider."`
	Delete ProviderDeleteCommand `cmd:"delete" help:"Delete a configured discovery provider."`
}

// ProviderListCommand lists all configured discovery providers
type ProviderListCommand struct {
	EnvWrapperCommand
	Output string `flag:"output,o" default:"table" help:"Output format: table, json, yaml."`
}

// ProviderGetCommand gets a specific configured discovery provider
type ProviderGetCommand struct {
	EnvWrapperCommand
	ProviderID string `arg:"" required:"" help:"Provider ID (UUID)."`
	Output     string `flag:"output,o" default:"json" help:"Output format: json, yaml."`
}

// ProviderDeleteCommand deletes a configured discovery provider
type ProviderDeleteCommand struct {
	EnvWrapperCommand
	ProviderID string `arg:"" required:"" help:"Provider ID (UUID)."`
	Yes        bool   `flag:"yes,y" help:"Skip confirmation prompt."`
}

func (p *ProviderListCommand) Run() error {
	client, err := util.GetAuthenticatedClient(p.Config)
	if err != nil {
		return fmt.Errorf("failed to create authenticated client: %w", err)
	}

	resp, err := client.ListConfiguredProviders(context.Background())
	if err != nil {
		return fmt.Errorf("failed to list providers: %w", err)
	}

	switch r := resp.(type) {
	case *api.ConfiguredProvidersListResponse:
		if len(r.Providers) == 0 {
			fmt.Println("No configured providers found.")
			return nil
		}

		switch p.Output {
		case "json":
			return p.displayAsJSON(r.Providers)
		case "yaml", "yml":
			return p.displayAsYAML(r.Providers)
		case "table":
			return p.displayAsTable(r.Providers)
		default:
			return fmt.Errorf("unsupported output format: %s", p.Output)
		}
	case *api.ListConfiguredProvidersNotFound:
		fmt.Println("No configured providers found.")
		return nil
	default:
		return fmt.Errorf("unexpected response type: %T", resp)
	}
}

func (p *ProviderListCommand) displayAsTable(providers []api.ConfiguredProviderResponse) error {
	headers := []string{"ID", "Provider Type", "Name", "Enabled", "Environment ID"}
	data := make([]map[string]interface{}, len(providers))

	for i, provider := range providers {
		enabled := "false"
		if provider.Enabled {
			enabled = "true"
		}

		data[i] = map[string]interface{}{
			"ID":             provider.ID.String(),
			"Provider Type":  provider.ProviderType,
			"Name":           provider.Name,
			"Enabled":        enabled,
			"Environment ID": provider.EnvironmentID.String(),
		}
	}

	displayEntityTable(data, headers)
	return nil
}

func (p *ProviderListCommand) displayAsJSON(providers []api.ConfiguredProviderResponse) error {
	jsonData, err := json.MarshalIndent(providers, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal providers to JSON: %w", err)
	}

	fmt.Println(string(jsonData))
	return nil
}

func (p *ProviderListCommand) displayAsYAML(providers []api.ConfiguredProviderResponse) error {
	yamlData, err := yaml.Marshal(providers)
	if err != nil {
		return fmt.Errorf("failed to marshal providers to YAML: %w", err)
	}

	fmt.Print(string(yamlData))
	return nil
}

func (p *ProviderGetCommand) Run() error {
	client, err := util.GetAuthenticatedClient(p.Config)
	if err != nil {
		return fmt.Errorf("failed to create authenticated client: %w", err)
	}

	providerID, err := uuid.Parse(p.ProviderID)
	if err != nil {
		return fmt.Errorf("invalid provider ID: %w", err)
	}

	params := api.GetConfiguredProviderParams{
		ProviderID: providerID,
	}

	resp, err := client.GetConfiguredProvider(context.Background(), params)
	if err != nil {
		return fmt.Errorf("failed to get provider: %w", err)
	}

	switch r := resp.(type) {
	case *api.ConfiguredProviderResponse:
		switch p.Output {
		case "json":
			jsonData, err := json.MarshalIndent(r, "", "  ")
			if err != nil {
				return fmt.Errorf("failed to marshal provider to JSON: %w", err)
			}
			fmt.Println(string(jsonData))
		case "yaml", "yml":
			yamlData, err := yaml.Marshal(r)
			if err != nil {
				return fmt.Errorf("failed to marshal provider to YAML: %w", err)
			}
			fmt.Print(string(yamlData))
		default:
			return fmt.Errorf("unsupported output format: %s", p.Output)
		}
		return nil
	case *api.GetConfiguredProviderNotFound:
		return fmt.Errorf("provider not found: %s", p.ProviderID)
	case *api.HTTPValidationError:
		return fmt.Errorf("validation error: %v", r.Detail)
	default:
		return fmt.Errorf("unexpected response type: %T", resp)
	}
}

func (p *ProviderDeleteCommand) Run() error {
	client, err := util.GetAuthenticatedClient(p.Config)
	if err != nil {
		return fmt.Errorf("failed to create authenticated client: %w", err)
	}

	providerID, err := uuid.Parse(p.ProviderID)
	if err != nil {
		return fmt.Errorf("invalid provider ID: %w", err)
	}

	// Confirm deletion unless --yes flag is provided
	if !p.Yes {
		fmt.Printf("Are you sure you want to delete provider %s? [y/N]: ", p.ProviderID)
		var response string
		fmt.Scanln(&response)
		if response != "y" && response != "Y" {
			fmt.Println("Deletion cancelled.")
			return nil
		}
	}

	params := api.DeleteConfiguredProviderParams{
		ProviderID: providerID,
	}

	resp, err := client.DeleteConfiguredProvider(context.Background(), params)
	if err != nil {
		return fmt.Errorf("failed to delete provider: %w", err)
	}

	switch resp.(type) {
	case *api.DeleteConfiguredProviderNoContent:
		fmt.Printf("✅ Provider '%s' deleted successfully.\n", p.ProviderID)
		return nil
	case *api.DeleteConfiguredProviderNotFound:
		return fmt.Errorf("provider not found: %s", p.ProviderID)
	case *api.HTTPValidationError:
		return fmt.Errorf("validation error")
	default:
		return fmt.Errorf("unexpected response type: %T", resp)
	}
}
