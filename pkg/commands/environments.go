package commands

import (
	"context"
	"fmt"

	"github.com/arctir/devgraph-cli/pkg/auth"
	"github.com/arctir/devgraph-cli/pkg/config"
	"github.com/arctir/devgraph-cli/pkg/util"
	devgraphv1 "github.com/arctir/go-devgraph/pkg/apis/devgraph/v1"
)

type EnvironmentListCommand struct {
	config.Config
}

type EnvironmentSwitchCommand struct {
	config.Config
	Environment string `arg:"" required:"" help:"Environment name, slug, or ID to switch to"`
}

type EnvironmentUserListCommand struct {
	EnvWrapperCommand
}

type EnvironmentUserAddCommand struct {
	EnvWrapperCommand
	Email string `arg:"" required:"" help:"Email address of user to invite"`
	Role  string `short:"r" help:"Role for the user" default:"member"`
}

type EnvironmentUserRemoveCommand struct {
	EnvWrapperCommand
	UserID string `arg:"" required:"" help:"User ID to remove"`
}

type EnvironmentUserCommand struct {
	List   EnvironmentUserListCommand   `cmd:"list" help:"List users in the current environment"`
	Add    EnvironmentUserAddCommand    `cmd:"add" help:"Invite a user to the current environment"`
	Remove EnvironmentUserRemoveCommand `cmd:"remove" help:"Remove a user from the current environment"`
}

type EnvironmentCommand struct {
	List   EnvironmentListCommand   `cmd:"list" help:"List all environments for Devgraph"`
	Switch EnvironmentSwitchCommand `cmd:"switch" help:"Switch to a different environment"`
	User   EnvironmentUserCommand   `cmd:"user" help:"Manage users in the current environment"`
}

func (e *EnvironmentListCommand) Run() error {
	envs, err := util.GetEnvironments(e.Config)
	if err != nil {
		return err
	}

	if envs == nil || len(*envs) == 0 {
		fmt.Println("No environments found.")
		return nil
	}

	displayEnvironments(envs)
	return nil
}

func (e *EnvironmentSwitchCommand) Run() error {
	// Resolve the environment identifier to UUID
	environmentUUID, err := util.ResolveEnvironmentUUID(e.Config, e.Environment)
	if err != nil {
		return fmt.Errorf("failed to resolve environment '%s': %w", e.Environment, err)
	}

	// Load user config
	userConfig, err := config.LoadUserConfig()
	if err != nil {
		return fmt.Errorf("failed to load user config: %w", err)
	}

	// Update default environment
	userConfig.Settings.DefaultEnvironment = environmentUUID

	// Save config
	err = config.SaveUserConfig(userConfig)
	if err != nil {
		return fmt.Errorf("failed to save user config: %w", err)
	}

	// Optionally update Clerk session for the new environment
	// This is currently a no-op but could be extended for Clerk organization switching
	err = auth.UpdateClerkSessionForEnvironment(e.Config, environmentUUID)
	if err != nil {
		fmt.Printf("Warning: Failed to update Clerk session: %v\n", err)
	}

	fmt.Printf("Switched to environment: %s (UUID: %s)\n", e.Environment, environmentUUID)
	return nil
}

func displayEnvironments(envs *[]devgraphv1.EnvironmentResponse) {
	if envs == nil || len(*envs) == 0 {
		fmt.Println("No environments found.")
		return
	}

	headers := []string{"ID", "Name", "Slug"}
	data := make([]map[string]any, len(*envs))
	for i, env := range *envs {
		data[i] = map[string]any{
			"Name": env.Name,
			"ID":   env.Id,
			"Slug": env.Slug,
		}
	}
	util.DisplaySimpleTable(data, headers)
}

func (e *EnvironmentUserListCommand) Run() error {
	client, err := util.GetAuthenticatedClient(e.Config)
	if err != nil {
		return err
	}

	if e.Config.Environment == "" {
		return fmt.Errorf("no environment configured. Use 'devgraph env switch' to set an environment")
	}

	ctx := context.TODO()
	resp, err := client.ListEnvironmentUsersWithResponse(ctx, e.Config.Environment)
	if err != nil {
		return err
	}

	if resp.StatusCode() != 200 {
		return fmt.Errorf("failed to fetch environment users: status code %d", resp.StatusCode())
	}

	users := resp.JSON200
	if users == nil || len(*users) == 0 {
		fmt.Println("No users found in this environment.")
		return nil
	}

	displayEnvironmentUsers(users)
	return nil
}

func (e *EnvironmentUserAddCommand) Run() error {
	client, err := util.GetAuthenticatedClient(e.Config)
	if err != nil {
		return err
	}

	if e.Config.Environment == "" {
		return fmt.Errorf("no environment configured. Use 'devgraph env switch' to set an environment")
	}

	ctx := context.TODO()
	invite := devgraphv1.EnvironmentUserInvite{
		EmailAddress: e.Email,
		Role:         &e.Role,
	}

	resp, err := client.InviteEnvironmentUserWithResponse(ctx, e.Config.Environment, invite)
	if err != nil {
		return err
	}

	if resp.StatusCode() != 200 && resp.StatusCode() != 201 {
		return fmt.Errorf("failed to invite user: status code %d", resp.StatusCode())
	}

	fmt.Printf("Successfully invited %s to the environment with role %s\n", e.Email, e.Role)
	return nil
}

func (e *EnvironmentUserRemoveCommand) Run() error {
	client, err := util.GetAuthenticatedClient(e.Config)
	if err != nil {
		return err
	}

	if e.Config.Environment == "" {
		return fmt.Errorf("no environment configured. Use 'devgraph env switch' to set an environment")
	}

	ctx := context.TODO()
	resp, err := client.DeleteEnvironmentUserWithResponse(ctx, e.Config.Environment, e.UserID)
	if err != nil {
		return err
	}

	if resp.StatusCode() != 200 && resp.StatusCode() != 204 {
		return fmt.Errorf("failed to remove user: status code %d", resp.StatusCode())
	}

	fmt.Printf("Successfully removed user %s from the environment\n", e.UserID)
	return nil
}

func displayEnvironmentUsers(users *[]devgraphv1.EnvironmentUserResponse) {
	if users == nil || len(*users) == 0 {
		fmt.Println("No users found in this environment.")
		return
	}

	headers := []string{"ID", "Email", "Role", "Status"}
	data := make([]map[string]any, len(*users))
	for i, user := range *users {
		data[i] = map[string]any{
			"ID":     user.Id,
			"Email":  user.EmailAddress,
			"Role":   user.Role,
			"Status": user.Status,
		}
	}
	util.DisplaySimpleTable(data, headers)
}
