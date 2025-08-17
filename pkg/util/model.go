package util

import (
	"context"
	"fmt"

	"github.com/arctir/devgraph-cli/pkg/config"
	devgraphv1 "github.com/arctir/go-devgraph/pkg/apis/devgraph/v1"
)

// GetModels retrieves all available models
func GetModels(config config.Config) (*[]devgraphv1.ModelResponse, error) {
	client, err := GetAuthenticatedClient(config)
	if err != nil {
		return nil, err
	}

	ctx := context.Background()
	resp, err := client.GetModelsWithResponse(ctx)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode() != 200 {
		return nil, fmt.Errorf("failed to fetch models: status code %d", resp.StatusCode())
	}

	return resp.JSON200, nil
}

// ValidateModel checks if the given model name exists and is accessible
func ValidateModel(config config.Config, modelName string) error {
	models, err := GetModels(config)
	if err != nil {
		return fmt.Errorf("failed to get models: %w", err)
	}

	if models == nil || len(*models) == 0 {
		return fmt.Errorf("no models available")
	}

	for _, model := range *models {
		// Check by name, ID, or any other identifier
		if model.Name == modelName || model.Id.String() == modelName {
			return nil
		}
	}

	return fmt.Errorf("model '%s' not found. Available models: %v", modelName, getModelList(*models))
}

// getModelList returns a list of model names for error messages
func getModelList(models []devgraphv1.ModelResponse) []string {
	var names []string
	for _, model := range models {
		names = append(names, model.Name)
	}
	return names
}
