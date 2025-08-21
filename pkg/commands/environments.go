package commands

import (
	"context"
	"fmt"

	"github.com/arctir/devgraph-cli/pkg/auth"
	"github.com/arctir/devgraph-cli/pkg/config"
	"github.com/arctir/devgraph-cli/pkg/util"
	api "github.com/arctir/go-devgraph/pkg/apis/devgraph/v1"
	"github.com/google/uuid"
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

func displayEnvironments(envs *[]api.EnvironmentResponse) {
	if envs == nil || len(*envs) == 0 {
		fmt.Println("No environments found.")
		return
	}

	headers := []string{"ID", "Name", "Slug"}
	data := make([]map[string]any, len(*envs))
	for i, env := range *envs {
		data[i] = map[string]any{
			"Name": env.Name,
			"ID":   env.ID,
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

	if e.Environment == "" {
		return fmt.Errorf("no environment configured. Use 'devgraph env switch' to set an environment")
	}

	ctx := context.TODO()
	envUUID, err := uuid.Parse(e.Environment)
	if err != nil {
		return fmt.Errorf("invalid environment UUID: %w", err)
	}
	params := api.ListEnvironmentUsersParams{
		EnvironmentID: envUUID,
	}
	resp, err := client.ListEnvironmentUsers(ctx, params)
	if err != nil {
		return err
	}

	// Check if response is successful
	switch r := resp.(type) {
	case *api.ListEnvironmentUsersOKApplicationJSON:
		users := []api.EnvironmentUserResponse(*r)
		if len(users) == 0 {
			fmt.Println("No users found in this environment.")
			return nil
		}
		displayEnvironmentUsers(&users)
	default:
		return fmt.Errorf("failed to list environment users")
	}
	return nil
}

func (e *EnvironmentUserAddCommand) Run() error {
	client, err := util.GetAuthenticatedClient(e.Config)
	if err != nil {
		return err
	}

	if e.Environment == "" {
		return fmt.Errorf("no environment configured. Use 'devgraph env switch' to set an environment")
	}

	ctx := context.TODO()
	invite := api.EnvironmentUserInvite{
		EmailAddress: e.Email,
		Role:         api.NewOptString(e.Role),
	}

	envUUID, err := uuid.Parse(e.Environment)
	if err != nil {
		return fmt.Errorf("invalid environment UUID: %w", err)
	}
	params := api.InviteEnvironmentUserParams{
		EnvironmentID: envUUID,
	}
	resp, err := client.InviteEnvironmentUser(ctx, &invite, params)
	if err != nil {
		return err
	}

	// Check if response is successful
	switch resp.(type) {
	case *api.EnvironmentUserResponse:
		// Success
	default:
		return fmt.Errorf("failed to invite user")
	}

	fmt.Printf("Successfully invited %s to the environment with role %s\n", e.Email, e.Role)
	return nil
}

func (e *EnvironmentUserRemoveCommand) Run() error {
	client, err := util.GetAuthenticatedClient(e.Config)
	if err != nil {
		return err
	}

	if e.Environment == "" {
		return fmt.Errorf("no environment configured. Use 'devgraph env switch' to set an environment")
	}

	ctx := context.TODO()
	envUUID, err := uuid.Parse(e.Environment)
	if err != nil {
		return fmt.Errorf("invalid environment UUID: %w", err)
	}
	params := api.DeleteEnvironmentUserParams{
		EnvironmentID: envUUID,
		UserID:        e.UserID,
	}
	resp, err := client.DeleteEnvironmentUser(ctx, params)
	if err != nil {
		return err
	}

	// Check if response is successful
	switch resp.(type) {
	case *api.DeleteEnvironmentUserNoContent:
		// Success
	default:
		return fmt.Errorf("failed to remove user")
	}

	fmt.Printf("Successfully removed user %s from the environment\n", e.UserID)
	return nil
}

func displayEnvironmentUsers(users *[]api.EnvironmentUserResponse) {
	if users == nil || len(*users) == 0 {
		fmt.Println("No users found in this environment.")
		return
	}

	headers := []string{"ID", "Email", "Role", "Status"}
	data := make([]map[string]any, len(*users))
	for i, user := range *users {
		data[i] = map[string]any{
			"ID":     user.ID,
			"Email":  user.EmailAddress,
			"Role":   user.Role,
			"Status": user.Status,
		}
	}
	util.DisplaySimpleTable(data, headers)
}
