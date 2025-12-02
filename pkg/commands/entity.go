package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sync"

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

// FilteredEntityDefinition represents an entity definition with only the required fields
type FilteredEntityDefinition struct {
	Group       string      `json:"group" yaml:"group"`
	Kind        string      `json:"kind" yaml:"kind"`
	ListKind    string      `json:"listKind" yaml:"listKind"`
	Plural      string      `json:"plural,omitempty" yaml:"plural,omitempty"`
	Singular    string      `json:"singular" yaml:"singular"`
	Name        string      `json:"name,omitempty" yaml:"name,omitempty"`
	Description string      `json:"description,omitempty" yaml:"description,omitempty"`
	Spec        interface{} `json:"spec" yaml:"spec"`
	Storage     bool        `json:"storage,omitempty" yaml:"storage,omitempty"`
	Served      bool        `json:"served,omitempty" yaml:"served,omitempty"`
}

// FilteredEntityRelation represents an entity relation with only the required fields
type FilteredEntityRelation struct {
	Namespace string `json:"namespace,omitempty" yaml:"namespace,omitempty"`
	Relation  string `json:"relation" yaml:"relation"`
	Source    string `json:"source" yaml:"source"`
	Target    string `json:"target" yaml:"target"`
}

// filterEntity creates a FilteredEntity with only the required fields
func filterEntity(entity api.EntityResponse) FilteredEntity {
	filtered := FilteredEntity{
		ApiVersion: entity.ApiVersion,
		Kind:       entity.Kind,
		Metadata:   cleanMetadata(entity.Metadata),
	}

	// Extract actual values from optional types
	if entity.Spec.IsSet() {
		if spec, ok := entity.Spec.Get(); ok {
			filtered.Spec = cleanSpec(spec)
		}
	}

	if entity.Status.IsSet() {
		if status, ok := entity.Status.Get(); ok {
			filtered.Status = cleanStatus(status)
		}
	}

	return filtered
}

// filterEntityDefinition creates a FilteredEntityDefinition with only the required fields
func filterEntityDefinition(def api.EntityDefinitionResponse) FilteredEntityDefinition {
	filtered := FilteredEntityDefinition{
		Group:    def.Group,
		Kind:     def.Kind,
		ListKind: def.ListKind,
		Singular: def.Singular,
		Spec:     cleanDefinitionSpec(def.Spec),
	}

	// Handle optional plural
	if def.Plural.IsSet() {
		if plural, ok := def.Plural.Get(); ok {
			filtered.Plural = plural
		}
	}

	// Handle optional name
	if def.Name.IsSet() {
		if name, ok := def.Name.Get(); ok {
			filtered.Name = name
		}
	}

	// Handle optional description
	if def.Description.IsSet() {
		if desc, ok := def.Description.Get(); ok {
			filtered.Description = desc
		}
	}

	// Handle optional storage
	if def.Storage.IsSet() {
		if storage, ok := def.Storage.Get(); ok {
			filtered.Storage = storage
		}
	}

	// Handle optional served
	if def.Served.IsSet() {
		if served, ok := def.Served.Get(); ok {
			filtered.Served = served
		}
	}

	return filtered
}

// cleanDefinitionSpec processes definition spec fields
func cleanDefinitionSpec(spec api.EntityDefinitionResponseSpec) map[string]interface{} {
	result := make(map[string]interface{})

	// Convert spec to map for processing
	specBytes, err := json.Marshal(spec)
	if err != nil {
		return result
	}

	var specMap map[string]interface{}
	if err := json.Unmarshal(specBytes, &specMap); err != nil {
		return result
	}

	// Process each field to clean up byte arrays and optional wrappers
	for key, value := range specMap {
		result[key] = cleanValue(value)
	}

	return result
}

// filterEntityRelation creates a FilteredEntityRelation with only the required fields
func filterEntityRelation(rel api.EntityRelationResponse) FilteredEntityRelation {
	filtered := FilteredEntityRelation{
		Relation: rel.Relation,
		Source:   rel.Source.ID,
		Target:   rel.Target.ID,
	}

	// Handle optional namespace
	if rel.Namespace.IsSet() {
		if ns, ok := rel.Namespace.Get(); ok {
			filtered.Namespace = ns
		}
	}

	return filtered
}

