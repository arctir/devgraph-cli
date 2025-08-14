package commands

import (
	"context"
	"fmt"

	"github.com/arctir/devgraph-cli/pkg/config"
	"github.com/arctir/devgraph-cli/pkg/util"
	devgraphv1 "github.com/arctir/go-devgraph/pkg/apis/devgraph/v1"
)

type ModelProviderCommand struct {
	Create ModelProviderCreateCommand `cmd:"create" help:"Create a new ModelProvider resource."`
	Get    ModelProviderGetCommand    `cmd:"get" help:"Retrieve an ModelProvider resource by ID."`
	List   ModelProviderListCommand   `cmd:"" help:"List ModelProvider resources."`
	Delete ModelProviderDeleteCommand `cmd:"delete" help:"Delete an ModelProvider resource by ID."`
}

type ModelProviderCreateCommand struct {
	config.Config
	Type    string `arg:"" enum:"openai,xai" required:"" help:"Type of the ModelProvider resource to create (e.g., 'openai')."`
	Name    string `arg:"" required:"" help:"Name of the ModelProvider resource to create."`
	ApiKey  string `arg:"" required:"" help:"API key for the ModelProvider resource."`
	Default *bool  `arg:"" optional:"" help:"Set this ModelProvider as the default for the project."`
}

type ModelProviderListCommand struct {
	config.Config
}

type ModelProviderGetCommand struct {
	config.Config
	Id string `arg:"" required:"" help:"ID of the ModelProvider resource to retrieve."`
}

type ModelProviderDeleteCommand struct {
	config.Config
	Id string `arg:"" required:"" help:"ID of the ModelProvider resource to delete."`
}

func (e *ModelProviderCreateCommand) Run() error {
	client, err := util.GetAuthenticatedClient(e.Config)
	if err != nil {
		return fmt.Errorf("failed to create authenticated client: %w", err)
	}

	var data devgraphv1.ModelProviderCreate_Data
	switch e.Type {
	case "openai":
		provider := devgraphv1.OpenAIModelProviderCreate{
			Type:    "openai",
			Name:    e.Name,
			ApiKey:  e.ApiKey,
			Default: e.Default,
		}
		err := data.FromOpenAIModelProviderCreate(provider)
		if err != nil {
			return fmt.Errorf("failed to create OpenAI model provider data: %w", err)
		}
	case "xai":
		provider := devgraphv1.XAIModelProviderCreate{
			Type:    "xai",
			Name:    e.Name,
			ApiKey:  e.ApiKey,
			Default: e.Default,
		}
		err := data.FromXAIModelProviderCreate(provider)
		if err != nil {
			return fmt.Errorf("failed to create XAI model provider data: %w", err)
		}
	default:
		return fmt.Errorf("unsupported model provider type: %s", e.Type)
	}

	body := devgraphv1.ModelProviderCreate{
		Data: data,
	}

	// Make the API call to create the model provider
	response, err := client.CreateModelproviderWithResponse(context.TODO(), body)
	if err != nil {
		return fmt.Errorf("failed to create model provider: %w", err)
	}

	// Check the response status
	if response.StatusCode() != 201 {
		if response.JSON422 != nil {
			return fmt.Errorf("validation error: %v", response.JSON422.Detail)
		}
		return fmt.Errorf("unexpected status code: %d", response.StatusCode())
	}

	// Return the created model provider
	return nil
}

func (e *ModelProviderGetCommand) Run() error {
	client, err := util.GetAuthenticatedClient(e.Config)
	if err != nil {
		return fmt.Errorf("failed to create authenticated client: %w", err)
	}
	resp, err := client.GetModelproviderWithResponse(context.Background(), e.Id)
	if err != nil {
		return fmt.Errorf("failed to get ModelProvider endpoint: %w", err)
	}
	if resp.StatusCode() != 200 {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode())
	}
	provider := resp.JSON200
	if provider == nil {
		return fmt.Errorf("ModelProvider endpoint with ID '%s' not found", e.Id)
	}

	return nil
}

func (e *ModelProviderListCommand) Run() error {
	client, err := util.GetAuthenticatedClient(e.Config)
	if err != nil {
		return fmt.Errorf("failed to create authenticated client: %w", err)
	}

	resp, err := client.GetModelprovidersWithResponse(context.Background())
	if err != nil {
		return fmt.Errorf("failed to list model providers: %w", err)
	}

	if resp.StatusCode() != 200 {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode())
	}

	providers := resp.JSON200
	if len(*providers) == 0 {
		fmt.Println("No model providers found.")
		return nil
	}

	displayModelProviders(providers)
	return nil
}

func (e *ModelProviderDeleteCommand) Run() error {
	client, err := util.GetAuthenticatedClient(e.Config)
	if err != nil {
		return fmt.Errorf("failed to create authenticated client: %w", err)
	}
	resp, err := client.DeleteModelproviderWithResponse(context.Background(), e.Id)
	if err != nil {
		return fmt.Errorf("failed to delete model provider: %w", err)
	}
	if resp.StatusCode() != 204 {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode())
	}
	fmt.Printf("Model provider with ID '%s' deleted successfully.\n", e.Id)
	return nil
}

func displayModelProviders(providers *[]devgraphv1.ModelProviderResponse) {
	if providers == nil || len(*providers) == 0 {
		fmt.Println("No environments found.")
	}

	headers := []string{"ID", "Name", "Type"}
	data := make([]map[string]interface{}, len(*providers))
	for i, provider := range *providers {
		providerType, err := provider.Discriminator()
		if err != nil {
			fmt.Printf("Failed to determine provider type: %v\n", err)
			continue
		}
		if providerType == "xai" {
			p, err := provider.AsXAIModelProviderResponse()
			if err == nil {
				data[i] = map[string]interface{}{
					"Name": p.Name,
					"ID":   p.Id,
					"Type": "xai",
				}
				continue
			}
		} else if providerType == "openai" {
			p, err := provider.AsOpenAIModelProviderResponse()
			if err == nil {
				data[i] = map[string]interface{}{
					"Name": p.Name,
					"ID":   p.Id,
					"Type": "openai",
				}
				continue
			}
		}
	}
	util.DisplayTable(data, headers)
}
