// Package main provides the command-line interface for Devgraph CLI.
// Devgraph CLI is a command-line tool for interacting with Devgraph.ai,
// providing AI-powered chat, authentication, and resource management capabilities.
package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/alecthomas/kong"
	"github.com/arctir/devgraph-cli/pkg/commands"
	admincommands "github.com/arctir/devgraph-cli/pkg/commands/admin"
	"github.com/arctir/devgraph-cli/pkg/config"
	"github.com/arctir/devgraph-cli/pkg/util"
)

// CLI represents the main command-line interface structure for Devgraph CLI.
// It defines all available commands and their subcommands using Kong command-line parser.
type CLI struct {
	// Chat provides interactive AI chat functionality
	Chat commands.Chat `kong:"cmd,help='Start an interactive chat with AI'"`
	// Auth handles authentication with Devgraph accounts
	Auth commands.AuthCommand `kong:"cmd,help='Manage authentication with your Devgraph account'"`
	// Setup runs the interactive configuration wizard
	Setup commands.SetupCommand `kong:"cmd,help='Run interactive configuration wizard'"`
	// Config manages CLI configuration settings
	Config commands.ConfigCommand `kong:"cmd,help='Manage configuration settings'"`
	// Token manages API tokens for Devgraph
	Token commands.TokenCommand `kong:"cmd,help='Manage opaque tokens for Devgraph'"`
	// Environment manages Devgraph environments
	Environment commands.EnvironmentCommand `kong:"cmd,name='env',help='Manage environments for Devgraph'"`
	// EntityDefinition manages entity definitions
	EntityDefinition commands.EntityDefinitionCommand `kong:"cmd,help='Manage entity definitions for Devgraph'"`
	// Entity manages entities within Devgraph
	Entity commands.EntityCommand `kong:"cmd,help='Manage entities for Devgraph'"`
	// MCP manages Model Context Protocol resources
	MCP commands.MCPCommand `kong:"cmd,help='Manage MCP resources for Devgraph'"`
	// ModelProvider manages AI model providers
	ModelProvider commands.ModelProviderCommand `kong:"cmd,name='modelprovider',help='Manage Model Provider resources for Devgraph'"`
	// Model manages AI models and configurations
	Model commands.ModelCommand `kong:"cmd,help='Manage Model resources for Devgraph'"`

	// Admin provides administrative commands (hidden from regular users)
	Admin admincommands.AdminCommand `kong:"cmd,hidden='',help='Admin commands for Devgraph'"`
}

// main is the entry point for the Devgraph CLI application.
// It initializes the Kong command parser, handles first-time setup,
// and executes the requested command.
func main() {

	cli := CLI{}

	// Parse command-line arguments using Kong
	ctx := kong.Parse(&cli,
		kong.Name("devgraph"),
		kong.Description("Turn chaos into clarity"),
		kong.UsageOnError(),
		kong.ConfigureHelp(kong.HelpOptions{
			Compact:             false,
			NoExpandSubcommands: true,
		}),
	)

	// Show first-time setup guidance for commands that need authentication
	// Skip for help, setup, and auth commands since they don't require full config
	if ctx.Command() != "help" && ctx.Command() != "setup" && !strings.HasPrefix(ctx.Command(), "auth") {
		if shouldShowFirstTimeSetup() {
			showFirstTimeSetupMessage()
			return // Don't proceed with the command
		}
	}

	// Execute the requested command
	err := ctx.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

// shouldShowFirstTimeSetup determines if the user needs to complete initial setup
func shouldShowFirstTimeSetup() bool {
	// Check if user has valid credentials
	if !util.IsAuthenticated() {
		return true // Need to authenticate first
	}

	// Check if this is truly first time (no config exists)
	return config.IsFirstTimeSetup()
}

// showFirstTimeSetupMessage displays helpful guidance for new users
func showFirstTimeSetupMessage() {
	fmt.Println("🆕 Welcome to Devgraph CLI!")
	fmt.Println()

	// Check if user has credentials
	if !util.IsAuthenticated() {
		fmt.Println("To get started, please authenticate:")
		fmt.Println("  devgraph auth login")
		fmt.Println()
		fmt.Println("After authentication, you can:")
		fmt.Println("  devgraph setup     # Configure your preferences")
		fmt.Println("  devgraph chat      # Start chatting with AI")
		fmt.Println("  devgraph --help    # See all available commands")
	} else {
		fmt.Println("You're authenticated! Complete the setup:")
		fmt.Println("  devgraph setup     # Configure your preferences")
		fmt.Println()
		fmt.Println("Or start using Devgraph right away:")
		fmt.Println("  devgraph chat      # Start chatting with AI")
		fmt.Println("  devgraph --help    # See all available commands")
	}
}
