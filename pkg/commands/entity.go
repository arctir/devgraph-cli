package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/arctir/devgraph-cli/pkg/util"
	api "github.com/arctir/go-devgraph/pkg/apis/devgraph/v1"
	"github.com/fatih/color"
	"gopkg.in/yaml.v3"
)

// parseEntityID parses an entity ID in the format [entity://]<group>/<version>/<plural>/<namespace>/<name>
// and returns the individual components
func parseEntityID(entityID string) (group, version, plural, namespace, name string, err error) {
	// Remove optional entity:// prefix
	id := strings.TrimPrefix(entityID, "entity://")
	
	// Split the ID into parts
	parts := strings.Split(id, "/")
	if len(parts) != 5 {
		return "", "", "", "", "", fmt.Errorf("invalid entity ID format: expected <group>/<version>/<plural>/<namespace>/<name>, got: %s", entityID)
	}
	
	return parts[0], parts[1], parts[2], parts[3], parts[4], nil
}

// FilteredEntity represents an entity with only the required fields
type FilteredEntity struct {
	ApiVersion string      `json:"apiVersion" yaml:"apiVersion"`
	Kind       string      `json:"kind" yaml:"kind"`
	Metadata   interface{} `json:"metadata" yaml:"metadata"`
	Spec       interface{} `json:"spec,omitempty" yaml:"spec,omitempty"`
	Status     interface{} `json:"status,omitempty" yaml:"status,omitempty"`
}

// filterEntity creates a FilteredEntity with only the required fields
func filterEntity(entity api.EntityResponse) FilteredEntity {
	filtered := FilteredEntity{
		ApiVersion: entity.ApiVersion,
		Kind:       entity.Kind,
		Metadata:   entity.Metadata,
	}
	
	// Extract actual values from optional types
	if entity.Spec.IsSet() {
		if spec, ok := entity.Spec.Get(); ok {
			filtered.Spec = spec
		}
	}
	
	if entity.Status.IsSet() {
		if status, ok := entity.Status.Get(); ok {
			filtered.Status = status
		}
	}
	
	return filtered
}

// displayEntityList displays a list of entities in a table format
func displayEntityList(entities []api.EntityResponse) error {
	if len(entities) == 0 {
		fmt.Println("No entities found.")
		return nil
	}

	return displayEntitiesAsTable(entities)
}

// displayEntitiesAsTable displays entities in table format
func displayEntitiesAsTable(entities []api.EntityResponse) error {
	// Prepare data for table display
	headers := []string{"Entity ID", "Name", "Namespace", "API Version", "Kind"}
	data := make([]map[string]interface{}, len(entities))
	
	for i, entity := range entities {
		// Use the entity ID provided by the API response
		data[i] = map[string]interface{}{
			"Entity ID":   entity.ID,
			"Name":       entity.Name,
			"Namespace":  entity.Namespace,
			"API Version": entity.ApiVersion,
			"Kind":       entity.Kind,
		}
	}
	
	displayEntityTable(data, headers)
	return nil
}

// displayEntitiesAsYAML displays entities in YAML format with filtered fields
func displayEntitiesAsYAML(entities []api.EntityResponse) error {
	var filteredEntities []FilteredEntity
	for _, entity := range entities {
		filteredEntities = append(filteredEntities, filterEntity(entity))
	}
	
	yamlData, err := yaml.Marshal(filteredEntities)
	if err != nil {
		return fmt.Errorf("failed to marshal entities to YAML: %w", err)
	}
	
	fmt.Print(string(yamlData))
	return nil
}

// displayEntitiesAsJSON displays entities in JSON format with filtered fields
func displayEntitiesAsJSON(entities []api.EntityResponse) error {
	var filteredEntities []FilteredEntity
	for _, entity := range entities {
		filteredEntities = append(filteredEntities, filterEntity(entity))
	}
	
	jsonData, err := json.MarshalIndent(filteredEntities, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal entities to JSON: %w", err)
	}
	
	fmt.Println(string(jsonData))
	return nil
}