// cleanMetadata removes the 'set' wrapper from optional fields in metadata
func cleanMetadata(metadata api.EntityMetadata) map[string]interface{} {
	result := map[string]interface{}{
		"name":      metadata.Name,
		"namespace": metadata.Namespace,
	}

	// Handle optional labels
	if metadata.Labels.IsSet() {
		if labels, ok := metadata.Labels.Get(); ok && len(labels) > 0 {
			result["labels"] = labels
		}
	}

	// Handle optional annotations
	if metadata.Annotations.IsSet() {
		if annotations, ok := metadata.Annotations.Get(); ok && len(annotations) > 0 {
			result["annotations"] = annotations
		}
	}

	return result
}

// cleanSpec processes spec fields to ensure proper serialization
func cleanSpec(spec api.EntityResponseSpec) map[string]interface{} {
	result := make(map[string]interface{})

	// Convert spec to map for processing
	specBytes, err := json.Marshal(spec)
	if err != nil {
		return result
	}

	var specMap map[string]interface{}
	if err := json.Unmarshal(specBytes, &specMap); err != nil {
		return result
	}

	// Process each field to clean up byte arrays and optional wrappers
	for key, value := range specMap {
		result[key] = cleanValue(value)
	}

	return result
}

// cleanStatus processes status fields to ensure proper serialization
func cleanStatus(status api.EntityStatus) map[string]interface{} {
	result := make(map[string]interface{})

	// Convert status to map for processing
	statusBytes, err := json.Marshal(status)
	if err != nil {
		return result
	}

	var statusMap map[string]interface{}
	if err := json.Unmarshal(statusBytes, &statusMap); err != nil {
		return result
	}

	// Process each field to clean up byte arrays and optional wrappers
	for key, value := range statusMap {
		result[key] = cleanValue(value)
	}

	return result
}

// cleanValue recursively cleans values by converting byte arrays to strings
// and removing 'set' wrappers from optional fields
func cleanValue(value interface{}) interface{} {
	switch v := value.(type) {
	case map[string]interface{}:
		// Check if this is an optional wrapper with 'set' and 'value' fields
		if set, hasSet := v["set"].(bool); hasSet && set {
			if val, hasValue := v["value"]; hasValue {
				return cleanValue(val)
			}
		}

		// Otherwise, recursively clean the map
		cleaned := make(map[string]interface{})
		for key, val := range v {
			cleaned[key] = cleanValue(val)
		}
		return cleaned

	case []interface{}:
		// Check if this looks like a byte array (all numbers 0-255)
		if len(v) > 0 {
			allBytes := true
			bytes := make([]byte, len(v))
			for i, item := range v {
				if num, ok := item.(float64); ok && num >= 0 && num <= 255 && num == float64(int(num)) {
					bytes[i] = byte(num)
				} else {
					allBytes = false
					break
				}
			}

			if allBytes {
				return string(bytes)
			}
		}

		// Otherwise, recursively clean the array
		cleaned := make([]interface{}, len(v))
		for i, val := range v {
			cleaned[i] = cleanValue(val)
		}
		return cleaned

	default:
		return v
	}
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
			"Name":        entity.Name,
			"Namespace":   entity.Namespace,
			"API Version": entity.ApiVersion,
			"Kind":        entity.Kind,
		}
	}

	displayEntityTable(data, headers)
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
	Backup        EntityBackupCommand        `cmd:"backup" help:"Backup entities to a directory."`
	Restore       EntityRestoreCommand       `cmd:"restore" help:"Restore entities from a backup directory."`
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
	Limit         int    `flag:"limit" default:"1000" help:"Maximum number of entities to return."`
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

