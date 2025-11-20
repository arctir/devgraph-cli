package commands

import (
	"context"
	"fmt"

	"github.com/arctir/devgraph-cli/pkg/config"
	"github.com/arctir/devgraph-cli/pkg/util"
	api "github.com/arctir/go-devgraph/pkg/apis/devgraph/v1"
	"github.com/google/uuid"
)

// getDefaultEnvironment returns the default environment UUID from user settings
func getDefaultEnvironment() (string, error) {
	userConfig, err := config.LoadUserConfig()
	if err != nil {
		return "", fmt.Errorf("failed to load user config: %w", err)
	}
	if userConfig.Settings.DefaultEnvironment == "" {
		return "", fmt.Errorf("no environment configured. Use 'dg config set-context <name> --env <env>' to set an environment")
	}
	return userConfig.Settings.DefaultEnvironment, nil
}

type EnvironmentListCommand struct {
	config.Config
}

type EnvironmentUserListCommand struct {
	EnvWrapperCommand
	Invited bool `short:"i" help:"Show only pending invitations"`
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

type EnvironmentCurrentCommand struct{}

type EnvironmentCommand struct {
	Current EnvironmentCurrentCommand `cmd:"current" help:"Display the current environment"`
	List    EnvironmentListCommand    `cmd:"list" help:"List all environments for Devgraph"`
}

// UserCommand manages users in the current environment
type UserCommand struct {
	List   EnvironmentUserListCommand   `cmd:"list" help:"List users in the current environment"`
	Add    EnvironmentUserAddCommand    `cmd:"add" help:"Invite a user to the current environment"`
	Remove EnvironmentUserRemoveCommand `cmd:"remove" help:"Remove a user from the current environment"`
}

func (e *EnvironmentCurrentCommand) Run() error {
	userConfig, err := config.LoadUserConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if userConfig.CurrentContext == "" {
		return fmt.Errorf("no current context set")
	}

	context, ok := userConfig.Contexts[userConfig.CurrentContext]
	if !ok {
		return fmt.Errorf("context '%s' not found", userConfig.CurrentContext)
	}

	if context.Environment == "" {
		return fmt.Errorf("no environment set for context '%s'. Use 'dg config set-context %s --env <env>' to set an environment", userConfig.CurrentContext, userConfig.CurrentContext)
	}

	fmt.Println(context.Environment)
	return nil
}

func (e *EnvironmentListCommand) Run() error {
	e.Config.ApplyDefaults()
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

	environment, err := getDefaultEnvironment()
	if err != nil {
		return err
	}

	ctx := context.TODO()
	envUUID, err := uuid.Parse(environment)
	if err != nil {
		return fmt.Errorf("invalid environment UUID: %w", err)
	}

	if e.Invited {
		params := api.GetPendingInvitationsParams{
			EnvironmentID: envUUID,
		}
		resp, err := client.GetPendingInvitations(ctx, params)
		if err != nil {
			return err
		}

		switch r := resp.(type) {
		case *api.GetPendingInvitationsOKApplicationJSON:
			invites := []api.PendingInvitationResponse(*r)
			if len(invites) == 0 {
				fmt.Println("No pending invitations found in this environment.")
				return nil
			}
			displayPendingInvitations(&invites)
		default:
			return fmt.Errorf("failed to list pending invitations")
		}
		return nil
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

	environment, err := getDefaultEnvironment()
	if err != nil {
		return err
	}

	ctx := context.TODO()
	invite := api.EnvironmentUserInvite{
		EmailAddress: e.Email,
		Role:         api.NewOptEnvironmentUserInviteRole(api.EnvironmentUserInviteRole(e.Role)),
	}

	envUUID, err := uuid.Parse(environment)
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

	fmt.Printf("✅ Invited '%s' to environment with role '%s'.\n", e.Email, e.Role)
	return nil
}

func (e *EnvironmentUserRemoveCommand) Run() error {
	client, err := util.GetAuthenticatedClient(e.Config)
	if err != nil {
		return err
	}

	environment, err := getDefaultEnvironment()
	if err != nil {
		return err
	}

	ctx := context.TODO()
	envUUID, err := uuid.Parse(environment)
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

	fmt.Printf("✅ Removed user '%s' from environment.\n", e.UserID)
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

func displayPendingInvitations(invites *[]api.PendingInvitationResponse) {
	if invites == nil || len(*invites) == 0 {
		fmt.Println("No pending invitations found in this environment.")
		return
	}

	headers := []string{"ID", "Email", "Role", "Status"}
	data := make([]map[string]any, len(*invites))
	for i, invite := range *invites {
		data[i] = map[string]any{
			"ID":     invite.ID,
			"Email":  invite.EmailAddress,
			"Role":   invite.Role,
			"Status": invite.Status,
		}
	}
	util.DisplaySimpleTable(data, headers)
}
