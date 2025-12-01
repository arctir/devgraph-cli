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

// RelationCommand is the top-level command for managing entity relations
type RelationCommand struct {
	Create RelationCreateCommand `cmd:"create" help:"Create a new relation between entities."`
	List   RelationListCommand   `cmd:"" help:"List entity relations."`
	Delete RelationDeleteCommand `cmd:"delete" help:"Delete a relation between entities."`
}

// RelationCreateCommand creates a new relation between two entities
type RelationCreateCommand struct {
	EnvWrapperCommand
	Relation  string `arg:"" required:"" help:"Type of relation (e.g., DEPENDS_ON, USES, OWNS)."`
	Source    string `arg:"" required:"" help:"Source entity ID in format: <group>/<version>/<plural>/<namespace>/<name>"`
	Target    string `arg:"" required:"" help:"Target entity ID in format: <group>/<version>/<plural>/<namespace>/<name>"`
	Namespace string `flag:"namespace,n" help:"Namespace for the relation (optional)."`
}

// RelationListCommand lists entity relations with optional filtering
type RelationListCommand struct {
	EnvWrapperCommand
	Source string `flag:"source,s" help:"Filter by source entity ID."`
	Target string `flag:"target,t" help:"Filter by target entity ID."`
	Label  string `flag:"label,l" help:"Filter relations by label selector."`
	Limit  int    `flag:"limit" default:"1000" help:"Maximum number of relations to return."`
	Offset int    `flag:"offset" default:"0" help:"Offset for pagination."`
	Output string `flag:"output,o" default:"table" help:"Output format: table, json, yaml."`
}

// RelationDeleteCommand deletes a relation between two entities
type RelationDeleteCommand struct {
	EnvWrapperCommand
	Relation  string `arg:"" required:"" help:"Type of relation to delete."`
	Source    string `arg:"" required:"" help:"Source entity ID in format: <group>/<version>/<plural>/<namespace>/<name>"`
	Target    string `arg:"" required:"" help:"Target entity ID in format: <group>/<version>/<plural>/<namespace>/<name>"`
	Namespace string `flag:"namespace,n" help:"Namespace for the relation (optional)."`
}

// parseEntityReference converts an entity ID string to an EntityReference
func parseEntityReference(entityID string) (api.EntityReference, error) {
	group, version, plural, namespace, name, err := parseEntityID(entityID)
	if err != nil {
		return api.EntityReference{}, err
	}

	// Construct apiVersion from group and version
	apiVersion := fmt.Sprintf("%s/%s", group, version)

	ref := api.EntityReference{
		ApiVersion: apiVersion,
		Kind:       plural,
		Name:       name,
		Namespace:  api.NewOptString(namespace),
	}

	return ref, nil
}

// Run executes the create relation command
func (r *RelationCreateCommand) Run() error {
	client, err := util.GetAuthenticatedClient(r.Config)
	if err != nil {
		return fmt.Errorf("failed to create authenticated client: %w", err)
	}

	// Parse source and target entity references
	source, err := parseEntityReference(r.Source)
	if err != nil {
		return fmt.Errorf("invalid source entity ID: %w", err)
	}

	target, err := parseEntityReference(r.Target)
	if err != nil {
		return fmt.Errorf("invalid target entity ID: %w", err)
	}

	// Build the relation request
	relation := api.EntityRelation{
		Relation: r.Relation,
		Source:   source,
		Target:   target,
	}

	// Determine namespace for the API call
	namespace := r.Namespace
	if namespace == "" {
		// Use the namespace from the source entity
		if ns, ok := source.Namespace.Get(); ok {
			namespace = ns
		}
	}

	if namespace != "" {
		relation.Namespace = api.NewOptString(namespace)
	}

	params := api.CreateEntityRelationParams{
		Namespace: namespace,
	}

	// Create the relation
	resp, err := client.CreateEntityRelation(context.Background(), &relation, params)
	if err != nil {
		return fmt.Errorf("failed to create relation: %w", err)
	}

	// Handle the response
	switch r := resp.(type) {
	case *api.EntityRelationResponse:
		fmt.Printf("✅ Relation created successfully\n")
		fmt.Printf("   Relation: %s\n", r.Relation)
		fmt.Printf("   Source:   %s\n", r.Source.ID)
		fmt.Printf("   Target:   %s\n", r.Target.ID)
		if r.Namespace.IsSet() {
			if ns, ok := r.Namespace.Get(); ok {
				fmt.Printf("   Namespace: %s\n", ns)
			}
		}
		return nil
	case *api.CreateEntityRelationNotFound:
		return fmt.Errorf("entity not found")
	case *api.HTTPValidationError:
		return fmt.Errorf("validation error: %v", r.Detail)
	default:
		return fmt.Errorf("unexpected response type: %T", resp)
	}
}

