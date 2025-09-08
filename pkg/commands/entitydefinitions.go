package commands

import (
	"context"
	"fmt"

	"github.com/arctir/devgraph-cli/pkg/util"
	api "github.com/arctir/go-devgraph/pkg/apis/devgraph/v1"
	"github.com/google/uuid"
)

type EntityDefinitionCommand struct {
	Create EntityDefinitionCreateCommand `cmd:"create" help:"Create a new entity definition."`
	List   EntityDefinitionListCommand   `cmd:"" help:"List entity definitions."`
	Get    EntityDefinitionGetCommand    `cmd:"get" help:"Get an entity definition by ID."`
	Delete EntityDefinitionDeleteCommand `cmd:"delete" help:"Delete an entity definition by ID."`
}

type EntityDefinitionCreateCommand struct {
	EnvWrapperCommand
	FileName string `arg:"" required:"" help:"Path to the entity definition JSON file."`
}

type EntityDefinitionListCommand struct {
	EnvWrapperCommand
}

type EntityDefinitionGetCommand struct {
	EnvWrapperCommand
	Id string `arg:"" required:"" help:"ID of the entity definition to retrieve."`
}

type EntityDefinitionDeleteCommand struct {
	EnvWrapperCommand
	Id string `arg:"" required:"" help:"ID of the entity definition to delete."`
}

func (e *EntityDefinitionCreateCommand) Run() error {
	return nil
}

func (e *EntityDefinitionListCommand) Run() error {
	client, err := util.GetAuthenticatedClient(e.Config)
	if err != nil {
		return fmt.Errorf("failed to create authenticated client: %w", err)
	}

	resp, err := client.GetEntityDefinitions(context.Background())
	if err != nil {
		return fmt.Errorf("failed to list entity definitions: %w", err)
	}
	// Check if response is successful
	switch r := resp.(type) {
	case *api.GetEntityDefinitionsOKApplicationJSON:
		defs := []api.EntityDefinitionResponse(*r)
		data := make([]map[string]any, 0, len(defs))
		for i, def := range defs {
			gvk := fmt.Sprintf("%s/%s", def.Group, def.Kind)
			name := ""
			if def.Name.IsSet() {
				name = def.Name.Value
			}
			
			// Debug: print spec for first entity to see available fields
			if i == 0 {
				fmt.Printf("DEBUG - Spec fields: %+v\n", def.Spec)
			}
			
			data = append(data, map[string]any{"ID": def.ID.String(), "GVK": gvk, "Description": name})
		}
		util.DisplaySimpleTable(data, []string{"ID", "GVK", "Description"})
	default:
		return fmt.Errorf("failed to fetch entity definitions")
	}
	return nil
}

func (e *EntityDefinitionGetCommand) Run() error {
	return nil
}

func (e *EntityDefinitionDeleteCommand) Run() error {
	client, err := util.GetAuthenticatedClient(e.Config)
	if err != nil {
		return fmt.Errorf("failed to create authenticated client: %w", err)
	}

	uuid, err := uuid.Parse(e.Id)
	if err != nil {
		return fmt.Errorf("invalid UUID: %w", err)
	}
	params := api.DeleteEntityDefinitionParams{
		DefinitionID: uuid,
	}
	resp, err := client.DeleteEntityDefinition(context.Background(), params)
	if err != nil {
		return fmt.Errorf("failed to delete entity definition: %w", err)
	}
	// Check if response is successful
	switch resp.(type) {
	case *api.DeleteEntityDefinitionNoContent:
		// Success
	default:
		return fmt.Errorf("failed to delete entity definition")
	}

	fmt.Printf("Entity definition with ID '%s' deleted successfully.\n", e.Id)

	return nil
}
