package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/arctir/devgraph-cli/pkg/util"
	api "github.com/arctir/go-devgraph/pkg/apis/devgraph/v1"
)

type EntityCommand struct {
	Create EntityCreateCommand `cmd:"create" help:"Create a new entity."`
	Get    EntityGetCommand    `cmd:"get" help:"Get an entity by group/version/kind/namespace/name."`
	Delete EntityDeleteCommand `cmd:"delete" help:"Delete an entity by group/version/kind/namespace/name."`
}

type EntityCreateCommand struct {
	EnvWrapperCommand
	Group     string `arg:"" required:"" help:"Group of the entity (e.g., apps, core, extensions)."`
	Version   string `arg:"" required:"" help:"Version of the entity (e.g., v1, v1beta1)."`
	Namespace string `arg:"" required:"" help:"Namespace of the entity."`
	Plural    string `arg:"" required:"" help:"Plural form of the entity kind (e.g., deployments, services)."`
	FileName  string `arg:"" required:"" help:"Path to the entity JSON file."`
}

type EntityGetCommand struct {
	EnvWrapperCommand
	Group     string `arg:"" required:"" help:"Group of the entity (e.g., apps, core, extensions)."`
	Version   string `arg:"" required:"" help:"Version of the entity (e.g., v1, v1beta1)."`
	Kind      string `arg:"" required:"" help:"Kind of the entity (e.g., Deployment, Service)."`
	Namespace string `arg:"" required:"" help:"Namespace of the entity."`
	Name      string `arg:"" required:"" help:"Name of the entity."`
}

type EntityDeleteCommand struct {
	EnvWrapperCommand
	Group     string `arg:"" required:"" help:"Group of the entity (e.g., apps, core, extensions)."`
	Version   string `arg:"" required:"" help:"Version of the entity (e.g., v1, v1beta1)."`
	Kind      string `arg:"" required:"" help:"Kind of the entity (e.g., Deployment, Service)."`
	Namespace string `arg:"" required:"" help:"Namespace of the entity."`
	Name      string `arg:"" required:"" help:"Name of the entity."`
}

func (e *EntityCreateCommand) Run() error {
	client, err := util.GetAuthenticatedClient(e.Config)
	if err != nil {
		return fmt.Errorf("failed to create authenticated client: %w", err)
	}

	data, err := os.ReadFile(e.FileName)
	if err != nil {
		return fmt.Errorf("failed to read file %s: %w", e.FileName, err)
	}

	var entity api.Entity
	if err := json.Unmarshal(data, &entity); err != nil {
		return fmt.Errorf("failed to parse entity JSON: %w", err)
	}

	params := api.CreateEntityParams{
		Group:     e.Group,
		Version:   e.Version,
		Namespace: e.Namespace,
		Plural:    e.Plural,
	}
	resp, err := client.CreateEntity(context.Background(), &entity, params)
	if err != nil {
		return fmt.Errorf("failed to create entity: %w", err)
	}
	// Check if response is successful
	switch resp.(type) {
	case *api.EntityResponse:
		// Success
	default:
		return fmt.Errorf("failed to create entity")
	}

	fmt.Printf("Entity '%s' created successfully in namespace '%s'.\n", entity.Metadata.Name, e.Namespace)
	return nil
}

func (e *EntityGetCommand) Run() error {
	client, err := util.GetAuthenticatedClient(e.Config)
	if err != nil {
		return fmt.Errorf("failed to create authenticated client: %w", err)
	}

	params := api.GetEntityParams{
		Group:     e.Group,
		Version:   e.Version,
		Kind:      e.Kind,
		Namespace: e.Namespace,
		Name:      e.Name,
	}
	resp, err := client.GetEntity(context.Background(), params)
	if err != nil {
		return fmt.Errorf("failed to get entity: %w", err)
	}
	// Check if response is successful
	switch r := resp.(type) {
	case *api.EntityResponse:
		output, err := json.MarshalIndent(*r, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal entity to JSON: %w", err)
		}
		fmt.Println(string(output))
	default:
		return fmt.Errorf("entity not found")
	}
	return nil
}

func (e *EntityDeleteCommand) Run() error {
	client, err := util.GetAuthenticatedClient(e.Config)
	if err != nil {
		return fmt.Errorf("failed to create authenticated client: %w", err)
	}

	params := api.DeleteEntityParams{
		Group:     e.Group,
		Version:   e.Version,
		Kind:      e.Kind,
		Namespace: e.Namespace,
		Name:      e.Name,
	}
	resp, err := client.DeleteEntity(context.Background(), params)
	if err != nil {
		return fmt.Errorf("failed to delete entity: %w", err)
	}
	// Check if response is successful
	switch resp.(type) {
	case *api.DeleteEntityNoContent:
		// Success
	default:
		return fmt.Errorf("failed to delete entity")
	}

	fmt.Printf("Entity '%s' deleted successfully from namespace '%s'.\n", e.Name, e.Namespace)
	return nil
}
