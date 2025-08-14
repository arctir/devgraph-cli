package main

import (
	"fmt"
	"os"

	"github.com/alecthomas/kong"
	"github.com/arctir/devgraph-cli/pkg/commands"
	admincommands "github.com/arctir/devgraph-cli/pkg/commands/admin"
)

type CLI struct {
	Chat          commands.Chat                 `kong:"cmd,help='Start an interactive chat with AI'"`
	Auth          commands.Auth                 `kong:"cmd,help='Authenticate your Devgraph client'"`
	Token         commands.TokenCommand         `kong:"cmd,help='Manage opaque tokens for Devgraph'"`
	Environment   commands.EnvironmentCommand   `kong:"cmd,name='env',help='Manage environments for Devgraph'"`
	MCP           commands.MCPCommand           `kong:"cmd,help='Manage MCP resources for Devgraph'"`
	ModelProvider commands.ModelProviderCommand `kong:"cmd,name='modelprovider',help='Manage Model Provider resources for Devgraph'"`
	Model         commands.ModelCommand         `kong:"cmd,help='Manage Model resources for Devgraph'"`

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

	err := ctx.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

}
