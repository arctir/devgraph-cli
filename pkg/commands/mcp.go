package commands

import (
	"context"
	"fmt"
	"strings"

	"github.com/arctir/devgraph-cli/pkg/util"
	api "github.com/arctir/go-devgraph/pkg/apis/devgraph/v1"
	"github.com/google/uuid"
)

type MCPCommand struct {
	Create MCPCreateCommand `cmd:"create" help:"Create a new MCP resource."`
	Get    MCPGetCommand    `cmd:"get" help:"Retrieve an MCP resource by ID."`
	List   MCPListCommand   `cmd:"" help:"List MCP resources."`
	Delete MCPDeleteCommand `cmd:"delete" help:"Delete an MCP resource by ID."`
}

type MCPCreateCommand struct {
	EnvWrapperCommand
	Name        string   `arg:"" required:"" help:"Name of the MCP resource to create."`
	Url         string   `arg:"" required:"" help:"URL of the MCP resource to create."`
	Description string   `arg:"" optional:"" help:"Description of the MCP resource."`
	Headers     []string `flag:"header,H" help:"Headers as key:value pairs (can be specified multiple times)."`
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

	// Parse headers from key:value format
	headers := make(map[string]string)
	for _, header := range e.Headers {
		parts := strings.SplitN(header, ":", 2)
		if len(parts) != 2 {
			return fmt.Errorf("invalid header format '%s', expected 'key:value'", header)
		}
		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		if key == "" {
			return fmt.Errorf("header key cannot be empty in '%s'", header)
		}
		headers[key] = value
	}

	request := api.MCPEndpointCreate{
		Name: e.Name,
		URL:  e.Url,
	}
	
	// Set optional fields if provided
	if e.Description != "" {
		request.Description = api.NewOptMCPEndpointCreateDescription(api.NewStringMCPEndpointCreateDescription(e.Description))
	}
	if len(headers) > 0 {
		request.Headers = api.NewOptMCPEndpointCreateHeaders(api.MCPEndpointCreateHeaders(headers))
	}

	resp, err := client.CreateMcpendpoint(context.Background(), &request)
	if err != nil {
		return fmt.Errorf("failed to create MCP endpoint: %w", err)
	}
	// Check the response type
	switch resp.(type) {
	case *api.MCPEndpointResponse:
		// Success
	default:
		return fmt.Errorf("failed to create MCP endpoint")
	}

	fmt.Printf("MCP endpoint '%s' created successfully.\n", e.Name)

	return nil
}

func (e *MCPGetCommand) Run() error {
	client, err := util.GetAuthenticatedClient(e.Config)
	if err != nil {
		return fmt.Errorf("failed to create authenticated client: %w", err)
	}
	uuid, err := uuid.Parse(e.Id)
	if err != nil {
		return fmt.Errorf("invalid UUID: %w", err)
	}
	params := api.GetMcpendpointParams{
		McpendpointID: uuid,
	}
	resp, err := client.GetMcpendpoint(context.Background(), params)
	if err != nil {
		return fmt.Errorf("failed to get MCP endpoint: %w", err)
	}
	// Check the response type
	switch r := resp.(type) {
	case *api.MCPEndpointResponse:
		description := ""
		if r.Description.IsSet() {
			if desc, ok := r.Description.Get(); ok && desc.IsString() {
				if s, ok := desc.GetString(); ok {
					description = s
				}
			}
		}
		fmt.Printf("ID: %s\nName: %s\nURL: %s\nDescription: %s\n", r.ID, r.Name, r.URL, description)
	default:
		return fmt.Errorf("MCP endpoint with ID '%s' not found", e.Id)
	}

	return nil
}

func (e *MCPListCommand) Run() error {
	client, err := util.GetAuthenticatedClient(e.Config)
	if err != nil {
		return fmt.Errorf("failed to create authenticated client: %w", err)
	}

	resp, err := client.GetMcpendpoints(context.Background())
	if err != nil {
		return fmt.Errorf("failed to list MCP endpoints: %w", err)
	}

	// Check the response type
	switch r := resp.(type) {
	case *api.GetMcpendpointsOKApplicationJSON:
		endpoints := []api.MCPEndpointResponse(*r)
		if len(endpoints) == 0 {
			fmt.Println("No MCP endpoints found.")
			return nil
		}

		headers := []string{"ID", "Name", "URL"}
		data := make([]map[string]interface{}, len(endpoints))
		for i, endpoint := range endpoints {
			data[i] = map[string]interface{}{
				"Name": endpoint.Name,
				"ID":   endpoint.ID,
				"URL":  endpoint.URL,
			}
		}
		util.DisplaySimpleTable(data, headers)
	default:
		return fmt.Errorf("failed to list MCP endpoints")
	}
	return nil
}

func (e *MCPDeleteCommand) Run() error {
	client, err := util.GetAuthenticatedClient(e.Config)
	if err != nil {
		return fmt.Errorf("failed to create authenticated client: %w", err)
	}
	uuid, err := uuid.Parse(e.Id)
	if err != nil {
		return fmt.Errorf("invalid UUID: %w", err)
	}
	params := api.DeleteMcpendpointParams{
		McpendpointID: uuid,
	}
	resp, err := client.DeleteMcpendpoint(context.Background(), params)
	if err != nil {
		return fmt.Errorf("failed to delete MCP endpoint: %w", err)
	}
	// Check the response type
	switch resp.(type) {
	case *api.DeleteMcpendpointNoContent:
		// Success
	default:
		return fmt.Errorf("failed to delete MCP endpoint")
	}
	fmt.Printf("MCP endpoint with ID '%s' deleted successfully.\n", e.Id)
	return nil
}
