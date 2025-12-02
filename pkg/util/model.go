package util

import (
	"context"
	"fmt"

	"github.com/arctir/devgraph-cli/pkg/config"
	api "github.com/arctir/go-devgraph/pkg/apis/devgraph/v1"
)

// GetModels retrieves all available models
func GetModels(config config.Config) (*[]api.ModelResponse, error) {
	client, err := GetAuthenticatedClient(config)
	if err != nil {
		return nil, err
	}

	ctx := context.Background()
	resp, err := client.GetModels(ctx)
	if err != nil {
		return nil, err
	}

	// Check if response is successful
	switch r := resp.(type) {
	case *api.GetModelsOKApplicationJSON:
		models := []api.ModelResponse(*r)
		return &models, nil
	default:
		return nil, fmt.Errorf("failed to fetch models")
	}
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
		if model.Name == modelName || model.ID.String() == modelName {
			return nil
		}
	}

	return fmt.Errorf("model '%s' not found. Available models: %v", modelName, getModelList(*models))
}

// getModelList returns a list of model names for error messages
func getModelList(models []api.ModelResponse) []string {
	var names []string
	for _, model := range models {
		names = append(names, model.Name)
	}
	return names
}
