package commands

import (
	"context"
	"fmt"

	"github.com/arctir/devgraph-cli/pkg/util"
	api "github.com/arctir/go-devgraph/pkg/apis/devgraph/v1"
	"github.com/google/uuid"
)

type ModelCommand struct {
	Create ModelCreateCommand `cmd:"create" help:"Create a new Model resource."`
	Get    ModelGetCommand    `cmd:"get" help:"Retrieve an Model resource by ID."`
	List   ModelListCommand   `cmd:"" help:"List Model resources."`
	Delete ModelDeleteCommand `cmd:"delete" help:"Delete an Model resource by ID."`
}

type ModelCreateCommand struct {
	EnvWrapperCommand
	ProviderID  string  `arg:"" required:"" help:"ID of the Model Provider to create this Model under."`
	Name        string  `arg:"" required:"" help:"Name of the Model resource to create."`
	Description *string `arg:"" optional:"" help:"Description of the Model resource."`
	Default     *bool   `arg:"" optional:"" help:"Set this Model as the default for the project."`
}

type ModelListCommand struct {
	EnvWrapperCommand
	Output string `short:"o" help:"Output format: table, json, yaml" default:"table"`
}

type ModelGetCommand struct {
	EnvWrapperCommand
	Id string `arg:"" required:"" help:"ID of the Model resource to retrieve."`
}

type ModelDeleteCommand struct {
	EnvWrapperCommand
	Id string `arg:"" required:"" help:"ID of the Model resource to delete."`
}

func (e *ModelCreateCommand) Run() error {
	client, err := util.GetAuthenticatedClient(e.Config)
	if err != nil {
		return fmt.Errorf("failed to create authenticated client: %w", err)
	}

	providerId, err := uuid.Parse(e.ProviderID)
	if err != nil {
		return fmt.Errorf("invalid provider ID format: %w", err)
	}

	body := api.ModelCreate{
		ProviderID: providerId,
		Name:       e.Name,
	}

	// Set optional fields if provided
	if e.Description != nil {
		body.Description = api.NewOptNilString(*e.Description)
	}
	if e.Default != nil {
		body.Default = api.NewOptBool(*e.Default)
	}

	// Make the API call to create the model
	response, err := client.CreateModel(context.TODO(), &body)
	if err != nil {
		return fmt.Errorf("failed to create model: %w", err)
	}

	// Check the response type
	switch r := response.(type) {
	case *api.ModelResponse:
		models := []api.ModelResponse{*r}
		displayModels(&models)
	default:
		return fmt.Errorf("failed to create model")
	}

	return nil
}

func (e *ModelGetCommand) Run() error {
	client, err := util.GetAuthenticatedClient(e.Config)
	if err != nil {
		return fmt.Errorf("failed to create authenticated client: %w", err)
	}
	params := api.GetModelParams{
		ModelName: e.Id,
	}
	resp, err := client.GetModel(context.Background(), params)
	if err != nil {
		return fmt.Errorf("failed to get model: %w", err)
	}
	// Check the response type
	switch r := resp.(type) {
	case *api.ModelResponse:
		models := []api.ModelResponse{*r}
		displayModels(&models)
	default:
		return fmt.Errorf("model with name '%s' not found", e.Id)
	}
	return nil
}

func (e *ModelListCommand) Run() error {
	client, err := util.GetAuthenticatedClient(e.Config)
	if err != nil {
		return fmt.Errorf("failed to create authenticated client: %w", err)
	}

	resp, err := client.GetModels(context.Background())
	if err != nil {
		return fmt.Errorf("failed to list models: %w", err)
	}

	// Check the response type
	switch r := resp.(type) {
	case *api.GetModelsOKApplicationJSON:
		models := []api.ModelResponse(*r)
		if len(models) == 0 {
			fmt.Println("No models found.")
			return nil
		}

		type modelOutput struct {
			ID         string `json:"id" yaml:"id"`
			Name       string `json:"name" yaml:"name"`
			ProviderID string `json:"provider_id" yaml:"provider_id"`
		}

		structured := make([]modelOutput, len(models))
		tableData := make([]map[string]any, len(models))
		for i, model := range models {
			structured[i] = modelOutput{
				ID:         model.ID.String(),
				Name:       model.Name,
				ProviderID: model.ProviderID.String(),
			}
			tableData[i] = map[string]any{
				"ID":          model.ID.String(),
				"Name":        model.Name,
				"Provider ID": model.ProviderID.String(),
			}
		}

		headers := []string{"ID", "Name", "Provider ID"}
		return util.FormatOutput(e.Output, structured, headers, tableData)
	default:
		return fmt.Errorf("failed to list models")
	}
}

func (e *ModelDeleteCommand) Run() error {
	client, err := util.GetAuthenticatedClient(e.Config)
	if err != nil {
		return fmt.Errorf("failed to create authenticated client: %w", err)
	}
	params := api.DeleteModelParams{
		ModelName: e.Id,
	}
	resp, err := client.DeleteModel(context.Background(), params)
	if err != nil {
		return fmt.Errorf("failed to delete model: %w", err)
	}
	// Check the response type
	switch resp.(type) {
	case *api.DeleteModelNoContent:
		// Success
	default:
		return fmt.Errorf("failed to delete model")
	}
	fmt.Printf("✅ Model '%s' deleted successfully.\n", e.Id)
	return nil
}

func displayModels(models *[]api.ModelResponse) {
	if models == nil || len(*models) == 0 {
		fmt.Println("No models found.")
		return
	}

	headers := []string{"ID", "Name", "Provider ID"}
	data := make([]map[string]interface{}, len(*models))
	for i, model := range *models {
		data[i] = map[string]interface{}{
			"Name":        model.Name,
			"ID":          model.ID,
			"Provider ID": model.ProviderID,
		}
	}
	util.DisplaySimpleTable(data, headers)
}
