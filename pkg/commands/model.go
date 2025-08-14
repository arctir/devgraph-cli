package commands

import (
	"context"
	"fmt"

	"github.com/arctir/devgraph-cli/pkg/config"
	"github.com/arctir/devgraph-cli/pkg/util"
	devgraphv1 "github.com/arctir/go-devgraph/pkg/apis/devgraph/v1"
	"github.com/google/uuid"
)

type ModelCommand struct {
	Create ModelCreateCommand `cmd:"create" help:"Create a new Model resource."`
	Get    ModelGetCommand    `cmd:"get" help:"Retrieve an Model resource by ID."`
	List   ModelListCommand   `cmd:"" help:"List Model resources."`
	Delete ModelDeleteCommand `cmd:"delete" help:"Delete an Model resource by ID."`
}

type ModelCreateCommand struct {
	config.Config
	ProviderID  string  `arg:"" required:"" help:"ID of the Model Provider to create this Model under."`
	Name        string  `arg:"" required:"" help:"Name of the Model resource to create."`
	Description *string `arg:"" optional:"" help:"Description of the Model resource."`
	Default     *bool   `arg:"" optional:"" help:"Set this Model as the default for the project."`
}

type ModelListCommand struct {
	config.Config
}

type ModelGetCommand struct {
	config.Config
	Id string `arg:"" required:"" help:"ID of the Model resource to retrieve."`
}

type ModelDeleteCommand struct {
	config.Config
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

	body := devgraphv1.ModelCreate{
		ProviderId:  providerId,
		Name:        e.Name,
		Description: e.Description,
		Default:     e.Default,
	}

	// Make the API call to create the model
	response, err := client.CreateModelWithResponse(context.TODO(), body)
	if err != nil {
		return fmt.Errorf("failed to create model: %w", err)
	}

	// Check the response status
	if response.StatusCode() != 201 {
		if response.JSON422 != nil {
			return fmt.Errorf("validation error: %v", response.JSON422.Detail)
		}
		return fmt.Errorf("unexpected status code: %d", response.StatusCode())
	}

	model := response.JSON201
	models := []devgraphv1.ModelResponse{*model}
	displayModels(&models)

	return nil
}

func (e *ModelGetCommand) Run() error {
	client, err := util.GetAuthenticatedClient(e.Config)
	if err != nil {
		return fmt.Errorf("failed to create authenticated client: %w", err)
	}
	resp, err := client.GetModelWithResponse(context.Background(), e.Id)
	if err != nil {
		return fmt.Errorf("failed to get Model endpoint: %w", err)
	}
	if resp.StatusCode() != 200 {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode())
	}
	model := resp.JSON200
	if model == nil {
		return fmt.Errorf("model with ID '%s' not found", e.Id)
	}
	return nil
}

func (e *ModelListCommand) Run() error {
	client, err := util.GetAuthenticatedClient(e.Config)
	if err != nil {
		return fmt.Errorf("failed to create authenticated client: %w", err)
	}

	resp, err := client.GetModelsWithResponse(context.Background())
	if err != nil {
		return fmt.Errorf("failed to list models: %w", err)
	}

	if resp.StatusCode() != 200 {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode())
	}

	models := resp.JSON200
	if len(*models) == 0 {
		fmt.Println("No models found.")
		return nil
	}

	displayModels(models)

	return nil
}

func (e *ModelDeleteCommand) Run() error {
	client, err := util.GetAuthenticatedClient(e.Config)
	if err != nil {
		return fmt.Errorf("failed to create authenticated client: %w", err)
	}
	resp, err := client.DeleteModelWithResponse(context.Background(), e.Id)
	if err != nil {
		return fmt.Errorf("failed to delete model: %w", err)
	}
	if resp.StatusCode() != 204 {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode())
	}
	fmt.Printf("Model with ID '%s' deleted successfully.\n", e.Id)
	return nil
}

func displayModels(models *[]devgraphv1.ModelResponse) {
	if models == nil || len(*models) == 0 {
		fmt.Println("No models found.")
	}

	headers := []string{"ID", "Name", "Provider ID"}
	data := make([]map[string]interface{}, len(*models))
	for i, model := range *models {
		data[i] = map[string]interface{}{
			"Name":        model.Name,
			"ID":          model.Id,
			"Provider ID": model.ProviderId,
		}
	}
	util.DisplayTable(data, headers)
}
