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
	Output string `short:"o" help:"Output format: table, json, yaml" default:"table"`
}

type EnvironmentUserListCommand struct {
	EnvWrapperCommand
	Invited bool   `short:"i" help:"Show only pending invitations"`
	Output  string `short:"o" help:"Output format: table, json, yaml" default:"table"`
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

type EnvironmentDeleteCommand struct {
	EnvWrapperCommand
	EnvironmentID string `arg:"" required:"" help:"Environment ID to delete"`
	Confirm       bool   `short:"y" help:"Skip confirmation prompt"`
}

type EnvironmentCommand struct {
	Current EnvironmentCurrentCommand `cmd:"current" help:"Display the current environment"`
	List    EnvironmentListCommand    `cmd:"list" help:"List all environments for Devgraph"`
	Delete  EnvironmentDeleteCommand  `cmd:"delete" help:"Delete an environment (WARNING: May be permanent after grace period)"`
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

	// Build structured data for json/yaml output
	type envOutput struct {
		ID   string `json:"id" yaml:"id"`
		Name string `json:"name" yaml:"name"`
		Slug string `json:"slug" yaml:"slug"`
	}

	structured := make([]envOutput, len(*envs))
	tableData := make([]map[string]any, len(*envs))
	for i, env := range *envs {
		structured[i] = envOutput{
			ID:   env.ID.String(),
			Name: env.Name,
			Slug: env.Slug,
		}
		tableData[i] = map[string]any{
			"ID":   env.ID.String(),
			"Name": env.Name,
			"Slug": env.Slug,
		}
	}

	headers := []string{"ID", "Name", "Slug"}
	return util.FormatOutput(e.Output, structured, headers, tableData)
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

			type inviteOutput struct {
				ID     string `json:"id" yaml:"id"`
				Email  string `json:"email" yaml:"email"`
				Role   string `json:"role" yaml:"role"`
				Status string `json:"status" yaml:"status"`
			}

			structured := make([]inviteOutput, len(invites))
			tableData := make([]map[string]any, len(invites))
			for i, invite := range invites {
				structured[i] = inviteOutput{
					ID:     invite.ID,
					Email:  invite.EmailAddress,
					Role:   string(invite.Role),
					Status: string(invite.Status),
				}
				tableData[i] = map[string]any{
					"ID":     invite.ID,
					"Email":  invite.EmailAddress,
					"Role":   invite.Role,
					"Status": invite.Status,
				}
			}

			headers := []string{"ID", "Email", "Role", "Status"}
			return util.FormatOutput(e.Output, structured, headers, tableData)
		default:
			return fmt.Errorf("failed to list pending invitations")
		}
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

		type userOutput struct {
			ID     string `json:"id" yaml:"id"`
			Email  string `json:"email" yaml:"email"`
			Role   string `json:"role" yaml:"role"`
			Status string `json:"status" yaml:"status"`
		}

		structured := make([]userOutput, len(users))
		tableData := make([]map[string]any, len(users))
		for i, user := range users {
			structured[i] = userOutput{
				ID:     user.ID,
				Email:  user.EmailAddress,
				Role:   string(user.Role),
				Status: string(user.Status),
			}
			tableData[i] = map[string]any{
				"ID":     user.ID,
				"Email":  user.EmailAddress,
				"Role":   user.Role,
				"Status": user.Status,
			}
		}

		headers := []string{"ID", "Email", "Role", "Status"}
		return util.FormatOutput(e.Output, structured, headers, tableData)
	default:
		return fmt.Errorf("failed to list environment users")
	}
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

func (e *EnvironmentDeleteCommand) Run() error {
	client, err := util.GetAuthenticatedClient(e.Config)
	if err != nil {
		return err
	}

	ctx := context.TODO()
	envUUID, err := uuid.Parse(e.EnvironmentID)
	if err != nil {
		return fmt.Errorf("invalid environment UUID: %w", err)
	}

	// Display BIG WARNING
	fmt.Println("╔════════════════════════════════════════════════════════════════════════════╗")
	fmt.Println("║                            ⚠️  WARNING ⚠️                                    ║")
	fmt.Println("║                                                                            ║")
	fmt.Println("║  You are about to DELETE an environment. This action will:                ║")
	fmt.Println("║                                                                            ║")
	fmt.Println("║  1. Mark the environment as DELETED immediately                           ║")
	fmt.Println("║  2. Enter a GRACE PERIOD (typically 30 days based on your plan)           ║")
	fmt.Println("║  3. After the grace period, deletion becomes PERMANENT and IRREVERSIBLE   ║")
	fmt.Println("║                                                                            ║")
	fmt.Println("║  During the grace period:                                                 ║")
	fmt.Println("║  • The environment will be INACCESSIBLE                                    ║")
	fmt.Println("║  • All data is retained but UNUSABLE                                       ║")
	fmt.Println("║  • Kubernetes resources will be CLEANED UP                                 ║")
	fmt.Println("║  • You may contact support to recover the environment                     ║")
	fmt.Println("║                                                                            ║")
	fmt.Println("║  After the grace period expires:                                          ║")
	fmt.Println("║  • ALL DATA WILL BE PERMANENTLY DELETED                                    ║")
	fmt.Println("║  • RECOVERY WILL NOT BE POSSIBLE                                           ║")
	fmt.Println("║                                                                            ║")
	fmt.Println("╚════════════════════════════════════════════════════════════════════════════╝")
	fmt.Println()

	// Prompt for confirmation unless -y flag is used
	if !e.Confirm {
		fmt.Printf("Environment ID: %s\n\n", e.EnvironmentID)
		fmt.Print("Type 'DELETE' (all caps) to confirm deletion: ")

		var confirmation string
		_, err := fmt.Scanln(&confirmation)
		if err != nil {
			return fmt.Errorf("failed to read confirmation: %w", err)
		}

		if confirmation != "DELETE" {
			fmt.Println("❌ Deletion cancelled.")
			return nil
		}
	}

	// Execute the delete
	params := api.DeleteEnvironmentParams{
		EnvID: envUUID,
	}
	resp, err := client.DeleteEnvironment(ctx, params)
	if err != nil {
		return fmt.Errorf("failed to delete environment: %w", err)
	}

	// Check if response is successful
	switch resp.(type) {
	case *api.DeleteEnvironmentNoContent:
		fmt.Println()
		fmt.Println("✅ Environment has been marked for deletion.")
		fmt.Println()
		fmt.Println("The environment is now entering the grace period.")
		fmt.Println("Contact support if you need to recover this environment before the grace period expires.")
		return nil
	case *api.DeleteEnvironmentNotFound:
		return fmt.Errorf("environment not found")
	default:
		return fmt.Errorf("unexpected response when deleting environment")
	}
}
