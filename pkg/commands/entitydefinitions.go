package commands

import (
	"context"
	"fmt"
	"net/http"

	"github.com/arctir/devgraph-cli/pkg/util"
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

	resp, err := client.GetEntityDefinitionsWithResponse(context.Background())
	if err != nil {
		return fmt.Errorf("failed to list entity definitions: %w", err)
	}
	if resp.StatusCode() != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode())
	}

	data := make([]map[string]interface{}, 0, len(*resp.JSON200))
	for _, def := range *resp.JSON200 {
		gvk := fmt.Sprintf("%s/%s", def.Group, def.Kind)
		data = append(data, map[string]interface{}{"ID": def.Id.String(), "GVK": gvk})
	}
	util.DisplaySimpleTable(data, []string{"ID", "GVK"})
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

	resp, err := client.DeleteEntityDefinitionWithResponse(context.Background(), e.Id)
	if err != nil {
		return fmt.Errorf("failed to delete entity definition: %w", err)
	}
	if resp.StatusCode() != 204 {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode())
	}

	fmt.Printf("Entity definition with ID '%s' deleted successfully.\n", e.Id)

	return nil
}
