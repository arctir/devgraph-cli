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
	Update MCPUpdateCommand `cmd:"update" help:"Update an existing MCP resource by ID."`
	Delete MCPDeleteCommand `cmd:"delete" help:"Delete an MCP resource by ID."`
}

type MCPCreateCommand struct {
	EnvWrapperCommand
	Name                string   `arg:"" required:"" help:"Name of the MCP resource to create."`
	Url                 string   `arg:"" required:"" help:"URL of the MCP resource to create."`
	Description         string   `arg:"" optional:"" help:"Description of the MCP resource."`
	Headers             []string `flag:"header,H" optional:"" help:"Headers as key:value pairs (can be specified multiple times)."`
	DevgraphAuth        *bool    `flag:"devgraph-auth" optional:"" help:"Enable Devgraph authentication for this endpoint."`
	SupportsResources   *bool    `flag:"supports-resources" optional:"" help:"Indicates if this endpoint supports MCP resources."`
	OAuthServiceID      *string  `flag:"oauth-service-id" optional:"" help:"Link to an OAuth service by ID."`
}

type MCPListCommand struct {
	EnvWrapperCommand
}

type MCPGetCommand struct {
	EnvWrapperCommand
	Id string `arg:"" required:"" help:"ID of the MCP resource to retrieve."`
}

type MCPUpdateCommand struct {
	EnvWrapperCommand
	Id                  string   `arg:"" required:"" help:"ID of the MCP resource to update."`
	Name                *string  `flag:"name" help:"Update the name of the MCP resource."`
	Url                 *string  `flag:"url" help:"Update the URL of the MCP resource."`
	Description         *string  `flag:"description" help:"Update the description of the MCP resource."`
	Headers             []string `flag:"header,H" help:"Update headers as key:value pairs (replaces all existing headers)."`
	DevgraphAuth        *bool    `flag:"devgraph-auth" help:"Update Devgraph authentication setting."`
	SupportsResources   *bool    `flag:"supports-resources" help:"Update supports resources setting."`
	OAuthServiceID      *string  `flag:"oauth-service-id" help:"Link to an OAuth service by ID (when API supports it)."`
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
	if e.DevgraphAuth != nil {
		request.DevgraphAuth = api.NewOptBool(*e.DevgraphAuth)
	}
	if e.SupportsResources != nil {
		request.SupportsResources = api.NewOptBool(*e.SupportsResources)
	}
	if e.OAuthServiceID != nil {
		oauthUUID, err := uuid.Parse(*e.OAuthServiceID)
		if err != nil {
			return fmt.Errorf("invalid OAuth service ID: %w", err)
		}
		request.OAuthServiceID = api.NewOptMCPEndpointCreateOAuthServiceID(
			api.NewUUIDMCPEndpointCreateOAuthServiceID(oauthUUID),
		)
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
		
		oauthServiceID := ""
		if r.OAuthServiceID.IsSet() {
			if oauth, ok := r.OAuthServiceID.Get(); ok {
				if oauth.IsUUID() {
					if id, ok := oauth.GetUUID(); ok {
						oauthServiceID = id.String()
					}
				} else if oauth.IsNull() {
					oauthServiceID = "(null)"
				}
			}
		} else {
			oauthServiceID = "(not set)"
		}
		
		fmt.Printf("ID: %s\nName: %s\nURL: %s\nDescription: %s\nOAuth Service ID: %s\n", 
			r.ID, r.Name, r.URL, description, oauthServiceID)
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

		headers := []string{"ID", "Name", "URL", "OAuth Service ID"}
		data := make([]map[string]interface{}, len(endpoints))
		for i, endpoint := range endpoints {
			oauthServiceID := ""
			if endpoint.OAuthServiceID.IsSet() {
				if oauth, ok := endpoint.OAuthServiceID.Get(); ok {
					if oauth.IsUUID() {
						if id, ok := oauth.GetUUID(); ok {
							oauthServiceID = id.String()
						}
					} else if oauth.IsNull() {
						oauthServiceID = "(null)"
					}
				}
			} else {
				oauthServiceID = "(not set)"
			}
			
			data[i] = map[string]interface{}{
				"Name":             endpoint.Name,
				"ID":               endpoint.ID,
				"URL":              endpoint.URL,
				"OAuth Service ID": oauthServiceID,
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

func (e *MCPUpdateCommand) Run() error {
	client, err := util.GetAuthenticatedClient(e.Config)
	if err != nil {
		return fmt.Errorf("failed to create authenticated client: %w", err)
	}

	mcpUUID, err := uuid.Parse(e.Id)
	if err != nil {
		return fmt.Errorf("invalid UUID: %w", err)
	}

	// Check if there's anything to update
	if e.Name == nil && e.Url == nil && e.Description == nil && 
		len(e.Headers) == 0 && e.DevgraphAuth == nil && 
		e.SupportsResources == nil && e.OAuthServiceID == nil {
		return fmt.Errorf("no fields specified to update")
	}

	// Parse headers from key:value format if provided
	var headers map[string]string
	if len(e.Headers) > 0 {
		headers = make(map[string]string)
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
	}

	// Create the update request
	request := api.MCPEndpointUpdate{}

	// Set fields that are provided
	if e.Name != nil {
		request.SetName(api.NewOptMCPEndpointUpdateName(
			api.NewStringMCPEndpointUpdateName(*e.Name),
		))
	}
	
	if e.Url != nil {
		request.SetURL(api.NewOptMCPEndpointUpdateURL(
			api.NewStringMCPEndpointUpdateURL(*e.Url),
		))
	}
	
	if e.Description != nil {
		request.SetDescription(api.NewOptMCPEndpointUpdateDescription(
			api.NewStringMCPEndpointUpdateDescription(*e.Description),
		))
	}
	
	if len(headers) > 0 {
		request.SetHeaders(api.NewOptMCPEndpointUpdateHeaders(
			api.NewMCPEndpointUpdateHeaders0MCPEndpointUpdateHeaders(api.MCPEndpointUpdateHeaders0(headers)),
		))
	}
	
	if e.DevgraphAuth != nil {
		request.SetDevgraphAuth(api.NewOptMCPEndpointUpdateDevgraphAuth(
			api.NewBoolMCPEndpointUpdateDevgraphAuth(*e.DevgraphAuth),
		))
	}
	
	if e.SupportsResources != nil {
		request.SetSupportsResources(api.NewOptMCPEndpointUpdateSupportsResources(
			api.NewBoolMCPEndpointUpdateSupportsResources(*e.SupportsResources),
		))
	}
	
	if e.OAuthServiceID != nil {
		oauthUUID, err := uuid.Parse(*e.OAuthServiceID)
		if err != nil {
			return fmt.Errorf("invalid OAuth service ID: %w", err)
		}
		request.SetOAuthServiceID(api.NewOptMCPEndpointUpdateOAuthServiceID(
			api.NewUUIDMCPEndpointUpdateOAuthServiceID(oauthUUID),
		))
	}

	// Make the API call
	params := api.UpdateMcpendpointParams{
		McpendpointID: mcpUUID,
	}
	
	resp, err := client.UpdateMcpendpoint(context.Background(), &request, params)
	if err != nil {
		return fmt.Errorf("failed to update MCP endpoint: %w", err)
	}

	// Check the response type
	switch resp.(type) {
	case *api.MCPEndpointResponse:
		fmt.Printf("MCP endpoint '%s' updated successfully.\n", e.Id)
	default:
		return fmt.Errorf("failed to update MCP endpoint")
	}

	return nil
}