type EntityBackupCommand struct {
	EnvWrapperCommand
	OutputDir     string `arg:"" required:"" help:"Path to output backup directory."`
	Name          string `flag:"name,n" help:"Filter entities by name."`
	Label         string `flag:"label,l" help:"Filter entities by label selector."`
	FieldSelector string `flag:"field-selector,f" help:"Filter entities by field selector."`
	Format        string `flag:"format" default:"yaml" help:"Output format: json, yaml."`
}

type EntityRestoreCommand struct {
	EnvWrapperCommand
	InputDir string `arg:"" required:"" help:"Path to backup directory to restore."`
	DryRun   bool   `flag:"dry-run" help:"Show what would be restored without actually restoring."`
	Workers  int    `flag:"workers,w" default:"10" help:"Number of concurrent workers for restore operations."`
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

	fmt.Printf("✅ Entity '%s' created successfully in namespace '%s'.\n", entity.Metadata.Name, e.Namespace)
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
	case *api.EntityWithRelationsResponse:
		return displaySingleEntity(r.Entity, e.Output)
	case *api.GetEntityNotFound:
		return fmt.Errorf("entity not found")
	case *api.HTTPValidationError:
		return fmt.Errorf("validation error: %v", r.Detail)
	default:
		return fmt.Errorf("unexpected response type")
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

	fmt.Printf("✅ Entity '%s' deleted successfully from namespace '%s'.\n", name, namespace)
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

func (e *EntityBackupCommand) Run() error {
	client, err := util.GetAuthenticatedClient(e.Config)
	if err != nil {
		return fmt.Errorf("failed to create authenticated client: %w", err)
	}

	// Create backup directory structure
	err = os.MkdirAll(e.OutputDir, 0755)
	if err != nil {
		return fmt.Errorf("failed to create backup directory: %w", err)
	}

	definitionsDir := fmt.Sprintf("%s/definitions", e.OutputDir)
	err = os.MkdirAll(definitionsDir, 0755)
	if err != nil {
		return fmt.Errorf("failed to create definitions directory: %w", err)
	}

	entitiesDir := fmt.Sprintf("%s/entities", e.OutputDir)
	err = os.MkdirAll(entitiesDir, 0755)
	if err != nil {
		return fmt.Errorf("failed to create entities directory: %w", err)
	}

	relationsDir := fmt.Sprintf("%s/relations", e.OutputDir)
	err = os.MkdirAll(relationsDir, 0755)
	if err != nil {
		return fmt.Errorf("failed to create relations directory: %w", err)
	}

	// Determine file extension
	ext := ".yaml"
	if e.Format == "json" {
		ext = ".json"
	}

	// Fetch and backup entity definitions
	defResp, err := client.GetEntityDefinitions(context.Background())
	if err != nil {
		return fmt.Errorf("failed to get entity definitions: %w", err)
	}

	var definitions []api.EntityDefinitionResponse
	switch r := defResp.(type) {
	case *api.GetEntityDefinitionsOKApplicationJSON:
		definitions = *r
	case *api.GetEntityDefinitionsNotFound:
		fmt.Println("No entity definitions found.")
	default:
		return fmt.Errorf("unexpected response type for definitions: %T", defResp)
	}

	// Write entity definitions
	defSuccessCount := 0
	for _, def := range definitions {
		filtered := filterEntityDefinition(def)

		// Create filename: <group>_<kind>.<ext>
		filename := fmt.Sprintf("%s_%s%s",
			def.Group,
			strings.ToLower(def.Kind),
			ext)

		filepath := fmt.Sprintf("%s/%s", definitionsDir, filename)

		// Marshal definition
		var data []byte
		switch e.Format {
		case "json":
			data, err = json.MarshalIndent(filtered, "", "  ")
		case "yaml":
			data, err = yaml.Marshal(filtered)
		default:
			return fmt.Errorf("unsupported format: %s (use json or yaml)", e.Format)
		}

		if err != nil {
			fmt.Printf("Warning: failed to marshal definition %s/%s: %v\n", def.Group, def.Kind, err)
			continue
		}

		// Write to file
		err = os.WriteFile(filepath, data, 0600)
		if err != nil {
			fmt.Printf("Warning: failed to write definition %s/%s: %v\n", def.Group, def.Kind, err)
			continue
		}

		defSuccessCount++
	}

	// Build query parameters for entities
	params := api.GetEntitiesParams{}

	if e.Name != "" {
		params.Name = api.NewOptString(e.Name)
	}
	if e.Label != "" {
		params.Label = api.NewOptString(e.Label)
	}
	if e.FieldSelector != "" {
		params.FieldSelector = api.NewOptString(e.FieldSelector)
	}

	// Fetch all entities
	resp, err := client.GetEntities(context.Background(), params)
	if err != nil {
		return fmt.Errorf("failed to get entities: %w", err)
	}

	var entities []api.EntityResponse
	switch r := resp.(type) {
	case *api.EntityResultSetResponse:
		entities = r.PrimaryEntities
	case *api.GetEntitiesNotFound:
		fmt.Println("No entities found to backup.")
	default:
		return fmt.Errorf("unexpected response type: %T", resp)
	}

	// Write each entity to a separate file
	entitySuccessCount := 0
	for _, entity := range entities {
		filtered := filterEntity(entity)

		// Create filename: <group>_<version>_<namespace>_<kind>_<name>.<ext>
		filename := fmt.Sprintf("%s_%s_%s_%s_%s%s",
			entity.Group,
			entity.Version,
			entity.Namespace,
			strings.ToLower(entity.Kind),
			entity.Name,
			ext)

		filepath := fmt.Sprintf("%s/%s", entitiesDir, filename)

		// Marshal entity
		var data []byte
		switch e.Format {
		case "json":
			data, err = json.MarshalIndent(filtered, "", "  ")
		case "yaml":
			data, err = yaml.Marshal(filtered)
		default:
			return fmt.Errorf("unsupported format: %s (use json or yaml)", e.Format)
		}

		if err != nil {
			fmt.Printf("Warning: failed to marshal entity %s/%s: %v\n", entity.Namespace, entity.Name, err)
			continue
		}

		// Write to file
		err = os.WriteFile(filepath, data, 0600)
		if err != nil {
			fmt.Printf("Warning: failed to write entity %s/%s: %v\n", entity.Namespace, entity.Name, err)
			continue
		}

		entitySuccessCount++
	}

	// Fetch all entities again to get their relations
	// We need to get all relations from the entity result set
	allParams := api.GetEntitiesParams{
		Limit: api.NewOptInt(10000), // Get a large number to capture all relations
	}

	allResp, err := client.GetEntities(context.Background(), allParams)
	if err != nil {
		fmt.Printf("Warning: failed to get relations: %v\n", err)
	}

	var relations []api.EntityRelationResponse
	switch r := allResp.(type) {
	case *api.EntityResultSetResponse:
		relations = r.Relations
	}

	// Write relationships
	relSuccessCount := 0
	if len(relations) > 0 {
		// Write all relations to a single file
		var filteredRelations []FilteredEntityRelation
		for _, rel := range relations {
			filteredRelations = append(filteredRelations, filterEntityRelation(rel))
		}

		filename := fmt.Sprintf("relations%s", ext)
		filepath := fmt.Sprintf("%s/%s", relationsDir, filename)

		// Marshal relations
		var data []byte
		switch e.Format {
		case "json":
			data, err = json.MarshalIndent(filteredRelations, "", "  ")
		case "yaml":
			data, err = yaml.Marshal(filteredRelations)
		default:
			return fmt.Errorf("unsupported format: %s (use json or yaml)", e.Format)
		}

		if err != nil {
			fmt.Printf("Warning: failed to marshal relations: %v\n", err)
		} else {
			// Write to file
			err = os.WriteFile(filepath, data, 0600)
			if err != nil {
				fmt.Printf("Warning: failed to write relations: %v\n", err)
			} else {
				relSuccessCount = len(filteredRelations)
			}
		}
	}

	fmt.Printf("Successfully backed up %d definitions, %d entities, and %d relations to %s\n",
		defSuccessCount, entitySuccessCount, relSuccessCount, e.OutputDir)
	return nil
}

func (e *EntityRestoreCommand) Run() error {
	client, err := util.GetAuthenticatedClient(e.Config)
	if err != nil {
		return fmt.Errorf("failed to create authenticated client: %w", err)
	}

	// Check for definitions directory
	definitionsDir := fmt.Sprintf("%s/definitions", e.InputDir)
	entitiesDir := fmt.Sprintf("%s/entities", e.InputDir)

	// Load entity definitions first
	var definitions []FilteredEntityDefinition
	if defFiles, err := os.ReadDir(definitionsDir); err == nil {
		for _, file := range defFiles {
			if file.IsDir() {
				continue
			}

			// Only process .yaml, .yml, and .json files
			filename := file.Name()
			if !strings.HasSuffix(filename, ".yaml") &&
				!strings.HasSuffix(filename, ".yml") &&
				!strings.HasSuffix(filename, ".json") {
				continue
			}

			filepath := fmt.Sprintf("%s/%s", definitionsDir, filename)
			data, err := os.ReadFile(filepath)
			if err != nil {
				fmt.Printf("Warning: failed to read definition file %s: %v\n", filename, err)
				continue
			}

			var def FilteredEntityDefinition
			err = yaml.Unmarshal(data, &def)
			if err != nil {
				fmt.Printf("Warning: failed to parse definition file %s: %v\n", filename, err)
				continue
			}

			definitions = append(definitions, def)
		}
	}

	// Load entity files
	var entities []FilteredEntity

	// Try new structure first (entities subdirectory)
	if entFiles, err := os.ReadDir(entitiesDir); err == nil {
		for _, file := range entFiles {
			if file.IsDir() {
				continue
			}

			filename := file.Name()
			if !strings.HasSuffix(filename, ".yaml") &&
				!strings.HasSuffix(filename, ".yml") &&
				!strings.HasSuffix(filename, ".json") {
				continue
			}

			filepath := fmt.Sprintf("%s/%s", entitiesDir, filename)
			data, err := os.ReadFile(filepath)
			if err != nil {
				fmt.Printf("Warning: failed to read entity file %s: %v\n", filename, err)
				continue
			}

			var entity FilteredEntity
			err = yaml.Unmarshal(data, &entity)
			if err != nil {
				fmt.Printf("Warning: failed to parse entity file %s: %v\n", filename, err)
				continue
			}

			entities = append(entities, entity)
		}
	} else {
		// Fall back to old structure (flat directory)
		files, err := os.ReadDir(e.InputDir)
		if err != nil {
			return fmt.Errorf("failed to read backup directory: %w", err)
		}

		for _, file := range files {
			if file.IsDir() {
				continue
			}

			filename := file.Name()
			if !strings.HasSuffix(filename, ".yaml") &&
				!strings.HasSuffix(filename, ".yml") &&
				!strings.HasSuffix(filename, ".json") {
				continue
			}

			filepath := fmt.Sprintf("%s/%s", e.InputDir, filename)
			data, err := os.ReadFile(filepath)
			if err != nil {
				fmt.Printf("Warning: failed to read file %s: %v\n", filename, err)
				continue
			}

			var entity FilteredEntity
			err = yaml.Unmarshal(data, &entity)
			if err != nil {
				fmt.Printf("Warning: failed to parse file %s: %v\n", filename, err)
				continue
			}

			entities = append(entities, entity)
		}
	}

	// Load relations
	var relations []FilteredEntityRelation
	relationsDir := fmt.Sprintf("%s/relations", e.InputDir)
	if relFiles, err := os.ReadDir(relationsDir); err == nil {
		for _, file := range relFiles {
			if file.IsDir() {
				continue
			}

			filename := file.Name()
			if !strings.HasSuffix(filename, ".yaml") &&
				!strings.HasSuffix(filename, ".yml") &&
				!strings.HasSuffix(filename, ".json") {
				continue
			}

			filepath := fmt.Sprintf("%s/%s", relationsDir, filename)
			data, err := os.ReadFile(filepath)
			if err != nil {
				fmt.Printf("Warning: failed to read relations file %s: %v\n", filename, err)
				continue
			}

			var rels []FilteredEntityRelation
			err = yaml.Unmarshal(data, &rels)
			if err != nil {
				fmt.Printf("Warning: failed to parse relations file %s: %v\n", filename, err)
				continue
			}

			relations = append(relations, rels...)
		}
	}

	if e.DryRun {
		fmt.Printf("Dry run: Would restore %d definitions, %d entities, and %d relations:\n", len(definitions), len(entities), len(relations))
		for _, def := range definitions {
			fmt.Printf("  Definition: %s/%s\n", def.Group, def.Kind)
		}
		for _, entity := range entities {
			if metadata, ok := entity.Metadata.(map[string]interface{}); ok {
				fmt.Printf("  Entity: %s/%s (%s)\n", metadata["namespace"], metadata["name"], entity.Kind)
			}
		}
		for _, rel := range relations {
			fmt.Printf("  Relation: %s -> %s (%s)\n", rel.Source, rel.Target, rel.Relation)
		}
		return nil
	}

	// Restore entity definitions first with concurrent workers
	defSuccessCount := 0
	defFailCount := 0

	if len(definitions) > 0 {
		type defResult struct {
			def     FilteredEntityDefinition
			success bool
			err     error
		}

		defChan := make(chan FilteredEntityDefinition, len(definitions))
		resultChan := make(chan defResult, len(definitions))

		// Start worker pool
		var wg sync.WaitGroup
		for i := 0; i < e.Workers; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for def := range defChan {
					// Convert definition to API type
					apiDef := &api.EntityDefinitionSpec{
						Group:    def.Group,
						Kind:     def.Kind,
						ListKind: def.ListKind,
						Singular: def.Singular,
					}

					// Handle optional plural
					if def.Plural != "" {
						apiDef.Plural.SetTo(def.Plural)
					}

					// Handle optional name
					if def.Name != "" {
						apiDef.Name.SetTo(def.Name)
					}

					// Handle optional description
					if def.Description != "" {
						apiDef.Description.SetTo(def.Description)
					}

					// Convert spec
					if def.Spec != nil {
						if specBytes, err := json.Marshal(def.Spec); err == nil {
							var defSpec api.EntityDefinitionSpecSpec
							if err := json.Unmarshal(specBytes, &defSpec); err == nil {
								apiDef.Spec = defSpec
							}
						}
					}

					// Handle optional storage
					if def.Storage {
						apiDef.Storage.SetTo(def.Storage)
					}

					// Handle optional served
					if def.Served {
						apiDef.Served.SetTo(def.Served)
					}

					// Create definition via API
					resp, err := client.CreateEntityDefinition(context.Background(), apiDef)

					result := defResult{def: def}
					if err != nil {
						result.err = err
						result.success = false
					} else {
						switch resp.(type) {
						case *api.EntityDefinitionResponse:
							result.success = true
						default:
							result.success = false
							result.err = fmt.Errorf("unexpected response type")
						}
					}
					resultChan <- result
				}
			}()
		}

		// Send definitions to workers
		for _, def := range definitions {
			defChan <- def
		}
		close(defChan)

		// Wait for all workers to complete
		go func() {
			wg.Wait()
			close(resultChan)
		}()

		// Collect results
		for result := range resultChan {
			if result.success {
				fmt.Printf("✅ Restored definition %s/%s\n", result.def.Group, result.def.Kind)
				defSuccessCount++
			} else {
				if result.err != nil {
					fmt.Printf("✗ Failed to restore definition %s/%s: %v\n", result.def.Group, result.def.Kind, result.err)
				} else {
					fmt.Printf("✗ Failed to restore definition %s/%s: unexpected response\n", result.def.Group, result.def.Kind)
				}
				defFailCount++
			}
		}
	}

	// Build a map of kind to plural from the restored definitions
	kindToPluralMap := make(map[string]string)
	for _, def := range definitions {
		key := fmt.Sprintf("%s/%s", def.Group, def.Kind)
		plural := def.Plural
		if plural == "" {
			// Fall back to simple pluralization if not specified
			plural = strings.ToLower(def.Kind) + "s"
		}
		kindToPluralMap[key] = plural
	}

	// Restore entities with concurrent workers
	entitySuccessCount := 0
	entityFailCount := 0

	if len(entities) > 0 {
		type entityResult struct {
			namespace string
			name      string
			kind      string
			success   bool
			err       error
		}

		entityChan := make(chan FilteredEntity, len(entities))
		resultChan := make(chan entityResult, len(entities))

		// Start worker pool
		var wg sync.WaitGroup
		for i := 0; i < e.Workers; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for entity := range entityChan {
					// Extract metadata
					metadata, ok := entity.Metadata.(map[string]interface{})
					if !ok {
						resultChan <- entityResult{
							success: false,
							err:     fmt.Errorf("invalid metadata format"),
						}
						continue
					}

					namespace, _ := metadata["namespace"].(string)
					name, _ := metadata["name"].(string)

					// Split apiVersion into group and version
					parts := strings.Split(entity.ApiVersion, "/")
					var group, version string
					if len(parts) == 2 {
						group = parts[0]
						version = parts[1]
					} else {
						version = parts[0]
						group = "core"
					}

					// Look up plural from definitions map
					key := fmt.Sprintf("%s/%s", group, entity.Kind)
					plural, ok := kindToPluralMap[key]
					if !ok {
						// Fall back to simple pluralization if definition not found
						plural = strings.ToLower(entity.Kind) + "s"
					}

					// Convert entity to API Entity type
					apiEntity := &api.Entity{
						ApiVersion: entity.ApiVersion,
						Kind:       entity.Kind,
					}

					// Convert metadata
					if metadataBytes, err := json.Marshal(entity.Metadata); err == nil {
						var entityMetadata api.EntityMetadata
						if err := json.Unmarshal(metadataBytes, &entityMetadata); err == nil {
							apiEntity.Metadata = entityMetadata
						}
					}

					// Convert spec if present
					if entity.Spec != nil {
						if specBytes, err := json.Marshal(entity.Spec); err == nil {
							var entitySpec api.EntitySpec
							if err := json.Unmarshal(specBytes, &entitySpec); err == nil {
								apiEntity.Spec.SetTo(entitySpec)
							}
						}
					}

					// Convert status if present
					if entity.Status != nil {
						if statusBytes, err := json.Marshal(entity.Status); err == nil {
							var entityStatus api.EntityStatus
							if err := json.Unmarshal(statusBytes, &entityStatus); err == nil {
								apiEntity.Status.SetTo(entityStatus)
							}
						}
					}

					// Create entity via API
					params := api.CreateEntityParams{
						Group:     group,
						Version:   version,
						Namespace: namespace,
						Plural:    plural,
					}

					resp, err := client.CreateEntity(context.Background(), apiEntity, params)

					result := entityResult{
						namespace: namespace,
						name:      name,
						kind:      entity.Kind,
					}

					if err != nil {
						result.err = err
						result.success = false
					} else {
						switch resp.(type) {
						case *api.EntityResponse:
							result.success = true
						default:
							result.success = false
							result.err = fmt.Errorf("unexpected response type")
						}
					}
					resultChan <- result
				}
			}()
		}

		// Send entities to workers
		for _, entity := range entities {
			entityChan <- entity
		}
		close(entityChan)

		// Wait for all workers to complete
		go func() {
			wg.Wait()
			close(resultChan)
		}()

		// Collect results
		for result := range resultChan {
			if result.success {
				fmt.Printf("✅ Restored %s/%s (%s)\n", result.namespace, result.name, result.kind)
				entitySuccessCount++
			} else {
				if result.err != nil {
					fmt.Printf("✗ Failed to restore %s/%s: %v\n", result.namespace, result.name, result.err)
				} else {
					fmt.Printf("✗ Failed to restore %s/%s: unexpected response\n", result.namespace, result.name)
				}
				entityFailCount++
			}
		}
	}

	// Restore relationships after entities with concurrent workers
	relSuccessCount := 0
	relFailCount := 0

	if len(relations) > 0 {
		type relResult struct {
			source   string
			target   string
			relation string
			success  bool
			err      error
		}

		relChan := make(chan FilteredEntityRelation, len(relations))
		resultChan := make(chan relResult, len(relations))

		// Start worker pool
		var wg sync.WaitGroup
		for i := 0; i < e.Workers; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for rel := range relChan {
					// Parse source and target entity IDs
					sourceParts := strings.Split(rel.Source, "/")
					targetParts := strings.Split(rel.Target, "/")

					if len(sourceParts) < 5 || len(targetParts) < 5 {
						resultChan <- relResult{
							source:   rel.Source,
							target:   rel.Target,
							relation: rel.Relation,
							success:  false,
							err:      fmt.Errorf("invalid relation format"),
						}
						continue
					}

					// Build apiVersion from group/version
					sourceApiVersion := fmt.Sprintf("%s/%s", sourceParts[0], sourceParts[1])
					targetApiVersion := fmt.Sprintf("%s/%s", targetParts[0], targetParts[1])

					// Create entity references
					sourceRef := api.EntityReference{
						ApiVersion: sourceApiVersion,
						Kind:       sourceParts[2],
						Name:       sourceParts[4],
					}
					sourceRef.Namespace.SetTo(sourceParts[3])

					targetRef := api.EntityReference{
						ApiVersion: targetApiVersion,
						Kind:       targetParts[2],
						Name:       targetParts[4],
					}
					targetRef.Namespace.SetTo(targetParts[3])

					// Create relation
					apiRel := &api.EntityRelation{
						Relation: rel.Relation,
						Source:   sourceRef,
						Target:   targetRef,
					}

					// Use source entity's namespace for the relation (no cross-namespace relationships)
					namespace := sourceParts[3]

					// Set namespace on relation object
					apiRel.Namespace.SetTo(namespace)

					// Create relation via API with namespace parameter
					params := api.CreateEntityRelationParams{
						Namespace: namespace,
					}
					resp, err := client.CreateEntityRelation(context.Background(), apiRel, params)

					result := relResult{
						source:   rel.Source,
						target:   rel.Target,
						relation: rel.Relation,
					}

					if err != nil {
						result.err = err
						result.success = false
					} else {
						switch resp.(type) {
						case *api.EntityRelationResponse:
							result.success = true
						default:
							result.success = false
							result.err = fmt.Errorf("unexpected response type")
						}
					}
					resultChan <- result
				}
			}()
		}

		// Send relations to workers
		for _, rel := range relations {
			relChan <- rel
		}
		close(relChan)

		// Wait for all workers to complete
		go func() {
			wg.Wait()
			close(resultChan)
		}()

		// Collect results
		for result := range resultChan {
			if result.success {
				fmt.Printf("✅ Restored relation %s -> %s (%s)\n", result.source, result.target, result.relation)
				relSuccessCount++
			} else {
				if result.err != nil {
					fmt.Printf("✗ Failed to restore relation %s -> %s (%s): %v\n", result.source, result.target, result.relation, result.err)
				} else {
					fmt.Printf("✗ Failed to restore relation %s -> %s (%s): unexpected response\n", result.source, result.target, result.relation)
				}
				relFailCount++
			}
		}
	}

	fmt.Printf("\nRestore complete:\n")
	fmt.Printf("  Definitions: %d succeeded, %d failed\n", defSuccessCount, defFailCount)
	fmt.Printf("  Entities: %d succeeded, %d failed\n", entitySuccessCount, entityFailCount)
	fmt.Printf("  Relations: %d succeeded, %d failed\n", relSuccessCount, relFailCount)

	if defFailCount > 0 || entityFailCount > 0 || relFailCount > 0 {
		return fmt.Errorf("%d definitions, %d entities, and %d relations failed to restore", defFailCount, entityFailCount, relFailCount)
	}

	return nil
}
