package util

import (
	"context"
	"fmt"

	"github.com/arctir/devgraph-cli/pkg/config"
	api "github.com/arctir/go-devgraph/pkg/apis/devgraph/v1"
)

// NoEnvironmentError is returned when a user is not associated with any environments
// in their Devgraph account. This typically happens for new users or users who
// haven't been granted access to any environments.
type NoEnvironmentError struct{}

// Error returns the error message for NoEnvironmentError.
func (e *NoEnvironmentError) Error() string {
	return "you don't have access to any environments"
}

// IsWarning indicates this error should be displayed as a warning, not an error.
func (e *NoEnvironmentError) IsWarning() bool {
	return true
}

// GetEnvironments retrieves all environments accessible to the authenticated user.
// It returns a slice of EnvironmentResponse objects or an error if the request fails.
func GetEnvironments(config config.Config) (*[]api.EnvironmentResponse, error) {
	client, err := GetAuthenticatedClient(config)
	if err != nil {
		return nil, err
	}

	ctx := context.TODO()
	resp, err := client.GetEnvironments(ctx)
	if err != nil {
		return nil, err
	}

	// Check if response is successful
	switch r := resp.(type) {
	case *api.GetEnvironmentsOKApplicationJSON:
		envs := []api.EnvironmentResponse(*r)
		return &envs, nil
	default:
		return nil, fmt.Errorf("failed to fetch environments")
	}
}

// CheckEnvironment validates that an environment is set in user settings.
// Returns true if an environment is configured, false otherwise.
func CheckEnvironment(cfg *config.Config) (bool, error) {
	userConfig, err := config.LoadUserConfig()
	if err != nil {
		return false, fmt.Errorf("failed to load user config: %w", err)
	}

	environment := userConfig.Settings.DefaultEnvironment
	if environment == "" {
		return false, fmt.Errorf("no environment configured. Run 'dg auth login' or 'dg config set-context <name> --env <env>' to set an environment")
	}

	// Validate that the environment exists
	err = ValidateEnvironment(*cfg, environment)
	if err != nil {
		return false, fmt.Errorf("configured environment '%s' is invalid: %w. Run 'dg config set-context <name> --env <env>' to select a valid environment", environment, err)
	}

	return true, nil
}

// ValidateEnvironment checks if the given environment ID exists and is accessible
// to the authenticated user. Returns nil if the environment is valid, or an error
// describing why the environment cannot be accessed.
func ValidateEnvironment(config config.Config, environmentID string) error {
	_, err := ResolveEnvironmentUUID(config, environmentID)
	return err
}

// ResolveEnvironmentUUID resolves an environment name, slug, or UUID to its UUID.
// The environmentIdentifier can be any of:
//   - Environment UUID (exact match)
//   - Environment slug (exact match)
//   - Environment name (exact match)
//
// Returns the UUID of the matching environment, or an error if no match is found.
func ResolveEnvironmentUUID(config config.Config, environmentIdentifier string) (string, error) {
	envs, err := GetEnvironments(config)
	if err != nil {
		return "", fmt.Errorf("failed to get environments: %w", err)
	}

	if envs == nil || len(*envs) == 0 {
		return "", &NoEnvironmentError{}
	}

	for _, env := range *envs {
		if env.ID.String() == environmentIdentifier || env.Slug == environmentIdentifier || env.Name == environmentIdentifier {
			return env.ID.String(), nil
		}
	}

	return "", fmt.Errorf("environment '%s' not found. Available environments: %v", environmentIdentifier, getEnvironmentList(*envs))
}

// getEnvironmentList returns a list of environment names/slugs for error messages.
// It formats each environment as "Name (slug)" for user-friendly error reporting.
func getEnvironmentList(envs []api.EnvironmentResponse) []string {
	var names []string
	for _, env := range envs {
		names = append(names, fmt.Sprintf("%s (%s)", env.Name, env.Slug))
	}
	return names
}
