package commands

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/arctir/devgraph-cli/pkg/auth"
	"github.com/arctir/devgraph-cli/pkg/config"
	"github.com/arctir/devgraph-cli/pkg/util"
)

type SetupCommand struct {
	config.Config
}

func (s *SetupCommand) Run() error {
	fmt.Println("🚀 Welcome to Devgraph CLI Setup!")
	fmt.Println("Let's configure your development environment.")
	fmt.Println()

	// Check if already authenticated - require it before setup
	creds, err := auth.LoadCredentials()
	if err != nil || creds.IDToken == "" || creds.AccessToken == "" {
		fmt.Println("❌ You must be authenticated to run setup.")
		fmt.Println("Please run 'devgraph auth login' first to authenticate.")
		return fmt.Errorf("authentication required")
	}

	// Check if tokens are expired
	if creds.Claims != nil {
		if exp, ok := (*creds.Claims)["exp"].(float64); ok {
			if time.Now().Unix() > int64(exp) {
				fmt.Println("❌ Your authentication token has expired.")
				fmt.Println("Please run 'devgraph auth login' to re-authenticate.")
				return fmt.Errorf("authentication token expired")
			}
		}
	}

	fmt.Println("✅ Authenticated with Devgraph.")
	fmt.Println()

	// Load or create user config
	userConfig, err := config.LoadUserConfig()
	if err != nil {
		userConfig = &config.UserConfig{}
	}

	// Configure environment
	err = s.configureEnvironment(userConfig)
	if err != nil {
		return err
	}

	// Configure model
	err = s.configureModel(userConfig)
	if err != nil {
		return err
	}

	// Configure max tokens
	err = s.configureMaxTokens(userConfig)
	if err != nil {
		return err
	}

	// Save configuration
	err = config.SaveUserConfig(userConfig)
	if err != nil {
		return fmt.Errorf("failed to save configuration: %w", err)
	}

	fmt.Println("\n🎉 Setup complete! Your configuration has been saved.")
	fmt.Println("You can modify these settings anytime with: devgraph config set")

	return nil
}

func (s *SetupCommand) configureEnvironment(userConfig *config.UserConfig) error {
	fmt.Println("🌍 Setting up your environment...")

	// Get available environments
	envs, err := util.GetEnvironments(s.Config)
	if err != nil {
		return fmt.Errorf("failed to get environments: %w", err)
	}

	if envs == nil || len(*envs) == 0 {
		fmt.Println("⚠️  No environments found. You may need to create one first.")
		return nil
	}

	if len(*envs) == 1 {
		env := (*envs)[0]
		userConfig.Settings.DefaultEnvironment = env.Id.String()
		fmt.Printf("✅ Only one environment available: %s\n", env.Name)
		return nil
	}

	fmt.Println("Available environments:")
	for i, env := range *envs {
		fmt.Printf("  %d. %s (%s)\n", i+1, env.Name, env.Slug)
	}

	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("\nSelect an environment (enter number): ")
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)

		choice, err := strconv.Atoi(input)
		if err != nil || choice < 1 || choice > len(*envs) {
			fmt.Printf("Invalid choice. Please enter a number between 1 and %d.\n", len(*envs))
			continue
		}

		selectedEnv := (*envs)[choice-1]
		userConfig.Settings.DefaultEnvironment = selectedEnv.Id.String()
		fmt.Printf("✅ Environment set to: %s\n", selectedEnv.Name)
		break
	}

	return nil
}

func (s *SetupCommand) configureModel(userConfig *config.UserConfig) error {
	fmt.Println("\n🤖 Setting up your AI model...")

	// Update config with selected environment for model API calls
	s.Environment = userConfig.Settings.DefaultEnvironment

	// Get available models
	models, err := util.GetModels(s.Config)
	if err != nil {
		fmt.Printf("⚠️  Could not fetch available models: %v\n", err)
		fmt.Println("Using default model: gpt-4o-mini")
		userConfig.Settings.DefaultModel = "gpt-4o-mini"
		return nil
	}

	if models == nil || len(*models) == 0 {
		fmt.Println("⚠️  No models found. Using default: gpt-4o-mini")
		userConfig.Settings.DefaultModel = "gpt-4o-mini"
		return nil
	}

	if len(*models) == 1 {
		model := (*models)[0]
		userConfig.Settings.DefaultModel = model.Name
		fmt.Printf("✅ Only one model available: %s\n", model.Name)
		return nil
	}

	fmt.Println("Available models:")
	for i, model := range *models {
		fmt.Printf("  %d. %s\n", i+1, model.Name)
	}

	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("\nSelect a model (enter number): ")
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)

		choice, err := strconv.Atoi(input)
		if err != nil || choice < 1 || choice > len(*models) {
			fmt.Printf("Invalid choice. Please enter a number between 1 and %d.\n", len(*models))
			continue
		}

		selectedModel := (*models)[choice-1]
		userConfig.Settings.DefaultModel = selectedModel.Name
		fmt.Printf("✅ Model set to: %s\n", selectedModel.Name)
		break
	}

	return nil
}

func (s *SetupCommand) configureMaxTokens(userConfig *config.UserConfig) error {
	fmt.Println("\n⚙️  Setting up token limits...")

	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("Enter maximum tokens for responses (1-100000, or press Enter for default 1000): ")
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)

		if input == "" {
			userConfig.Settings.DefaultMaxTokens = 1000
			fmt.Println("✅ Using default: 1000 tokens")
			break
		}

		tokens, err := strconv.Atoi(input)
		if err != nil || tokens < 1 || tokens > 100000 {
			fmt.Println("Invalid input. Please enter a number between 1 and 100000.")
			continue
		}

		userConfig.Settings.DefaultMaxTokens = tokens
		fmt.Printf("✅ Max tokens set to: %d\n", tokens)
		break
	}

	return nil
}

// RunConfigurationWizard runs the setup wizard if this is a first-time setup
func RunConfigurationWizard() error {
	// Check if credentials exist - if not, this is definitely first time
	_, credErr := auth.LoadCredentials()
	hasCredentials := credErr == nil

	// If credentials exist, check other first-time indicators
	if hasCredentials && !config.IsFirstTimeSetup() {
		return nil // Not first time, skip
	}

	fmt.Println("🆕 First time using Devgraph CLI!")

	if !hasCredentials {
		fmt.Println("To get started:")
		fmt.Println("1. First authenticate: devgraph auth login")
		fmt.Println("2. Then run setup: devgraph setup")
		return nil
	}

	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Would you like to run the configuration wizard? (Y/n): ")
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(strings.ToLower(input))

	if input == "n" || input == "no" {
		fmt.Println("Skipping configuration wizard. You can run it later with: devgraph setup")
		return nil
	}

	setupCmd := &SetupCommand{
		Config: config.Config{
			ApiURL:      "https://api.staging.devgraph.ai",
			IssuerURL:   config.DefaultIssuerURL,
			ClientID:    config.DefaultClientID,
			RedirectURL: config.DefaultRedirectURL,
		},
	}

	return setupCmd.Run()
}
