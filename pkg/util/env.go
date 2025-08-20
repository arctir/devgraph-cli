package util

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/arctir/devgraph-cli/pkg/config"
	devgraphv1 "github.com/arctir/go-devgraph/pkg/apis/devgraph/v1"
)

// NoEnvironmentError is returned when a user is not associated with any environments
// in their Devgraph account. This typically happens for new users or users who
// haven't been granted access to any environments.
type NoEnvironmentError struct {
}

// Error returns the error message for NoEnvironmentError.
func (e *NoEnvironmentError) Error() string {
	return "User is not associated with any environments"
}

// GetEnvironments retrieves all environments accessible to the authenticated user.
// It returns a slice of EnvironmentResponse objects or an error if the request fails.
func GetEnvironments(config config.Config) (*[]devgraphv1.EnvironmentResponse, error) {
	client, err := GetAuthenticatedClient(config)
	if err != nil {
		return nil, err
	}

	ctx := context.TODO()
	resp, err := client.GetEnvironmentsWithResponse(ctx)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode() != 200 {
		return nil, fmt.Errorf("failed to fetch environments: status code %d", resp.StatusCode())
	}

	return resp.JSON200, nil
}

// CheckEnvironment validates and ensures an environment is set in the config.
// If no environment is set, it prompts the user to select from available environments.
// Returns true if an environment was successfully set or validated, false otherwise.
// This function may prompt the user for input if multiple environments are available.
func CheckEnvironment(config *config.Config) (bool, error) {
	if config.Environment != "" {
		// Validate that the environment exists on the current API server
		err := ValidateEnvironment(*config, config.Environment)
		if err != nil {
			fmt.Printf("Warning: Current environment '%s' is not valid for this API server. Clearing environment setting.\n", config.Environment)
			config.Environment = ""
			// Fall through to environment selection logic below
		} else {
			return true, nil
		}
	}

	if config.Environment == "" {
		envs, err := GetEnvironments(*config)
		if err != nil {
			return false, fmt.Errorf("failed to get environments: %w", err)
		}

		if envs == nil || len(*envs) == 0 {
			return false, &NoEnvironmentError{}
		}

		if len(*envs) == 1 {
			config.Environment = (*envs)[0].Id.String()
			fmt.Printf("Only one environment available. Environment set to: %s\n", config.Environment)
			return true, nil
		}

		fmt.Println("Environment not set. Available environments:")
		for i, env := range *envs {
			fmt.Printf("%d. %s - %s (ID: %s)\n", i+1, env.Name, env.Slug, env.Id)
		}

		reader := bufio.NewReader(os.Stdin)
		fmt.Print("Enter the number of your choice: ")
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)

		// Convert input to integer
		choice, err := strconv.Atoi(input)
		if err != nil || choice < 1 || choice > len(*envs) {
			fmt.Println("Invalid choice. Please enter a number between 1 and", len(*envs))
			return false, fmt.Errorf("invalid environment choice")
		}
		// Set the selected environment in the config
		config.Environment = (*envs)[choice-1].Id.String()
		fmt.Printf("Environment set to: %s\n", config.Environment)
		return true, nil
	}

	return false, nil
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
// Returns the UUID of the matching environment, or an error if no match is found.
func ResolveEnvironmentUUID(config config.Config, environmentIdentifier string) (string, error) {
	envs, err := GetEnvironments(config)
	if err != nil {
		return "", fmt.Errorf("failed to get environments: %w", err)
	}

	if envs == nil || len(*envs) == 0 {
		return "", fmt.Errorf("no environments available")
	}

	for _, env := range *envs {
		if env.Id.String() == environmentIdentifier || env.Slug == environmentIdentifier || env.Name == environmentIdentifier {
			return env.Id.String(), nil
		}
	}

	return "", fmt.Errorf("environment '%s' not found. Available environments: %v", environmentIdentifier, getEnvironmentList(*envs))
}

// getEnvironmentList returns a list of environment names/slugs for error messages.
// It formats each environment as "Name (slug)" for user-friendly error reporting.
func getEnvironmentList(envs []devgraphv1.EnvironmentResponse) []string {
	var names []string
	for _, env := range envs {
		names = append(names, fmt.Sprintf("%s (%s)", env.Name, env.Slug))
	}
	return names
}