// displaySingleEntity displays a single entity in the specified format with filtered fields
func displaySingleEntity(entity api.EntityResponse, outputFormat string) error {
	// First convert to JSON to get clean serialization
	jsonData, err := json.Marshal(entity)
	if err != nil {
		return fmt.Errorf("failed to marshal entity to JSON: %w", err)
	}
	
	// Parse back into a generic map to filter fields
	var entityMap map[string]interface{}
	if err := json.Unmarshal(jsonData, &entityMap); err != nil {
		return fmt.Errorf("failed to unmarshal entity: %w", err)
	}
	
	// Create filtered map with only desired fields
	filteredMap := map[string]interface{}{
		"apiVersion": entityMap["apiVersion"],
		"kind":       entityMap["kind"],
		"metadata":   entityMap["metadata"],
	}
	
	// Add spec and status if they exist
	if spec, exists := entityMap["spec"]; exists && spec != nil {
		filteredMap["spec"] = spec
	}
	if status, exists := entityMap["status"]; exists && status != nil {
		filteredMap["status"] = status
	}
	
	switch strings.ToLower(outputFormat) {
	case "yaml", "yml":
		yamlData, err := yaml.Marshal(filteredMap)
		if err != nil {
			return fmt.Errorf("failed to marshal entity to YAML: %w", err)
		}
		fmt.Print(string(yamlData))
	case "json":
		fallthrough
	default:
		jsonData, err := json.MarshalIndent(filteredMap, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal entity to JSON: %w", err)
		}
		fmt.Println(string(jsonData))
	}
	
	return nil
}

// displayEntityTable creates a table for entities with no truncation on Entity ID column
func displayEntityTable(data []map[string]interface{}, headers []string) {
	if len(data) == 0 {
		fmt.Println("No data to display.")
		return
	}

	// Calculate column widths
	colWidths := make([]int, len(headers))

	// Initialize with header widths
	for i, header := range headers {
		colWidths[i] = len(header)
	}

	// Check data widths
	for _, row := range data {
		for i, header := range headers {
			if val, ok := row[header]; ok {
				var valueStr string
				switch v := val.(type) {
				case string:
					valueStr = v
				case int:
					valueStr = fmt.Sprintf("%d", v)
				case float64:
					valueStr = fmt.Sprintf("%.2f", v)
				default:
					valueStr = fmt.Sprintf("%v", v)
				}

				// Don't truncate Entity ID column, but limit other columns for readability
				if header != "Entity ID" {
					maxWidth := 60
					if len(valueStr) > maxWidth {
						valueStr = valueStr[:maxWidth-3] + "..."
					}
				}

				if len(valueStr) > colWidths[i] {
					colWidths[i] = len(valueStr)
				}
			}
		}
	}

	// Add some spacing
	fmt.Println()

	// Print headers with color
	headerColor := color.New(color.FgBlue, color.Bold)
	for i, header := range headers {
		if i > 0 {
			fmt.Print("  ")
		}
		coloredHeader := headerColor.Sprint(header)
		fmt.Print(coloredHeader)
		padding := colWidths[i] - len(header)
		if padding > 0 {
			fmt.Print(strings.Repeat(" ", padding))
		}
	}
	fmt.Println()

	// Print separator line
	for i := range headers {
		if i > 0 {
			fmt.Print("  ")
		}
		fmt.Print(strings.Repeat("─", colWidths[i]))
	}
	fmt.Println()

	// Print data rows
	gray := color.New(color.FgHiBlack)
	for _, row := range data {
		for i, header := range headers {
			if i > 0 {
				fmt.Print("  ")
			}

			var valueStr string
			if val, ok := row[header]; ok {
				switch v := val.(type) {
				case string:
					valueStr = v
				case int:
					valueStr = fmt.Sprintf("%d", v)
				case float64:
					valueStr = fmt.Sprintf("%.2f", v)
				default:
					valueStr = fmt.Sprintf("%v", v)
				}
			} else {
				valueStr = gray.Sprint("-")
			}

			// Don't truncate Entity ID column
			if header != "Entity ID" {
				maxWidth := 60
				if len(valueStr) > maxWidth {
					valueStr = valueStr[:maxWidth-3] + "..."
				}
			}

			fmt.Printf("%-*s", colWidths[i], valueStr)
		}
		fmt.Println()
	}

	// Add spacing after
	fmt.Println()
}

type EntityCommand struct {
	Create        EntityCreateCommand        `cmd:"create" help:"Create a new entity."`
	List          EntityListCommand          `cmd:"" help:"List entities."`
	Get           EntityGetCommand           `cmd:"get" help:"Get an entity by ID."`
	Delete        EntityDeleteCommand        `cmd:"delete" help:"Delete an entity by ID."`
	Relationships EntityRelationshipsCommand `cmd:"relationships" help:"Show relationships for an entity."`
}

