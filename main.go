// Package main provides the command-line interface for Devgraph CLI.
// Devgraph CLI is a command-line tool for interacting with Devgraph.ai,
// providing AI-powered chat, authentication, and resource management capabilities.
package main

import (
	"errors"
	"fmt"
	"os"
	"reflect"
	"strings"

	"github.com/alecthomas/kong"
	"github.com/arctir/devgraph-cli/pkg/commands"
	"github.com/arctir/devgraph-cli/pkg/config"
	"github.com/arctir/devgraph-cli/pkg/util"
)

// CLI represents the main command-line interface structure for Devgraph CLI.
// It defines all available commands and their subcommands using Kong command-line parser.
type CLI struct {
	// Auth handles authentication with Devgraph accounts
	Auth commands.AuthCommand `kong:"cmd,help='Manage authentication with your Devgraph account'"`
	// Chat provides interactive AI chat functionality
	Chat commands.Chat `kong:"cmd,help='Start an interactive chat with AI'"`
	// Complete is a hidden command for dynamic shell completions
	Complete commands.CompleteCommand `kong:"cmd,hidden,help='Generate dynamic completions for resources'"`
	// Completion generates shell completion scripts
	Completion commands.CompletionCommand `kong:"cmd,help='Generate shell completion scripts'"`
	// Config manages CLI configuration settings
	Config commands.ConfigCommand `kong:"cmd,help='Manage configuration settings'"`
	// Entity manages entities within Devgraph
	Entity commands.EntityCommand `kong:"cmd,help='Manage entities for Devgraph'"`
	// EntityDefinition manages entity definitions
	EntityDefinition commands.EntityDefinitionCommand `kong:"cmd,help='Manage entity definitions for Devgraph'"`
	// Environment manages Devgraph environments
	Environment commands.EnvironmentCommand `kong:"cmd,name='env',help='Manage environments for Devgraph'"`
	// MCP manages Model Context Protocol resources
	MCP commands.MCPCommand `kong:"cmd,help='Manage MCP resources for Devgraph'"`
	// Model manages AI models and configurations
	Model commands.ModelCommand `kong:"cmd,help='Manage Model resources for Devgraph'"`
	// ModelProvider manages AI model providers
	ModelProvider commands.ModelProviderCommand `kong:"cmd,name='modelprovider',help='Manage Model Provider resources for Devgraph'"`
	// OAuthService manages OAuth service configurations
	OAuthService commands.OAuthServiceCommand `kong:"cmd,name='oauthservice',help='Manage OAuth services for Devgraph'"`
	// Provider manages discovery providers
	Provider commands.ProviderCommand `kong:"cmd,help='Manage discovery providers'"`
	// Subscription manages subscription information
	Subscription commands.SubscriptionCommand `kong:"cmd,help='Manage subscriptions'"`
	// Suggestion manages chat suggestions
	Suggestion commands.SuggestionCommand `kong:"cmd,help='Manage chat suggestions'"`
	// Token manages API tokens for Devgraph
	Token commands.TokenCommand `kong:"cmd,help='Manage opaque tokens for Devgraph'"`
	// User manages users in the current environment
	User commands.UserCommand `kong:"cmd,help='Manage users in the current environment'"`
}

// main is the entry point for the Devgraph CLI application.
// It initializes the Kong command parser, handles first-time setup,
// and executes the requested command.
func main() {

	cli := CLI{}

	// Parse command-line arguments using Kong
	ctx := kong.Parse(&cli,
		kong.Name("dg"),
		kong.Description("Turn chaos into clarity"),
		kong.UsageOnError(),
		kong.ConfigureHelp(kong.HelpOptions{
			Compact:             false,
			NoExpandSubcommands: true,
		}),
	)

	// Apply defaults to embedded Config structs after parsing
	if cmd := ctx.Selected(); cmd != nil {
		applyConfigDefaults(cmd.Target)
	}

	// Show first-time setup guidance for commands that need authentication
	// Skip for help, auth, completion, and complete commands since they don't require full config
	if ctx.Command() != "help" && ctx.Command() != "completion" && !strings.HasPrefix(ctx.Command(), "auth") && !strings.HasPrefix(ctx.Command(), "complete") {
		if shouldShowFirstTimeSetup() {
			showFirstTimeSetupMessage()
			return // Don't proceed with the command
		}
	}

	// Execute the requested command
	err := ctx.Run()
	if err != nil {
		// Check if this is a warning-type error
		var noEnvErr *util.NoEnvironmentError
		if errors.As(err, &noEnvErr) {
			fmt.Fprintf(os.Stderr, "⚠️  %v\n", err)
		} else {
			fmt.Fprintf(os.Stderr, "❌ %v\n", err)
		}
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
	fmt.Println("🆕 Welcome to Devgraph!")
	fmt.Println()
	fmt.Println("To get started, please authenticate:")
	fmt.Println("  dg auth login")
	fmt.Println()
	fmt.Println("For help:")
	fmt.Println("  dg --help")
}

// applyConfigDefaults walks the struct and applies defaults to any embedded config.Config
func applyConfigDefaults(target interface{}) {
	v := reflect.ValueOf(target)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		return
	}

	t := v.Type()
	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		fieldType := t.Field(i)

		// Check if this field is config.Config or embeds it
		if fieldType.Type == reflect.TypeOf(config.Config{}) && field.CanAddr() {
			if cfg, ok := field.Addr().Interface().(*config.Config); ok {
				cfg.ApplyDefaults()
			}
		} else if fieldType.Anonymous && fieldType.Type == reflect.TypeOf(config.Config{}) {
			// Handle anonymous/embedded Config
			if cfg, ok := field.Addr().Interface().(*config.Config); ok {
				cfg.ApplyDefaults()
			}
		}
	}
}
