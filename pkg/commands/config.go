package commands

import (
	"fmt"

	"github.com/arctir/devgraph-cli/pkg/config"
	"github.com/arctir/devgraph-cli/pkg/util"
)

type ConfigCommand struct {
	Set  ConfigSetCommand  `kong:"cmd,help='Set configuration values'"`
	Get  ConfigGetCommand  `kong:"cmd,help='Get configuration values'"`
	Show ConfigShowCommand `kong:"cmd,help='Show current configuration'"`
}

type ConfigSetCommand struct {
	Environment string `kong:"help='Set default environment'"`
	Model       string `kong:"help='Set default chat model'"`
	MaxTokens   int    `kong:"help='Set default max tokens'"`
}

type ConfigGetCommand struct {
	Key string `kong:"arg,help='Configuration key to get (environment|model|max_tokens)'"`
}

type ConfigShowCommand struct {
}

func (c *ConfigSetCommand) Run() error {
	userConfig, err := config.LoadUserConfig()
	if err != nil {
		return fmt.Errorf("failed to load user config: %w", err)
	}

	// Create a temporary config for API validation
	apiConfig := config.Config{
		ApiURL:      "https://api.staging.devgraph.ai",
		IssuerURL:   config.DefaultIssuerURL,
		ClientID:    config.DefaultClientID,
		RedirectURL: config.DefaultRedirectURL,
		Environment: userConfig.Settings.DefaultEnvironment, // Use existing environment for validation
	}

	// Validate environment if provided
	if c.Environment != "" {
		environmentUUID, err := util.ResolveEnvironmentUUID(apiConfig, c.Environment)
		if err != nil {
			return fmt.Errorf("invalid environment: %w\nHint: Use 'devgraph env list' to see available environments", err)
		}
		userConfig.Settings.DefaultEnvironment = environmentUUID
		fmt.Printf("Set default environment to: %s (UUID: %s)\n", c.Environment, environmentUUID)
	}

	// Validate model if provided
	if c.Model != "" {
		// Model validation requires an environment to be set
		if apiConfig.Environment == "" && c.Environment == "" {
			// Skip validation if no environment is available
			fmt.Println("Warning: Skipping model validation because no environment is set. Set an environment first for model validation.")
		} else {
			// Update environment if being set in this command
			if c.Environment != "" {
				environmentUUID, _ := util.ResolveEnvironmentUUID(apiConfig, c.Environment)
				apiConfig.Environment = environmentUUID
			}

			err := util.ValidateModel(apiConfig, c.Model)
			if err != nil {
				return fmt.Errorf("invalid model: %w\nHint: Use 'devgraph model list' to see available models", err)
			}
		}
		userConfig.Settings.DefaultModel = c.Model
		fmt.Printf("Set default model to: %s\n", c.Model)
	}

	if c.MaxTokens > 0 {
		if c.MaxTokens < 1 || c.MaxTokens > 100000 {
			return fmt.Errorf("invalid max tokens: must be between 1 and 100000")
		}
		userConfig.Settings.DefaultMaxTokens = c.MaxTokens
		fmt.Printf("Set default max tokens to: %d\n", c.MaxTokens)
	}

	err = config.SaveUserConfig(userConfig)
	if err != nil {
		return fmt.Errorf("failed to save user config: %w", err)
	}

	return nil
}

func (c *ConfigGetCommand) Run() error {
	userConfig, err := config.LoadUserConfig()
	if err != nil {
		return fmt.Errorf("failed to load user config: %w", err)
	}

	switch c.Key {
	case "environment":
		if userConfig.Settings.DefaultEnvironment != "" {
			fmt.Println(userConfig.Settings.DefaultEnvironment)
		} else {
			fmt.Println("(not set)")
		}
	case "model":
		if userConfig.Settings.DefaultModel != "" {
			fmt.Println(userConfig.Settings.DefaultModel)
		} else {
			fmt.Println("gpt-4o-mini (default)")
		}
	case "max_tokens":
		if userConfig.Settings.DefaultMaxTokens > 0 {
			fmt.Println(userConfig.Settings.DefaultMaxTokens)
		} else {
			fmt.Println("1000 (default)")
		}
	default:
		return fmt.Errorf("unknown configuration key: %s. Valid keys are: environment, model, max_tokens", c.Key)
	}

	return nil
}

func (c *ConfigShowCommand) Run() error {
	userConfig, err := config.LoadUserConfig()
	if err != nil {
		return fmt.Errorf("failed to load user config: %w", err)
	}

	fmt.Println("Current configuration:")

	// Show environment with name if possible
	envDisplay := getValueOrDefault(userConfig.Settings.DefaultEnvironment, "(not set)")
	if userConfig.Settings.DefaultEnvironment != "" {
		// Try to resolve UUID to name for display
		apiConfig := config.Config{
			ApiURL:      "https://api.staging.devgraph.ai",
			IssuerURL:   config.DefaultIssuerURL,
			ClientID:    config.DefaultClientID,
			RedirectURL: config.DefaultRedirectURL,
		}
		if envs, err := util.GetEnvironments(apiConfig); err == nil && envs != nil {
			for _, env := range *envs {
				if env.Id.String() == userConfig.Settings.DefaultEnvironment {
					envDisplay = fmt.Sprintf("%s (%s)", env.Name, env.Id.String())
					break
				}
			}
		}
	}
	fmt.Printf("  Environment: %s\n", envDisplay)
	fmt.Printf("  Model: %s\n", getValueOrDefault(userConfig.Settings.DefaultModel, "gpt-4o-mini (default)"))

	var maxTokensStr string
	if userConfig.Settings.DefaultMaxTokens > 0 {
		maxTokensStr = fmt.Sprintf("%d", userConfig.Settings.DefaultMaxTokens)
	} else {
		maxTokensStr = "1000 (default)"
	}
	fmt.Printf("  Max Tokens: %s\n", maxTokensStr)

	configPath, _ := config.GetUserConfigPath()
	fmt.Printf("\nConfig file: %s\n", configPath)

	return nil
}

func getValueOrDefault(value, defaultValue string) string {
	if value != "" {
		return value
	}
	return defaultValue
}