type EntityCreateCommand struct {
	EnvWrapperCommand
	Group     string `arg:"" required:"" help:"Group of the entity (e.g., apps, core, extensions)."`
	Version   string `arg:"" required:"" help:"Version of the entity (e.g., v1, v1beta1)."`
	Namespace string `arg:"" required:"" help:"Namespace of the entity."`
	Plural    string `arg:"" required:"" help:"Plural form of the entity kind (e.g., deployments, services)."`
	FileName  string `arg:"" required:"" help:"Path to the entity JSON file."`
}

type EntityListCommand struct {
	EnvWrapperCommand
	Name          string `flag:"name,n" help:"Filter entities by name."`
	Label         string `flag:"label,l" help:"Filter entities by label selector."`
	FieldSelector string `flag:"field-selector,f" help:"Filter entities by field selector (e.g., 'spec.metadata.owner=team-a')."`
	Limit         int    `flag:"limit" default:"100" help:"Maximum number of entities to return."`
	Offset        int    `flag:"offset" default:"0" help:"Offset for pagination."`
}

type EntityGetCommand struct {
	EnvWrapperCommand
	EntityID string `arg:"" required:"" help:"Entity ID in the format [entity://]<group>/<version>/<plural>/<namespace>/<name>."`
	Output   string `flag:"output,o" default:"json" help:"Output format: json, yaml."`
}

type EntityDeleteCommand struct {
	EnvWrapperCommand
	EntityID string `arg:"" required:"" help:"Entity ID in the format [entity://]<group>/<version>/<plural>/<namespace>/<name>."`
}

