package commands

import (
	"context"
	"fmt"

	"github.com/arctir/devgraph-cli/pkg/util"
	devgraphv1 "github.com/arctir/go-devgraph/pkg/apis/devgraph/v1"
)

type MCPCommand struct {
	Create MCPCreateCommand `cmd:"create" help:"Create a new MCP resource."`
	Get    MCPGetCommand    `cmd:"get" help:"Retrieve an MCP resource by ID."`
	List   MCPListCommand   `cmd:"" help:"List MCP resources."`
	Delete MCPDeleteCommand `cmd:"delete" help:"Delete an MCP resource by ID."`
}

type MCPCreateCommand struct {
	EnvWrapperCommand
	Name        string `arg:"" required:"" help:"Name of the MCP resource to create."`
	Url         string `arg:"" required:"" help:"URL of the MCP resource to create."`
	Description string `arg:"" optional:"" help:"Description of the MCP resource."`
}

type MCPListCommand struct {
	EnvWrapperCommand
}

type MCPGetCommand struct {
	EnvWrapperCommand
	Id string `arg:"" required:"" help:"ID of the MCP resource to retrieve."`
}

type MCPDeleteCommand struct {
	EnvWrapperCommand
	Id string `arg:"" required:"" help:"ID of the MCP resource to delete."`
}

func (e *MCPCreateCommand) Run() error {
	client, err := util.GetAuthenticatedClient(e.Config)
	if err != nil {
		return fmt.Errorf("failed to create authenticated client: %w", err)
	}

	request := devgraphv1.CreateMcpendpointJSONRequestBody{
		Name:        e.Name,
		Url:         e.Url,
		Description: &e.Description,
	}

	resp, err := client.CreateMcpendpointWithResponse(context.Background(), request)
	if err != nil {
		return fmt.Errorf("failed to create MCP endpoint: %w", err)
	}
	if resp.StatusCode() != 201 {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode())
	}

	fmt.Printf("MCP endpoint '%s' created successfully.\n", e.Name)

	return nil
}

func (e *MCPGetCommand) Run() error {
	client, err := util.GetAuthenticatedClient(e.Config)
	if err != nil {
		return fmt.Errorf("failed to create authenticated client: %w", err)
	}
	resp, err := client.GetMcpendpointWithResponse(context.Background(), e.Id)
	if err != nil {
		return fmt.Errorf("failed to get MCP endpoint: %w", err)
	}
	if resp.StatusCode() != 200 {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode())
	}
	endpoint := resp.JSON200
	if endpoint == nil {
		return fmt.Errorf("MCP endpoint with ID '%s' not found", e.Id)
	}

	description := ""
	if endpoint.Description != nil {
		description = *endpoint.Description
	}

	fmt.Printf("ID: %s\nName: %s\nURL: %s\nDescription: %s\n", endpoint.Id, endpoint.Name, endpoint.Url, description)

	return nil
}

func (e *MCPListCommand) Run() error {
	client, err := util.GetAuthenticatedClient(e.Config)
	if err != nil {
		return fmt.Errorf("failed to create authenticated client: %w", err)
	}

	resp, err := client.GetMcpendpointsWithResponse(context.Background())
	if err != nil {
		return fmt.Errorf("failed to list MCP endpoints: %w", err)
	}

	if resp.StatusCode() != 200 {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode())
	}

	endpoints := resp.JSON200
	if len(*endpoints) == 0 {
		fmt.Println("No MCP endpoints found.")
		return nil
	}

	headers := []string{"ID", "Name", "Url"}
	data := make([]map[string]interface{}, len(*endpoints))
	for i, env := range *endpoints {
		data[i] = map[string]interface{}{
			"Name": env.Name,
			"ID":   env.Id,
			"Url":  env.Url,
		}
	}
	util.DisplayTable(data, headers)
	return nil
}

func (e *MCPDeleteCommand) Run() error {
	client, err := util.GetAuthenticatedClient(e.Config)
	if err != nil {
		return fmt.Errorf("failed to create authenticated client: %w", err)
	}
	resp, err := client.DeleteMcpendpointWithResponse(context.Background(), e.Id)
	if err != nil {
		return fmt.Errorf("failed to delete MCP endpoint: %w", err)
	}
	if resp.StatusCode() != 204 {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode())
	}
	fmt.Printf("MCP endpoint with ID '%s' deleted successfully.\n", e.Id)
	return nil
}