// Run executes the list relations command
func (r *RelationListCommand) Run() error {
	client, err := util.GetAuthenticatedClient(r.Config)
	if err != nil {
		return fmt.Errorf("failed to create authenticated client: %w", err)
	}

	// Build the parameters for the API call
	// Relations are retrieved through the GetEntities endpoint
	params := api.GetEntitiesParams{}

	// Set optional filters if provided
	if r.Label != "" {
		params.Label = api.NewOptString(r.Label)
	}
	if r.Limit > 0 {
		params.Limit = api.NewOptInt(r.Limit)
	}
	if r.Offset > 0 {
		params.Offset = api.NewOptInt(r.Offset)
	}

	// If source or target is specified, we need to use field selectors
	// However, for now we'll retrieve all relations and filter client-side
	// TODO: Add server-side filtering when API supports it

	resp, err := client.GetEntities(context.Background(), params)
	if err != nil {
		return fmt.Errorf("failed to list relations: %w", err)
	}

	// Handle the response
	switch result := resp.(type) {
	case *api.EntityResultSetResponse:
		relations := result.Relations
		if len(relations) == 0 {
			fmt.Println("No relations found.")
			return nil
		}

		// Filter relations if source or target is specified
		filteredRelations := filterRelations(relations, r.Source, r.Target)
		if len(filteredRelations) == 0 {
			fmt.Println("No relations found matching the specified filters.")
			return nil
		}

		return displayRelationList(filteredRelations, r.Output)
	case *api.GetEntitiesNotFound:
		fmt.Println("No relations found.")
		return nil
	default:
		return fmt.Errorf("unexpected response type: %T", resp)
	}
}

// Run executes the delete relation command
func (r *RelationDeleteCommand) Run() error {
	client, err := util.GetAuthenticatedClient(r.Config)
	if err != nil {
		return fmt.Errorf("failed to create authenticated client: %w", err)
	}

	// Parse source and target entity references
	source, err := parseEntityReference(r.Source)
	if err != nil {
		return fmt.Errorf("invalid source entity ID: %w", err)
	}

	target, err := parseEntityReference(r.Target)
	if err != nil {
		return fmt.Errorf("invalid target entity ID: %w", err)
	}

	// Build the relation request
	relation := api.EntityRelation{
		Relation: r.Relation,
		Source:   source,
		Target:   target,
	}

	// Determine namespace for the API call
	namespace := r.Namespace
	if namespace == "" {
		// Use the namespace from the source entity
		if ns, ok := source.Namespace.Get(); ok {
			namespace = ns
		}
	}

	if namespace != "" {
		relation.Namespace = api.NewOptString(namespace)
	}

	params := api.DeleteEntityRelationParams{
		Namespace: namespace,
	}

	// Delete the relation
	resp, err := client.DeleteEntityRelation(context.Background(), &relation, params)
	if err != nil {
		return fmt.Errorf("failed to delete relation: %w", err)
	}

	// Handle the response
	switch r := resp.(type) {
	case *api.DeleteEntityRelationNoContent:
		fmt.Printf("✅ Relation deleted successfully\n")
		return nil
	case *api.DeleteEntityRelationNotFound:
		return fmt.Errorf("relation not found")
	case *api.HTTPValidationError:
		return fmt.Errorf("validation error: %v", r.Detail)
	default:
		return fmt.Errorf("unexpected response type: %T", resp)
	}
}

// filterRelations filters relations by source and/or target entity ID
func filterRelations(relations []api.EntityRelationResponse, sourceFilter, targetFilter string) []api.EntityRelationResponse {
	if sourceFilter == "" && targetFilter == "" {
		return relations
	}

	var filtered []api.EntityRelationResponse
	for _, rel := range relations {
		matches := true
		if sourceFilter != "" && rel.Source.ID != sourceFilter {
			matches = false
		}
		if targetFilter != "" && rel.Target.ID != targetFilter {
			matches = false
		}
		if matches {
			filtered = append(filtered, rel)
		}
	}
	return filtered
}

// displayRelationList displays a list of relations in the specified format
func displayRelationList(relations []api.EntityRelationResponse, outputFormat string) error {
	// Filter relations to only show required fields
	filtered := make([]FilteredEntityRelation, len(relations))
	for i, rel := range relations {
		filtered[i] = filterEntityRelation(rel)
	}

	switch outputFormat {
	case "json":
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")
		return encoder.Encode(filtered)
	case "yaml":
		encoder := yaml.NewEncoder(os.Stdout)
		encoder.SetIndent(2)
		return encoder.Encode(filtered)
	default:
		// Table format
		return displayRelationTable(filtered)
	}
}

// displayRelationTable displays relations in a formatted table
func displayRelationTable(relations []FilteredEntityRelation) error {
	if len(relations) == 0 {
		fmt.Println("No relations found.")
		return nil
	}

	// Define colors
	headerColor := color.New(color.Bold, color.FgCyan)
	relationColor := color.New(color.FgYellow)

	// Print header
	headerColor.Printf("%-20s %-50s %-50s %s\n", "RELATION", "SOURCE", "TARGET", "NAMESPACE")
	fmt.Println(strings.Repeat("-", 140))

	// Print each relation
	for _, rel := range relations {
		namespace := ""
		if rel.Namespace != "" {
			namespace = rel.Namespace
		}
		relationColor.Printf("%-20s", rel.Relation)
		fmt.Printf(" %-50s %-50s %s\n", truncate(rel.Source, 50), truncate(rel.Target, 50), namespace)
	}

	fmt.Printf("\nTotal: %d relations\n", len(relations))
	return nil
}

// truncate truncates a string to the specified length
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