type EntityRelationshipsCommand struct {
	EnvWrapperCommand
	EntityID string `arg:"" required:"" help:"Entity ID in the format [entity://]<group>/<version>/<plural>/<namespace>/<name>."`
	Output   string `flag:"output,o" default:"table" help:"Output format: table, json, yaml."`
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

func (e *EntityListCommand) Run() error {
	client, err := util.GetAuthenticatedClient(e.Config)
	if err != nil {
		return fmt.Errorf("failed to create authenticated client: %w", err)
	}

	// Build the parameters for the API call
	params := api.GetEntitiesParams{}
	
	// Set optional filters if provided
	if e.Name != "" {
		params.Name = api.NewOptString(e.Name)
	}
	if e.Label != "" {
		params.Label = api.NewOptString(e.Label)
	}
	if e.FieldSelector != "" {
		params.FieldSelector = api.NewOptString(e.FieldSelector)
	}
	if e.Limit > 0 {
		params.Limit = api.NewOptInt(e.Limit)
	}
	if e.Offset > 0 {
		params.Offset = api.NewOptInt(e.Offset)
	}

	resp, err := client.GetEntities(context.Background(), params)
	if err != nil {
		return fmt.Errorf("failed to list entities: %w", err)
	}

	// Handle the response
	switch r := resp.(type) {
	case *api.EntityResultSetResponse:
		// EntityResultSetResponse contains PrimaryEntities, RelatedEntities, and Relations
		// For the list command, we're primarily interested in PrimaryEntities
		entities := r.PrimaryEntities
		if len(entities) == 0 {
			fmt.Println("No entities found.")
			return nil
		}
		return displayEntityList(entities)
	case *api.GetEntitiesNotFound:
		fmt.Println("No entities found.")
		return nil
	default:
		return fmt.Errorf("unexpected response type: %T", resp)
	}
}

func (e *EntityGetCommand) Run() error {
	client, err := util.GetAuthenticatedClient(e.Config)
	if err != nil {
		return fmt.Errorf("failed to create authenticated client: %w", err)
	}

	// Parse the entity ID to extract individual components
	group, version, plural, namespace, name, err := parseEntityID(e.EntityID)
	if err != nil {
		return err
	}

	params := api.GetEntityParams{
		Group:     group,
		Version:   version,
		Kind:      plural, // Kind is synonymous with plural
		Namespace: namespace,
		Name:      name,
	}
	resp, err := client.GetEntity(context.Background(), params)
	if err != nil {
		return fmt.Errorf("failed to get entity: %w", err)
	}
	// Check if response is successful
	switch r := resp.(type) {
	case *api.EntityResponse:
		return displaySingleEntity(*r, e.Output)
	default:
		return fmt.Errorf("entity not found")
	}
}

func (e *EntityDeleteCommand) Run() error {
	client, err := util.GetAuthenticatedClient(e.Config)
	if err != nil {
		return fmt.Errorf("failed to create authenticated client: %w", err)
	}

	// Parse the entity ID to extract individual components
	group, version, plural, namespace, name, err := parseEntityID(e.EntityID)
	if err != nil {
		return err
	}

	params := api.DeleteEntityParams{
		Group:     group,
		Version:   version,
		Kind:      plural, // Kind is synonymous with plural
		Namespace: namespace,
		Name:      name,
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

	fmt.Printf("Entity '%s' deleted successfully from namespace '%s'.\n", name, namespace)
	return nil
}

func (e *EntityRelationshipsCommand) Run() error {
	client, err := util.GetAuthenticatedClient(e.Config)
	if err != nil {
		return fmt.Errorf("failed to create authenticated client: %w", err)
	}

	// Parse the entity ID to extract individual components
	group, version, plural, namespace, name, err := parseEntityID(e.EntityID)
	if err != nil {
		return err
	}

	// Build the entity reference
	entityRef := fmt.Sprintf("%s/%s/%s/%s/%s", group, version, plural, namespace, name)

	// Instead of trying to filter with field selectors, let's get all entities and filter relationships
	params := api.GetEntitiesParams{
		Limit: api.NewOptInt(1000), // Get more results to ensure we capture all relationships
	}

	resp, err := client.GetEntities(context.Background(), params)
	if err != nil {
		return fmt.Errorf("failed to get entities: %w", err)
	}

	// Handle the response
	switch r := resp.(type) {
	case *api.EntityResultSetResponse:
		// Filter relations that involve our target entity
		var relevantRelations []api.EntityRelationResponse
		for _, relation := range r.Relations {
			if relation.Source.ID == entityRef || relation.Target.ID == entityRef {
				relevantRelations = append(relevantRelations, relation)
			}
		}

		if len(relevantRelations) == 0 {
			fmt.Printf("No relationships found for entity: %s\n", e.EntityID)
			return nil
		}

		return e.displayRelationships(relevantRelations, entityRef)
	case *api.GetEntitiesNotFound:
		fmt.Printf("No relationships found for entity: %s\n", e.EntityID)
		return nil
	default:
		return fmt.Errorf("unexpected response type: %T", resp)
	}
}

func (e *EntityRelationshipsCommand) displayRelationships(relations []api.EntityRelationResponse, targetEntityRef string) error {
	if len(relations) == 0 {
		fmt.Printf("No relationships found for entity: %s\n", e.EntityID)
		return nil
	}

	switch strings.ToLower(e.Output) {
	case "table":
		return e.displayRelationshipsAsTable(relations, targetEntityRef)
	case "yaml", "yml":
		return e.displayRelationshipsAsYAML(relations)
	case "json":
		return e.displayRelationshipsAsJSON(relations)
	default:
		return fmt.Errorf("unsupported output format: %s", e.Output)
	}
}

func (e *EntityRelationshipsCommand) displayRelationshipsAsTable(relations []api.EntityRelationResponse, targetEntityRef string) error {
	headers := []string{"Direction", "Relation Type", "Related Entity", "Namespace"}
	data := make([]map[string]interface{}, 0)

	for _, relation := range relations {
		var direction, relatedEntity string
		
		// Determine direction and related entity
		if relation.Source.ID == targetEntityRef {
			direction = "Outgoing"
			relatedEntity = relation.Target.ID
		} else if relation.Target.ID == targetEntityRef {
			direction = "Incoming"  
			relatedEntity = relation.Source.ID
		} else {
			// This relation doesn't involve our target entity, skip it
			continue
		}

		namespace := ""
		if relation.Namespace.IsSet() {
			if ns, ok := relation.Namespace.Get(); ok {
				namespace = ns
			}
		}

		data = append(data, map[string]interface{}{
			"Direction":      direction,
			"Relation Type":  relation.Relation,
			"Related Entity": relatedEntity,
			"Namespace":      namespace,
		})
	}

	if len(data) == 0 {
		fmt.Printf("No relationships found for entity: %s\n", e.EntityID)
		return nil
	}

	displayEntityTable(data, headers)
	return nil
}

func (e *EntityRelationshipsCommand) displayRelationshipsAsYAML(relations []api.EntityRelationResponse) error {
	yamlData, err := yaml.Marshal(relations)
	if err != nil {
		return fmt.Errorf("failed to marshal relationships to YAML: %w", err)
	}
	
	fmt.Print(string(yamlData))
	return nil
}

func (e *EntityRelationshipsCommand) displayRelationshipsAsJSON(relations []api.EntityRelationResponse) error {
	jsonData, err := json.MarshalIndent(relations, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal relationships to JSON: %w", err)
	}
	
	fmt.Println(string(jsonData))
	return nil
}
