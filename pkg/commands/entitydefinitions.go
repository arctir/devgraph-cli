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
	Output string `short:"o" help:"Output format: table, json, yaml" default:"table"`
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
		if len(defs) == 0 {
			fmt.Println("No entity definitions found.")
			return nil
		}

		type defOutput struct {
			ID          string `json:"id" yaml:"id"`
			Type        string `json:"type" yaml:"type"`
			Description string `json:"description,omitempty" yaml:"description,omitempty"`
		}

		structured := make([]defOutput, len(defs))
		tableData := make([]map[string]any, len(defs))
		for i, def := range defs {
			version := ""
			if def.Name.IsSet() {
				version = def.Name.Value
			}

			// Format Type as group/version/kind
			typeStr := fmt.Sprintf("%s/%s/%s", def.Group, version, def.Kind)

			description := ""
			if def.Description.IsSet() {
				description = def.Description.Value
			}

			structured[i] = defOutput{
				ID:          def.ID.String(),
				Type:        typeStr,
				Description: description,
			}
			tableData[i] = map[string]any{
				"ID":          def.ID.String(),
				"Type":        typeStr,
				"Description": description,
			}
		}

		headers := []string{"ID", "Type", "Description"}
		return util.FormatOutput(e.Output, structured, headers, tableData)
	default:
		return fmt.Errorf("failed to fetch entity definitions")
	}
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

	fmt.Printf("✅ Entity definition '%s' deleted successfully.\n", e.Id)

	return nil
}
