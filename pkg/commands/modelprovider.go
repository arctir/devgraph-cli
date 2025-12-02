package commands

import (
	"context"
	"fmt"

	"github.com/arctir/devgraph-cli/pkg/util"
	api "github.com/arctir/go-devgraph/pkg/apis/devgraph/v1"
	"github.com/google/uuid"
)

type ModelProviderCommand struct {
	Create ModelProviderCreateCommand `cmd:"create" help:"Create a new ModelProvider resource."`
	Get    ModelProviderGetCommand    `cmd:"get" help:"Retrieve an ModelProvider resource by ID."`
	List   ModelProviderListCommand   `cmd:"" help:"List ModelProvider resources."`
	Delete ModelProviderDeleteCommand `cmd:"delete" help:"Delete an ModelProvider resource by ID."`
}

type ModelProviderCreateCommand struct {
	EnvWrapperCommand
	Type    string `arg:"" enum:"openai,xai,anthropic" required:"" help:"Type of the ModelProvider resource to create (e.g., 'openai')."`
	Name    string `arg:"" required:"" help:"Name of the ModelProvider resource to create."`
	ApiKey  string `arg:"" required:"" help:"API key for the ModelProvider resource."`
	Default *bool  `arg:"" optional:"" help:"Set this ModelProvider as the default for the project."`
}

type ModelProviderListCommand struct {
	EnvWrapperCommand
	Output string `short:"o" help:"Output format: table, json, yaml" default:"table"`
}

type ModelProviderGetCommand struct {
	EnvWrapperCommand
	Id string `arg:"" required:"" help:"ID of the ModelProvider resource to retrieve."`
}

type ModelProviderDeleteCommand struct {
	EnvWrapperCommand
	Id string `arg:"" required:"" help:"ID of the ModelProvider resource to delete."`
}

func (e *ModelProviderCreateCommand) Run() error {
	client, err := util.GetAuthenticatedClient(e.Config)
	if err != nil {
		return fmt.Errorf("failed to create authenticated client: %w", err)
	}

	var data api.ModelProviderCreateData
	switch e.Type {
	case "openai":
		provider := api.OpenAIModelProviderCreate{
			Type:   "openai",
			Name:   e.Name,
			APIKey: e.ApiKey,
		}
		// Set optional fields if provided
		if e.Default != nil {
			provider.Default = api.NewOptBool(*e.Default)
		}
		data = api.NewOpenAIModelProviderCreateModelProviderCreateData(provider)
	case "xai":
		provider := api.XAIModelProviderCreate{
			Type:   "xai",
			Name:   e.Name,
			APIKey: e.ApiKey,
		}
		// Set optional fields if provided
		if e.Default != nil {
			provider.Default = api.NewOptBool(*e.Default)
		}
		data = api.NewXAIModelProviderCreateModelProviderCreateData(provider)
	case "anthropic":
		provider := api.AnthropicModelProviderCreate{
			Type:   "anthropic",
			Name:   e.Name,
			APIKey: e.ApiKey,
		}
		// Set optional fields if provided
		if e.Default != nil {
			provider.Default = api.NewOptBool(*e.Default)
		}
		data = api.NewAnthropicModelProviderCreateModelProviderCreateData(provider)
	default:
		return fmt.Errorf("unsupported model provider type: %s", e.Type)
	}

	body := api.ModelProviderCreate{
		Data: data,
	}

	// Make the API call to create the model provider
	response, err := client.CreateModelprovider(context.TODO(), &body)
	if err != nil {
		return fmt.Errorf("failed to create model provider: %w", err)
	}

	// Check the response type
	switch response.(type) {
	case *api.ModelProviderResponse:
		fmt.Printf("✅ Model provider '%s' created successfully.\n", e.Name)
	default:
		return fmt.Errorf("failed to create model provider")
	}

	return nil
}

func (e *ModelProviderGetCommand) Run() error {
	client, err := util.GetAuthenticatedClient(e.Config)
	if err != nil {
		return fmt.Errorf("failed to create authenticated client: %w", err)
	}
	uuid, err := uuid.Parse(e.Id)
	if err != nil {
		return fmt.Errorf("invalid UUID: %w", err)
	}
	params := api.GetModelproviderParams{
		ProviderID: uuid,
	}
	resp, err := client.GetModelprovider(context.Background(), params)
	if err != nil {
		return fmt.Errorf("failed to get model provider: %w", err)
	}
	// Check the response type
	switch r := resp.(type) {
	case *api.ModelProviderResponse:
		fmt.Printf("Model provider found: %v\n", *r)
	default:
		return fmt.Errorf("model provider with ID '%s' not found", e.Id)
	}

	return nil
}

func (e *ModelProviderListCommand) Run() error {
	client, err := util.GetAuthenticatedClient(e.Config)
	if err != nil {
		return fmt.Errorf("failed to create authenticated client: %w", err)
	}

	resp, err := client.GetModelproviders(context.Background())
	if err != nil {
		return fmt.Errorf("failed to list model providers: %w", err)
	}

	// Check the response type
	switch r := resp.(type) {
	case *api.GetModelprovidersOKApplicationJSON:
		providers := []api.ModelProviderResponse(*r)
		if len(providers) == 0 {
			fmt.Println("No model providers found.")
			return nil
		}

		type providerOutput struct {
			ID   string `json:"id" yaml:"id"`
			Name string `json:"name" yaml:"name"`
			Type string `json:"type" yaml:"type"`
		}

		structured := make([]providerOutput, len(providers))
		tableData := make([]map[string]any, len(providers))
		for i, provider := range providers {
			var name, id, providerType string
			if provider.IsXAIModelProviderResponse() {
				if p, ok := provider.GetXAIModelProviderResponse(); ok {
					name, id, providerType = p.Name, p.ID.String(), "xai"
				}
			} else if provider.IsOpenAIModelProviderResponse() {
				if p, ok := provider.GetOpenAIModelProviderResponse(); ok {
					name, id, providerType = p.Name, p.ID.String(), "openai"
				}
			} else if provider.IsAnthropicModelProviderResponse() {
				if p, ok := provider.GetAnthropicModelProviderResponse(); ok {
					name, id, providerType = p.Name, p.ID.String(), "anthropic"
				}
			} else {
				name, id, providerType = "Unknown", "Unknown", "unknown"
			}

			structured[i] = providerOutput{ID: id, Name: name, Type: providerType}
			tableData[i] = map[string]any{"ID": id, "Name": name, "Type": providerType}
		}

		headers := []string{"ID", "Name", "Type"}
		return util.FormatOutput(e.Output, structured, headers, tableData)
	default:
		return fmt.Errorf("failed to list model providers")
	}
}

func (e *ModelProviderDeleteCommand) Run() error {
	client, err := util.GetAuthenticatedClient(e.Config)
	if err != nil {
		return fmt.Errorf("failed to create authenticated client: %w", err)
	}
	uuid, err := uuid.Parse(e.Id)
	if err != nil {
		return fmt.Errorf("invalid UUID: %w", err)
	}
	params := api.DeleteModelproviderParams{
		ProviderID: uuid,
	}
	resp, err := client.DeleteModelprovider(context.Background(), params)
	if err != nil {
		return fmt.Errorf("failed to delete model provider: %w", err)
	}
	// Check the response type
	switch resp.(type) {
	case *api.DeleteModelproviderNoContent:
		// Success
	default:
		return fmt.Errorf("failed to delete model provider")
	}
	fmt.Printf("✅ Model provider '%s' deleted successfully.\n", e.Id)
	return nil
}
