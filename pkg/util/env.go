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

func CheckEnvironment(config *config.Config) (bool, error) {
	if config.Environment != "" {
		return true, nil
	}

	if config.Environment == "" {
		envs, err := GetEnvironments(*config)
		if err != nil {
			return false, fmt.Errorf("failed to get environments: %w", err)
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
