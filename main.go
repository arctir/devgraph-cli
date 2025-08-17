package main

import (
	"fmt"
	"os"

	"github.com/alecthomas/kong"
	"github.com/arctir/devgraph-cli/pkg/commands"
	admincommands "github.com/arctir/devgraph-cli/pkg/commands/admin"
)

type CLI struct {
	Chat             commands.Chat                    `kong:"cmd,help='Start an interactive chat with AI'"`
	Auth             commands.Auth                    `kong:"cmd,help='Authenticate your Devgraph client'"`
	Setup            commands.SetupCommand            `kong:"cmd,help='Run interactive configuration wizard'"`
	Config           commands.ConfigCommand           `kong:"cmd,help='Manage configuration settings'"`
	Token            commands.TokenCommand            `kong:"cmd,help='Manage opaque tokens for Devgraph'"`
	Environment      commands.EnvironmentCommand      `kong:"cmd,name='env',help='Manage environments for Devgraph'"`
	EntityDefinition commands.EntityDefinitionCommand `kong:"cmd,help='Manage entity definitions for Devgraph'"`
	Entity           commands.EntityCommand           `kong:"cmd,help='Manage entities for Devgraph'"`
	MCP              commands.MCPCommand              `kong:"cmd,help='Manage MCP resources for Devgraph'"`
	ModelProvider    commands.ModelProviderCommand    `kong:"cmd,name='modelprovider',help='Manage Model Provider resources for Devgraph'"`
	Model            commands.ModelCommand            `kong:"cmd,help='Manage Model resources for Devgraph'"`

	Admin admincommands.AdminCommand `kong:"cmd,hidden='',help='Admin commands for Devgraph'"`
}

func main() {

	cli := CLI{}

	ctx := kong.Parse(&cli,
		kong.Name("devgraph"),
		kong.Description("Turn chaos into clarity"),
		//kong.UsageOnError(),
		kong.ConfigureHelp(kong.HelpOptions{
			Compact:             true,
			NoExpandSubcommands: true,
		}),
	)

	// Check if this is first-time setup and run wizard if needed
	// Skip wizard for help commands and setup command itself
	if ctx.Command() != "help" && ctx.Command() != "setup" && ctx.Command() != "auth" {
		err := commands.RunConfigurationWizard()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Configuration wizard error: %v\n", err)
			os.Exit(1)
		}
	}

	err := ctx.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

}
